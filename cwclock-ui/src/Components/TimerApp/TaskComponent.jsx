import React, { useEffect, useState } from "react";
import styles from "./Styles/TaskComp.module.css";
import { MdDeleteForever } from "react-icons/md";
import { FaRegEdit, FaRegCopy } from "react-icons/fa";
import { useDispatch, useSelector } from "react-redux";
import { toast } from "react-toastify";
import { deleteTasksApi, updateTasksApi } from "../../Redux/Tasks/Task.actions";
import ConfirmModal from "../common/ConfirmModal";
import memberLabel from "../common/memberLabel";
import projectLabel from "../common/projectLabel";
import ProjectBadge from "../common/ProjectBadge";
import AutocompleteSelect from "../common/AutocompleteSelect";
import Tooltip from "../common/Tooltip";
import toastOptions from "../../Redux/toastOptions";
import { isAdminOrOwner as computeIsAdminOrOwner, memberRole } from "../common/permissions";
import padTimeString from "../common/padTimeString";
import { useI18n } from "../../i18n/I18nContext";

const fieldsFromItem = (item) => ({
  text: item.text,
  day: item.day,
  start: padTimeString(item.start) || "",
  end: padTimeString(item.end) || "",
  allDay: item.allDay,
  projectId: item.projectId,
  userId: item.userId,
});

