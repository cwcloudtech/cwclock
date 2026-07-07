import React, { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { toast } from "react-toastify";
import Modal from "../common/Modal";
import Button from "../common/Button";
import { importCSVApi } from "../../Redux/Import/Import.actions";
import { getTasksApi } from "../../Redux/Tasks/Task.actions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import toastOptions from "../../Redux/toastOptions";
import styles from "./Styles/ImportCSVModal.module.css";

// Detailed-report CSV import, restricted to org admins/owners (see TasksApp).
// Accepts the same column format this app's own detailed report exports
// (see report.DetailedCSV on the backend) as well as Clockify's, so it also
// works as the way to migrate time entries from another cwclock instance.
const ImportCSVModal = ({ show, onClose }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId } = useSelector((state) => state.organizations);
  const [file, setFile] = useState(null);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleClose = () => {
    setFile(null);
    setError("");
    setSubmitting(false);
    onClose();
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!file || submitting) return;
    setError("");
    setSubmitting(true);
    try {
      const text = await file.text();
      const { created, skipped } = await dispatch(importCSVApi(currentOrgId, text, user.token));
      toast.success(t("timeTracker.importResult", { created, skipped }), toastOptions);
      dispatch(getTasksApi(currentOrgId, user.token));
      handleClose();
    } catch (err) {
      setError(apiErrorMessage(err, locale));
      setSubmitting(false);
    }
  };

  return (
    <Modal show={show} title={t("timeTracker.importCsvTitle")} onClose={handleClose}>
      <form className={styles.form} onSubmit={handleSubmit}>
        <p className={styles.hint}>{t("timeTracker.importCsvHint")}</p>
        <div className="cw-field">
          <label className="cw-label">{t("timeTracker.importCsvFile")}</label>
          <input
            className="cw-input"
            type="file"
            accept=".csv,text/csv"
            onChange={(e) => setFile(e.target.files[0] || null)}
          />
        </div>
        {error && <p className="cw-error">{error}</p>}
        <div className={styles.actions}>
          <Button variant="secondary" onClick={handleClose} title={t("common.cancel")}>
            {t("common.cancel")}
          </Button>
          <Button type="submit" disabled={!file || submitting} title={t("timeTracker.importCsvTitle")}>
            {t("common.import")}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default ImportCSVModal;
