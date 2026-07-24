import React, { useEffect, useState } from "react";
import Modal from "../common/Modal";
import Button from "../common/Button";
import Switch from "../common/Switch";
import AutocompleteSelect from "../common/AutocompleteSelect";
import ConfirmModal from "../common/ConfirmModal";
import projectLabel from "../common/projectLabel";
import padTimeString from "../common/padTimeString";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/CalendarEventModal.module.css";

const timeOnly = (hhmmss) => (hhmmss || "").slice(0, 5);

const fieldsFromEntry = (entry, defaultDay) => {
  if (entry) {
    return {
      text: entry.text,
      projectId: entry.projectId,
      day: entry.day,
      startTime: padTimeString(entry.start) || "09:00:00",
      endTime: padTimeString(entry.end) || "10:00:00",
      allDay: entry.allDay,
    };
  }
  return { text: "", projectId: "", day: defaultDay, startTime: "09:00:00", endTime: "10:00:00", allDay: false };
};

// CalendarEventModal is the Calendar view's "add/edit a meeting" popup - two
// date pickers (Start, End), a project autocomplete (client is derived from
// the chosen project, same convention as the classic time tracker screen),
// a label, and an all-day switch (ai-instruct-84).
const CalendarEventModal = ({ show, entry, defaultDay, projects, clients, onClose, onSave, onDelete }) => {
  const { t } = useI18n();
  const [form, setForm] = useState(() => fieldsFromEntry(entry, defaultDay));
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  useEffect(() => {
    if (show) {
      setForm(fieldsFromEntry(entry, defaultDay));
      setError("");
      setShowDeleteConfirm(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [show, entry, defaultDay]);

  const handleStartChange = (value) => {
    if (!value) return;
    if (form.allDay) {
      setForm((f) => ({ ...f, day: value }));
      return;
    }
    const [day, time] = value.split("T");
    setForm((f) => ({ ...f, day, startTime: `${time}:00` }));
  };

  // The End picker only ever edits the time of day: this app's time entries
  // have a single "day" field (no separate end day), so a multi-day span
  // isn't representable yet - the date portion always snaps back to Start's
  // day, however the native picker was used to set it.
  const handleEndChange = (value) => {
    if (!value) return;
    const [, time] = value.split("T");
    if (time) setForm((f) => ({ ...f, endTime: `${time}:00` }));
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!form.text.trim() || !form.projectId || !form.day) {
      setError(t("calendar.fillRequiredFields"));
      return;
    }
    if (!form.allDay && (!form.startTime || !form.endTime)) {
      setError(t("calendar.fillRequiredFields"));
      return;
    }
    const project = projects.find((p) => p.id === form.projectId);
    const payload = {
      ...(entry || {}),
      text: form.text.trim(),
      day: form.day,
      allDay: form.allDay,
      start: form.allDay ? null : form.startTime,
      end: form.allDay ? null : form.endTime,
      projectId: form.projectId,
      clientId: project ? project.clientId : entry?.clientId,
    };
    setError("");
    setSubmitting(true);
    Promise.resolve(onSave(payload))
      .catch(() => {})
      .finally(() => setSubmitting(false));
  };

  return (
    <Modal show={show} title={entry ? t("calendar.editTimeRecord") : t("calendar.addTimeRecord")} onClose={onClose}>
      <form className={styles.form} onSubmit={handleSubmit}>
        <div className="cw-field">
          <label className="cw-label">{t("timeTracker.taskDescription")}</label>
          <input
            className="cw-input"
            type="text"
            autoFocus
            value={form.text}
            onChange={(e) => setForm({ ...form, text: e.target.value })}
          />
        </div>

        <div className="cw-field">
          <AutocompleteSelect
            label={t("timeTracker.project")}
            placeholder={t("timeTracker.project")}
            options={projects.map((p) => ({ value: p.id, label: projectLabel(p, clients) }))}
            value={form.projectId}
            onChange={(projectId) => setForm((f) => ({ ...f, projectId }))}
          />
        </div>

        <div className={styles.switchField}>
          <Switch
            checked={form.allDay}
            onChange={(e) => setForm({ ...form, allDay: e.target.checked })}
            aria-label={t("timeTracker.allDay")}
            title={t("timeTracker.markAllDay")}
          />
          <span className="cw-label">{t("timeTracker.allDay")}</span>
        </div>

        <div className="cw-field">
          <label className="cw-label">{t("calendar.start")}</label>
          <input
            className="cw-input"
            type={form.allDay ? "date" : "datetime-local"}
            value={form.allDay ? form.day : `${form.day}T${timeOnly(form.startTime)}`}
            onChange={(e) => handleStartChange(e.target.value)}
          />
        </div>

        <div className="cw-field">
          <label className="cw-label">{t("calendar.end")}</label>
          <input
            className="cw-input"
            type={form.allDay ? "date" : "datetime-local"}
            value={form.allDay ? form.day : `${form.day}T${timeOnly(form.endTime)}`}
            disabled={form.allDay}
            onChange={(e) => handleEndChange(e.target.value)}
          />
          <span className={styles.hint}>{t("calendar.endSameDayHint")}</span>
        </div>

        {error && <p className="cw-error">{error}</p>}

        <div className={styles.actions}>
          {entry && (
            <Button
              type="button"
              variant="danger"
              onClick={() => setShowDeleteConfirm(true)}
              title={t("common.delete")}
            >
              {t("common.delete")}
            </Button>
          )}
          <Button type="button" variant="secondary" onClick={onClose} title={t("common.cancel")}>
            {t("common.cancel")}
          </Button>
          <Button type="submit" disabled={submitting} title={t("common.save")}>
            {t("common.save")}
          </Button>
        </div>
      </form>

      <ConfirmModal
        show={showDeleteConfirm}
        title={t("timeTracker.deleteRecordTitle")}
        body={t("timeTracker.deleteRecordBody", { text: entry?.text })}
        confirmLabel={t("common.delete")}
        onConfirm={() => onDelete(entry.id)}
        onCancel={() => setShowDeleteConfirm(false)}
      />
    </Modal>
  );
};

export default CalendarEventModal;
