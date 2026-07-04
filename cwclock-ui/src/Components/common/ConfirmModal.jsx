import React from "react";
import Modal from "./Modal";
import Button from "./Button";

const ConfirmModal = ({ show, title, body, confirmLabel = "Confirm", onConfirm, onCancel }) => (
  <Modal
    show={show}
    title={title}
    onClose={onCancel}
    footer={
      <>
        <Button variant="secondary" onClick={onCancel} title="Cancel">
          Cancel
        </Button>
        <Button variant="danger" onClick={onConfirm} title={confirmLabel}>
          {confirmLabel}
        </Button>
      </>
    }
  >
    {body}
  </Modal>
);

export default ConfirmModal;
