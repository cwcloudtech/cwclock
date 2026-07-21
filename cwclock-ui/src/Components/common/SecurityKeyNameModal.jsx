import React, { useEffect, useState } from "react";
import Modal from "./Modal";
import Button from "./Button";
import { useI18n } from "../../i18n/I18nContext";

// SecurityKeyNameModal asks for a security key's device name once the
// browser's native WebAuthn ceremony has completed (see MfaSettings) - a
// real modal replacing the previous window.prompt (see ai-instruct-70).
const SecurityKeyNameModal = ({ show, busy, onConfirm, onCancel }) => {
  const { t } = useI18n();
  const [name, setName] = useState("");

  useEffect(() => {
    if (show) {
      setName("");
    }
  }, [show]);

  const handleSubmit = (e) => {
    e.preventDefault();
    onConfirm(name.trim() || t("profile.mfaSecurityKeyDefaultName"));
  };

  return (
    <Modal show={show} title={t("profile.mfaSecurityKeyNamePrompt")} onClose={onCancel}>
      <form onSubmit={handleSubmit}>
        <div className="cw-field">
          <label className="cw-label">{t("profile.mfaSecurityKeyNameLabel")}</label>
          <input
            className="cw-input"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t("profile.mfaSecurityKeyDefaultName")}
            title={t("profile.mfaSecurityKeyNameLabel")}
            autoFocus
          />
        </div>
        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 16 }}>
          <Button type="button" variant="secondary" onClick={onCancel} title={t("common.cancel")}>
            {t("common.cancel")}
          </Button>
          <Button type="submit" disabled={busy} title={t("common.save")}>
            {t("common.save")}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default SecurityKeyNameModal;
