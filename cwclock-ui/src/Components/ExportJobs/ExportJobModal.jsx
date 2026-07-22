import React, { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { useI18n } from "../../i18n/I18nContext";
import Modal from "../common/Modal";
import MultiSelect from "../common/MultiSelect";
import Button from "../common/Button";
import Tooltip from "../common/Tooltip";
import styles from "./Styles/ExportJobModal.module.css";

const REPORT_TYPES = [
  { value: "summary-pdf", label: "summary-pdf" },
  { value: "summary-csv", label: "summary-csv" },
  { value: "detailed-pdf", label: "detailed-pdf" },
  { value: "detailed-csv", label: "detailed-csv" },
];

const CRON_HELPERS = [
  { label: "Every day at 9am", value: "0 9 * * *" },
  { label: "Every Monday at 9am", value: "0 9 * * 1" },
  { label: "Every 1st of month at 9am", value: "0 9 1 * *" },
  { label: "Every day at 6pm", value: "0 18 * * *" },
];

const TIME_PERIOD_HELPERS = [
  { label: "now()", value: "now()" },
  { label: "Last 24 hours", value: "now()-1d" },
  { label: "Last 7 days", value: "now()-7d" },
  { label: "Last 30 days", value: "now()-30d" },
  { label: "Last hour", value: "now()-1h" },
];

const ExportJobModal = ({ show, job, onSave, onClose }) => {
  const { t } = useI18n();
  const { clients } = useSelector((state) => state.clients);
  const { projects } = useSelector((state) => state.projects);

  const [formData, setFormData] = useState({
    name: "",
    cronExpression: "",
    reportTypes: [],
    timePeriod: "",
    clientIds: [],
    projectIds: [],
    includeFinancial: false,
    enabled: true,
    targets: [{ type: "email", toEmails: [], ccEmails: [] }],
  });

  const [errors, setErrors] = useState({});

  useEffect(() => {
    if (job) {
      setFormData(job);
    } else {
      setFormData({
        name: "",
        cronExpression: "",
        reportTypes: [],
        timePeriod: "",
        clientIds: [],
        projectIds: [],
        includeFinancial: false,
        enabled: true,
        targets: [{ type: "email", toEmails: [], ccEmails: [] }],
      });
    }
    setErrors({});
  }, [job, show]);

  const validateForm = () => {
    const newErrors = {};
    if (!formData.name.trim()) {
      newErrors.name = t("common.nameRequired");
    }
    if (!formData.cronExpression.trim()) {
      newErrors.cronExpression = t("exportJobs.cronExpressionRequired");
    }
    if (formData.reportTypes.length === 0) {
      newErrors.reportTypes = t("exportJobs.selectReportTypes");
    }
    if (!formData.timePeriod.trim()) {
      newErrors.timePeriod = t("exportJobs.timePeriodRequired");
    }
    if (formData.targets.length === 0 || formData.targets[0].toEmails.length === 0) {
      newErrors.targets = t("exportJobs.selectTargets");
    }
    return newErrors;
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    const newErrors = validateForm();
    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }
    onSave(formData);
  };

  const handleTargetEmailChange = (index, field, value) => {
    const newTargets = [...formData.targets];
    if (field === "toEmails" || field === "ccEmails") {
      newTargets[index][field] = value.split(",").map((e) => e.trim());
    } else {
      newTargets[index][field] = value;
    }
    setFormData({ ...formData, targets: newTargets });
  };

  if (!show) return null;

  return (
    <Modal title={job ? t("exportJobs.editJob") : t("exportJobs.newJob")} onClose={onClose}>
      <form onSubmit={handleSubmit} className={styles.form}>
        <div className={styles.formGroup}>
          <label htmlFor="name">{t("common.name")} *</label>
          <input
            id="name"
            type="text"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            className={errors.name ? styles.inputError : ""}
            placeholder={t("exportJobs.jobNamePlaceholder")}
          />
          {errors.name && <span className={styles.error}>{errors.name}</span>}
        </div>

        <div className={styles.formGroup}>
          <label htmlFor="cronExpression">{t("exportJobs.cronExpression")} *</label>
          <div className={styles.inputWithHelper}>
            <input
              id="cronExpression"
              type="text"
              value={formData.cronExpression}
              onChange={(e) =>
                setFormData({ ...formData, cronExpression: e.target.value })
              }
              className={errors.cronExpression ? styles.inputError : ""}
              placeholder="0 9 * * *"
            />
            <div className={styles.helperButtons}>
              {CRON_HELPERS.map((h) => (
                <Tooltip key={h.value} label={h.label} position="top">
                  <button
                    type="button"
                    className={styles.helperBtn}
                    onClick={() =>
                      setFormData({ ...formData, cronExpression: h.value })
                    }
                  >
                    {h.label.split(" ").slice(0, 2).join(" ")}
                  </button>
                </Tooltip>
              ))}
            </div>
          </div>
          {errors.cronExpression && (
            <span className={styles.error}>{errors.cronExpression}</span>
          )}
        </div>

        <div className={styles.formGroup}>
          <label htmlFor="timePeriod">{t("exportJobs.timePeriod")} *</label>
          <div className={styles.inputWithHelper}>
            <input
              id="timePeriod"
              type="text"
              value={formData.timePeriod}
              onChange={(e) => setFormData({ ...formData, timePeriod: e.target.value })}
              className={errors.timePeriod ? styles.inputError : ""}
              placeholder="now()-1d"
            />
            <div className={styles.helperButtons}>
              {TIME_PERIOD_HELPERS.map((h) => (
                <button
                  key={h.value}
                  type="button"
                  className={styles.helperBtn}
                  onClick={() => setFormData({ ...formData, timePeriod: h.value })}
                  title={h.label}
                >
                  {h.label.split(" ")[0]}
                </button>
              ))}
            </div>
          </div>
          {errors.timePeriod && (
            <span className={styles.error}>{errors.timePeriod}</span>
          )}
        </div>

        <div className={styles.formGroup}>
          <label>{t("exportJobs.reportTypes")} *</label>
          <MultiSelect
            options={REPORT_TYPES}
            selected={formData.reportTypes}
            onChange={(selected) => setFormData({ ...formData, reportTypes: selected })}
            placeholder={t("exportJobs.selectReportTypes")}
          />
          {errors.reportTypes && (
            <span className={styles.error}>{errors.reportTypes}</span>
          )}
        </div>

        <div className={styles.formGroup}>
          <label>{t("exportJobs.clients")}</label>
          <MultiSelect
            options={clients.map((c) => ({ value: c.id, label: c.name }))}
            selected={formData.clientIds}
            onChange={(selected) => setFormData({ ...formData, clientIds: selected })}
            placeholder={t("exportJobs.allClientsSelected")}
          />
        </div>

        <div className={styles.formGroup}>
          <label>{t("exportJobs.projects")}</label>
          <MultiSelect
            options={projects.map((p) => ({ value: p.id, label: p.name }))}
            selected={formData.projectIds}
            onChange={(selected) => setFormData({ ...formData, projectIds: selected })}
            placeholder={t("exportJobs.allProjectsSelected")}
          />
        </div>

        <div className={styles.formGroup}>
          <label>
            <input
              type="checkbox"
              checked={formData.includeFinancial}
              onChange={(e) =>
                setFormData({ ...formData, includeFinancial: e.target.checked })
              }
            />
            {t("exportJobs.includeFinancialData")}
          </label>
        </div>

        <div className={styles.formGroup}>
          <label>{t("exportJobs.emailTargets")} *</label>
          <div className={styles.targetSection}>
            <div>
              <label htmlFor="toEmails">{t("exportJobs.toEmails")}</label>
              <input
                id="toEmails"
                type="text"
                value={(formData.targets[0]?.toEmails || []).join(", ")}
                onChange={(e) => handleTargetEmailChange(0, "toEmails", e.target.value)}
                placeholder="email1@example.com, email2@example.com"
              />
            </div>
            <div>
              <label htmlFor="ccEmails">{t("exportJobs.ccEmails")}</label>
              <input
                id="ccEmails"
                type="text"
                value={(formData.targets[0]?.ccEmails || []).join(", ")}
                onChange={(e) => handleTargetEmailChange(0, "ccEmails", e.target.value)}
                placeholder="cc@example.com"
              />
            </div>
          </div>
          {errors.targets && (
            <span className={styles.error}>{errors.targets}</span>
          )}
        </div>

        <div className={styles.formGroup}>
          <label>
            <input
              type="checkbox"
              checked={formData.enabled}
              onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
            />
            {t("exportJobs.enabled")}
          </label>
        </div>

        <div className={styles.actions}>
          <Button onClick={onClose} variant="secondary">
            {t("common.cancel")}
          </Button>
          <Button type="submit" variant="primary">
            {job ? t("common.update") : t("common.create")}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default ExportJobModal;
