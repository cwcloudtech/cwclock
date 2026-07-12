import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaRegEdit } from "react-icons/fa";
import { MdDeleteForever } from "react-icons/md";
import { updateTasksApi, deleteTasksApi } from "../../Redux/Tasks/Task.actions";
import ConfirmModal from "../common/ConfirmModal";
import memberLabel from "../common/memberLabel";
import projectLabel from "../common/projectLabel";
import AutocompleteSelect from "../common/AutocompleteSelect";
import Tooltip from "../common/Tooltip";
import padTimeString from "../common/padTimeString";
import { useI18n } from "../../i18n/I18nContext";
import { formatHMS } from "./reportFormat";
import styles from "./Styles/Reports.module.css";

const fieldsFromEntry = (e) => ({
  text: e.text,
  day: e.day,
  start: padTimeString(e.start) || "",
  end: padTimeString(e.end) || "",
  allDay: e.allDay,
  clientId: e.clientId,
  projectId: e.projectId,
  userId: e.userId,
});

// One detailed-report row: read-only by default, switching to an inline
// edit form (reusing the same update/delete actions as the time tracker) so
// descriptions can be fixed or entries reassigned without leaving the report.
const ReportEntryRow = ({ entry, orgId, currency, isAdminOrOwner, showAmount, onChanged }) => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { members } = useSelector((state) => state.organizations);
  const { clients } = useSelector((state) => state.clients);
  const { projects } = useSelector((state) => state.projects);
  const [isEditing, setIsEditing] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [form, setForm] = useState(fieldsFromEntry(entry));

  useEffect(() => {
    if (!isEditing) {
      setForm(fieldsFromEntry(entry));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [entry, isEditing]);

  const canEditRow = isAdminOrOwner || entry.userId === user.id;

  // A single project/client picker (like the time record webview's) since
  // picking a project already determines its client.
  const handleProjectChange = (projectId) => {
    const project = projects.find((p) => p.id === projectId);
    setForm((f) => ({ ...f, projectId, clientId: project ? project.clientId : f.clientId }));
  };

  const handleSave = () => {
    dispatch(
      updateTasksApi(
        {
          id: entry.id,
          clientId: form.clientId,
          projectId: form.projectId,
          userId: form.userId,
          text: form.text,
          day: form.day,
          allDay: form.allDay,
          start: form.allDay ? null : form.start,
          end: form.allDay ? null : form.end,
        },
        orgId,
        user.token
      )
    );
    setIsEditing(false);
    onChanged();
  };

  const handleDelete = () => {
    setShowDeleteConfirm(false);
    dispatch(deleteTasksApi(entry.id, orgId, user.token));
    onChanged();
  };

  if (isEditing) {
    return (
      <div className={styles.editRow}>
        <input
          className="cw-input"
          type="text"
          value={form.text}
          onChange={(e) => setForm({ ...form, text: e.target.value })}
          title={t("timeTracker.taskDescription")}
          autoFocus
        />
        <input
          className="cw-input"
          type="date"
          value={form.day}
          onChange={(e) => setForm({ ...form, day: e.target.value })}
          title={t("timeTracker.day")}
        />
        <label className={styles.allDayLabel}>
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
              value={form.start}
              onChange={(e) => setForm({ ...form, start: e.target.value })}
              title={t("timeTracker.startTime")}
            />
            <input
              className="cw-input"
              type="time"
              step="1"
              value={form.end}
              onChange={(e) => setForm({ ...form, end: e.target.value })}
              title={t("timeTracker.endTime")}
            />
          </>
        )}
        {isAdminOrOwner && (
          <>
            <AutocompleteSelect
              label={t("projects.title")}
              placeholder={t("timeTracker.project")}
              options={projects.map((p) => ({ value: p.id, label: projectLabel(p, clients) }))}
              value={form.projectId}
              onChange={handleProjectChange}
            />
            <AutocompleteSelect
              label={t("nav.users")}
              placeholder={t("timeTracker.searchMember")}
              options={members.map((m) => ({ value: m.userId, label: memberLabel(m) }))}
              value={form.userId}
              onChange={(userId) => setForm((f) => ({ ...f, userId }))}
            />
          </>
        )}
        <div className={styles.editRowActions}>
          <button type="button" className="cw-button cw-button--sm" onClick={handleSave} title={t("common.saveChanges")}>
            {t("common.save")}
          </button>
          <button
            type="button"
            className="cw-button cw-button--secondary cw-button--sm"
            onClick={() => setIsEditing(false)}
            title={t("common.discardChanges")}
          >
            {t("common.cancel")}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.tableRow}>
      <span>{entry.day}</span>
      <span>
        {entry.text}
        <div className={styles.subText}>
          {entry.projectName} - {entry.clientName}
        </div>
      </span>
      <span>
        {padTimeString(entry.start) || "?"} - {padTimeString(entry.end) || "?"}
        {entry.allDay && ` (${t("timeTracker.allDay")})`}
      </span>
      <span>{formatHMS(entry.durationSecs)}</span>
      <span>{entry.userName}</span>
      <span>{entry.userEmail}</span>
      {showAmount && (
        <span>
          {(entry.amount ?? 0).toFixed(2)} {currency}
        </span>
      )}
      <span className={styles.rowActions}>
        {canEditRow && (
          <>
            <Tooltip label={t("common.edit")}>
              <button type="button" className={styles.iconBtn} onClick={() => setIsEditing(true)}>
                <FaRegEdit />
              </button>
            </Tooltip>
            <Tooltip label={t("common.delete")}>
              <button type="button" className={styles.iconBtn} onClick={() => setShowDeleteConfirm(true)}>
                <MdDeleteForever />
              </button>
            </Tooltip>
          </>
        )}
      </span>

      <ConfirmModal
        show={showDeleteConfirm}
        title={t("timeTracker.deleteRecordTitle")}
        body={t("timeTracker.deleteRecordBody", { text: entry.text })}
        confirmLabel={t("common.delete")}
        onConfirm={handleDelete}
        onCancel={() => setShowDeleteConfirm(false)}
      />
    </div>
  );
};

export default ReportEntryRow;
