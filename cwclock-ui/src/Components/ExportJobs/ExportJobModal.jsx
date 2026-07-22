import React, { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { useI18n } from "../../i18n/I18nContext";
import Modal from "../common/Modal";
import MultiSelect from "../common/MultiSelect";
import InputWithHelper from "../common/InputWithHelper";
import ExportTargetsEditor from "./ExportTargetsEditor";
import Button from "../common/Button";
import RequiredMark from "../common/RequiredMark";
import styles from "./Styles/ExportJobModal.module.css";

const REPORT_TYPES = [
  { value: "summary-pdf", label: "summary-pdf" },
  { value: "summary-csv", label: "summary-csv" },
  { value: "detailed-pdf", label: "detailed-pdf" },
  { value: "detailed-csv", label: "detailed-csv" },
];

const CRON_HELPERS = [
  { label: "Every minute", value: "* * * * *" },
  { label: "Every hour", value: "0 * * * *" },
  { label: "Every day", value: "0 0 * * *" },
  { label: "Every day at 9am", value: "0 9 * * *" },
  { label: "Every day at 6pm", value: "0 18 * * *" },
  { label: "Every Monday at 9am", value: "0 9 * * 1" },
  { label: "Every 1st of month at 9am", value: "0 9 1 * *" },
];

const TIME_PERIOD_HELPERS = [
  { label: "now()", value: "now()" },
  { label: "Last 24 hours", value: "now()-1d" },
  { label: "Last 7 days", value: "now()-7d" },
  { label: "Last 30 days", value: "now()-30d" },
  { label: "Last hour", value: "now()-1h" },
];

const emptyFormData = {
  name: "",
  cronExpression: "",
  reportTypes: [],
  timePeriod: "",
  clientIds: [],
  projectIds: [],
  includeFinancial: false,
  enabled: true,
  targets: [],
};

// Targets coming back from the API only carry the fields relevant to their
// own type (see models.ExportTarget's omitempty tags) - default the rest so
// ExportTargetsEditor's inputs are always controlled.
const normalizeTargets = (targets) =>
  (targets || []).map((target) => ({
    type: target.type,
    toEmails: target.toEmails || "",
    ccEmails: target.ccEmails || "",
    connection: target.connection || "",
  }));

const ExportJobModal = ({ show, job, onSave, onClose }) => {
  const { t } = useI18n();
  const { clients } = useSelector((state) => state.clients);
  const { projects } = useSelector((state) => state.projects);
  const { organizations, currentOrgId } = useSelector((state) => state.organizations);
  const orgConnections = organizations.find((o) => o.id === currentOrgId)?.externalConnections || [];

  const [formData, setFormData] = useState(emptyFormData);
  const [errors, setErrors] = useState({});

  useEffect(() => {
    if (job) {
      setFormData({ ...job, targets: normalizeTargets(job.targets) });
    } else {
      setFormData(emptyFormData);
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
    if (formData.targets.length === 0) {
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

  if (!show) return null;

  return (
    <Modal show={show} title={job ? t("exportJobs.editJob") : t("exportJobs.newJob")} onClose={onClose}>
      <form onSubmit={handleSubmit} className={styles.form}>
        <div className="cw-field">
          <label htmlFor="name" className="cw-label">
            {t("common.name")}
            <RequiredMark />
          </label>
          <input
            id="name"
            type="text"
            className="cw-input"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            placeholder={t("exportJobs.jobNamePlaceholder")}
          />
          {errors.name && <p className="cw-error">{errors.name}</p>}
        </div>

        <div className="cw-field">
          <label htmlFor="cronExpression" className="cw-label">
            {t("exportJobs.cronExpression")}
            <RequiredMark />
          </label>
          <InputWithHelper
            id="cronExpression"
            value={formData.cronExpression}
            onChange={(value) => setFormData({ ...formData, cronExpression: value })}
            options={CRON_HELPERS}
            placeholder="0 9 * * *"
            error={!!errors.cronExpression}
          />
          {errors.cronExpression && <p className="cw-error">{errors.cronExpression}</p>}
        </div>

        <div className="cw-field">
          <label htmlFor="timePeriod" className="cw-label">
            {t("exportJobs.timePeriod")}
            <RequiredMark />
          </label>
          <InputWithHelper
            id="timePeriod"
            value={formData.timePeriod}
            onChange={(value) => setFormData({ ...formData, timePeriod: value })}
            options={TIME_PERIOD_HELPERS}
            placeholder="now()-1d"
            error={!!errors.timePeriod}
          />
          {errors.timePeriod && <p className="cw-error">{errors.timePeriod}</p>}
        </div>

        <div className="cw-field">
          <label className="cw-label">
            {t("exportJobs.reportTypes")}
            <RequiredMark />
          </label>
          <MultiSelect
            options={REPORT_TYPES}
            selected={formData.reportTypes}
            onChange={(selected) => setFormData({ ...formData, reportTypes: selected })}
            placeholder={t("exportJobs.selectReportTypes")}
          />
          {errors.reportTypes && <p className="cw-error">{errors.reportTypes}</p>}
        </div>

        <div className="cw-field">
          <label className="cw-label">{t("exportJobs.clients")}</label>
          <MultiSelect
            options={clients.map((c) => ({ value: c.id, label: c.name }))}
            selected={formData.clientIds}
            onChange={(selected) => setFormData({ ...formData, clientIds: selected })}
            placeholder={t("exportJobs.allClientsSelected")}
          />
        </div>

        <div className="cw-field">
          <label className="cw-label">{t("exportJobs.projects")}</label>
          <MultiSelect
            options={projects.map((p) => ({ value: p.id, label: p.name }))}
            selected={formData.projectIds}
            onChange={(selected) => setFormData({ ...formData, projectIds: selected })}
            placeholder={t("exportJobs.allProjectsSelected")}
          />
        </div>

        <div className="cw-field">
          <label className="cw-checkbox">
            <input
              type="checkbox"
              checked={formData.includeFinancial}
              onChange={(e) => setFormData({ ...formData, includeFinancial: e.target.checked })}
            />
            {t("exportJobs.includeFinancialData")}
          </label>
        </div>

        <div className="cw-field">
          <label className="cw-label">
            {t("exportJobs.targetsLabel")}
            <RequiredMark />
          </label>
          <ExportTargetsEditor
            targets={formData.targets}
            onChange={(targets) => setFormData({ ...formData, targets })}
            connections={orgConnections}
            error={errors.targets}
          />
        </div>

        <div className="cw-field">
          <label className="cw-checkbox">
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
