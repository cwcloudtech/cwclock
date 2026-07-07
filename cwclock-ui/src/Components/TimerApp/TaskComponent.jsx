import React, { useEffect, useState } from "react";
import styles from "./Styles/TaskComp.module.css";
import { MdDeleteForever } from "react-icons/md";
import { FaRegEdit } from "react-icons/fa";
import { useDispatch, useSelector } from "react-redux";
import { deleteTasksApi, updateTasksApi } from "../../Redux/Tasks/Task.actions";
import ConfirmModal from "../common/ConfirmModal";
import memberLabel from "../common/memberLabel";
import projectLabel from "../common/projectLabel";
import Tooltip from "../common/Tooltip";
import { useI18n } from "../../i18n/I18nContext";

const fieldsFromItem = (item) => ({
  text: item.text,
  day: item.day,
  start: item.start || "",
  end: item.end || "",
  allDay: item.allDay,
  projectId: item.projectId,
  userId: item.userId,
});

const TaskComponent = ({ item }) => {
  const { t } = useI18n();
  const [isEditing, setIsEditing] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [form, setForm] = useState(fieldsFromItem(item));
  const [reassignText, setReassignText] = useState("");
  const [projectQuery, setProjectQuery] = useState("");
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId, members } = useSelector((state) => state.organizations);
  const { projects } = useSelector((state) => state.projects);
  const { clients } = useSelector((state) => state.clients);
  const dispatch = useDispatch();

  const project = projects.find((p) => p.id === item.projectId);
  const myRole = members.find((m) => m.userId === user.id)?.role;
  const isAdminOrOwner = myRole === "admin" || myRole === "owner";
  const canEdit = myRole && myRole !== "reader";
  const assignee = members.find((m) => m.userId === item.userId);

  useEffect(() => {
    if (!isEditing) {
      setForm(fieldsFromItem(item));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [item, isEditing]);

  useEffect(() => {
    if (isEditing) {
      const current = members.find((m) => m.userId === form.userId);
      setReassignText(current ? memberLabel(current) : "");
      setProjectQuery(project ? projectLabel(project, clients) : "");
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isEditing]);

  const handleReassignInput = (text) => {
    setReassignText(text);
    const match = members.find((m) => memberLabel(m) === text || m.email === text);
    if (match) {
      setForm((f) => ({ ...f, userId: match.userId }));
    }
  };

  const handleProjectInput = (text) => {
    setProjectQuery(text);
    const match = projects.find((p) => projectLabel(p, clients) === text);
    if (match) {
      setForm((f) => ({ ...f, projectId: match.id }));
    }
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

  const timeLabel = item.allDay ? t("timeTracker.allDay") : `${item.start || "?"} - ${item.end || "?"}`;

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
              <input
                className="cw-input"
                list={`project-options-${item.id}`}
                title={t("timeTracker.searchByCustomerOrProject")}
                placeholder={t("timeTracker.project")}
                value={projectQuery}
                onChange={(e) => handleProjectInput(e.target.value)}
              />
              <datalist id={`project-options-${item.id}`}>
                {projects.map((p) => (
                  <option key={p.id} value={projectLabel(p, clients)} />
                ))}
              </datalist>
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
                <>
                  <input
                    className="cw-input"
                    list={`reassign-options-${item.id}`}
                    title={t("timeTracker.reassignToMember")}
                    placeholder={t("timeTracker.searchMember")}
                    value={reassignText}
                    onChange={(e) => handleReassignInput(e.target.value)}
                  />
                  <datalist id={`reassign-options-${item.id}`}>
                    {members.map((m) => (
                      <option key={m.userId} value={memberLabel(m)} />
                    ))}
                  </datalist>
                </>
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
              <h6>{item.text}</h6>
            </div>
            <div className={styles.Projects}>
              <h6>{project ? projectLabel(project, clients) : ""}</h6>
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