const TaskComponent = ({ item }) => {
  const { t } = useI18n();
  const [isEditing, setIsEditing] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [form, setForm] = useState(fieldsFromItem(item));
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId, members } = useSelector((state) => state.organizations);
  const { projects } = useSelector((state) => state.projects);
  const { clients } = useSelector((state) => state.clients);
  const dispatch = useDispatch();

  const project = projects.find((p) => p.id === item.projectId);
  const myRole = memberRole(user, members);
  const isAdminOrOwner = computeIsAdminOrOwner(user, members);
  const canEdit = isAdminOrOwner || (myRole && myRole !== "reader");
  const assignee = members.find((m) => m.userId === item.userId);

  useEffect(() => {
    if (!isEditing) {
      setForm(fieldsFromItem(item));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [item, isEditing]);

  const handleCopy = (text) => {
    navigator.clipboard.writeText(text).then(() => {
      toast.success(t("common.copied"), toastOptions);
    });
  };

  const handleDelete = () => {
    setShowDeleteConfirm(false);
    dispatch(deleteTasksApi(item.id, currentOrgId, user.token));
  };

  const handleSave = () => {
    const selectedProject = projects.find((p) => p.id === form.projectId);
    const update = {
      ...item,
      text: form.text,
      day: form.day,
      allDay: form.allDay,
      start: form.allDay ? null : form.start,
      end: form.allDay ? null : form.end,
      projectId: form.projectId,
      clientId: selectedProject ? selectedProject.clientId : item.clientId,
      userId: form.userId,
    };
    dispatch(updateTasksApi(update, currentOrgId, user.token));
    setIsEditing(false);
  };

  const timeLabel = item.allDay
    ? t("timeTracker.allDay")
    : `${padTimeString(item.start) || "?"} - ${padTimeString(item.end) || "?"}`;

  const maxLabelLen = parseInt(process.env.REACT_APP_TASK_LABEL_MAX_LENGTH) || 50;
  const isTextTruncated = maxLabelLen > 0 && item.text.length > maxLabelLen;
  const truncatedText = isTextTruncated ? item.text.slice(0, maxLabelLen) + "..." : item.text;

  return (
    <div className={styles.Task}>
      <div className={styles.Desc}>
        {isEditing ? (
          <>
            <div className={styles.Up}>
              <input
                className="cw-input"
                type="text"
                autoFocus
                title={t("timeTracker.taskDescription")}
                value={form.text}
                onChange={(e) => setForm({ ...form, text: e.target.value })}
              />
            </div>
            <div className={styles.Projects}>
              <AutocompleteSelect
                label={t("timeTracker.project")}
                placeholder={t("timeTracker.project")}
                options={projects.map((p) => ({ value: p.id, label: projectLabel(p, clients) }))}
                value={form.projectId}
                onChange={(projectId) => setForm((f) => ({ ...f, projectId }))}
              />
            </div>
            <div className={styles.Time}>
              <input
                className="cw-input"
                type="date"
                title={t("timeTracker.day")}
                value={form.day}
                onChange={(e) => setForm({ ...form, day: e.target.value })}
              />
              <label className={styles.allDayLabel} title={t("timeTracker.markAllDay")}>
                <input
                  type="checkbox"
                  checked={form.allDay}
                  onChange={(e) => setForm({ ...form, allDay: e.target.checked })}
                />{" "}
                {t("timeTracker.allDay")}
              </label>
              {!form.allDay && (
                <>
                  <input
                    className="cw-input"
                    type="time"
                    step="1"
                    title={t("timeTracker.startTime")}
                    value={form.start}
                    onChange={(e) => setForm({ ...form, start: e.target.value })}
                  />
                  <input
                    className="cw-input"
                    type="time"
                    step="1"
                    title={t("timeTracker.endTime")}
                    value={form.end}
                    onChange={(e) => setForm({ ...form, end: e.target.value })}
                  />
                </>
              )}
              {isAdminOrOwner && (
                <AutocompleteSelect
                  label={t("nav.users")}
                  placeholder={t("timeTracker.searchMember")}
                  options={members.map((m) => ({ value: m.userId, label: memberLabel(m) }))}
                  value={form.userId}
                  onChange={(userId) => setForm((f) => ({ ...f, userId }))}
                />
              )}
            </div>
            <button type="button" className={styles.Tags3} onClick={handleSave} title={t("common.saveChanges")}>
              {t("common.save")}
            </button>
            <button
              type="button"
              className={styles.Tags}
              onClick={() => setIsEditing(false)}
              title={t("common.discardChanges")}
            >
              {t("common.cancel")}
            </button>
          </>
        ) : (
          <>
            <div className={styles.Up}>
              <div className={styles.CopyLabel}>
                <Tooltip label={isTextTruncated ? item.text : null}>
                  <h6>{truncatedText}</h6>
                </Tooltip>
                <Tooltip label={t("common.copy")}>
                  <button
                    type="button"
                    className={styles.CopyBtn}
                    onClick={() => handleCopy(item.text)}
                  >
                    <FaRegCopy style={{ fontSize: "14px" }} />
                  </button>
                </Tooltip>
              </div>
            </div>
            <div className={styles.Projects}>
              <ProjectBadge project={project} clients={clients} />
            </div>
            <div className={styles.Time}>
              <span>{item.day}</span>
              <span>{timeLabel}</span>
              {assignee && <span>{memberLabel(assignee)}</span>}
            </div>
          </>
        )}
        {!isEditing && canEdit && (
          <div className={styles.RowActions}>
            <Tooltip label={t("common.edit")}>
              <button type="button" className={styles.Tags} onClick={() => setIsEditing(true)}>
                <FaRegEdit style={{ fontSize: "18px" }} />
              </button>
            </Tooltip>
            <Tooltip label={t("common.delete")}>
              <button
                type="button"
                className={styles.Tags2}
                onClick={() => setShowDeleteConfirm(true)}
              >
                <MdDeleteForever style={{ fontSize: "20px" }} />
              </button>
            </Tooltip>
          </div>
        )}
      </div>

      <ConfirmModal
        show={showDeleteConfirm}
        title={t("timeTracker.deleteRecordTitle")}
        body={t("timeTracker.deleteRecordBody", { text: item.text })}
        confirmLabel={t("common.delete")}
        onConfirm={handleDelete}
        onCancel={() => setShowDeleteConfirm(false)}
      />
    </div>
  );
};

export default TaskComponent;
