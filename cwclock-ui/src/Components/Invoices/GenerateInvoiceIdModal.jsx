import React, { useEffect, useState } from "react";
import Modal from "../common/Modal";
import Button from "../common/Button";
import RequiredMark from "../common/RequiredMark";
import { useI18n } from "../../i18n/I18nContext";

// GenerateInvoiceIdModal prompts for an invoice id/number to use instead of
// the usual computed one, for the "Generate with id" action.
const GenerateInvoiceIdModal = ({ show, onClose, onConfirm }) => {
  const { t } = useI18n();
  const [number, setNumber] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (show) {
      setNumber("");
      setError("");
    }
  }, [show]);

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!number.trim()) {
      setError(t("invoices.invoiceIdRequired"));
      return;
    }
    onConfirm(number.trim());
  };

  return (
    <Modal show={show} title={t("invoices.generateWithIdTitle")} onClose={onClose}>
      <form onSubmit={handleSubmit}>
        <div className="cw-field">
          <label className="cw-label">
            {t("invoices.invoiceIdLabel")}
            <RequiredMark />
          </label>
          <input
            className="cw-input"
            type="text"
            value={number}
            onChange={(e) => setNumber(e.target.value)}
            placeholder={t("invoices.invoiceIdPlaceholder")}
            title={t("invoices.invoiceIdLabel")}
            autoFocus
          />
        </div>
        {error && <p className="cw-error">{error}</p>}
        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 16 }}>
          <Button type="button" variant="secondary" onClick={onClose} title={t("common.cancel")}>
            {t("common.cancel")}
          </Button>
          <Button type="submit" title={t("invoices.generateWithId")}>
            {t("invoices.generate")}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default GenerateInvoiceIdModal;
