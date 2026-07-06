import React from "react";
import Modal from "./Modal";
import Button from "./Button";
import { useI18n } from "../../i18n/I18nContext";

const ConfirmModal = ({ show, title, body, confirmLabel, onConfirm, onCancel }) => {
  const { t } = useI18n();
  const resolvedConfirmLabel = confirmLabel ?? t("common.confirm");

  return (
    <Modal
      show={show}
      title={title}
      onClose={onCancel}
      footer={
        <>
          <Button variant="secondary" onClick={onCancel} title={t("common.cancel")}>
            {t("common.cancel")}
          </Button>
          <Button variant="danger" onClick={onConfirm} title={resolvedConfirmLabel}>
            {resolvedConfirmLabel}
          </Button>
        </>
      }
    >
      {body}
    </Modal>
  );
};

export default ConfirmModal;
