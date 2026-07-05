import React, { useEffect } from "react";
import Tooltip from "./Tooltip";
import styles from "./Styles/Modal.module.css";

const Modal = ({ show, title, onClose, children, footer }) => {
  useEffect(() => {
    if (!show) return;
    const onKeyDown = (e) => e.key === "Escape" && onClose?.();
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [show, onClose]);

  if (!show) return null;

  return (
    <div className={styles.backdrop} onMouseDown={onClose}>
      <div
        className={styles.dialog}
        role="dialog"
        aria-modal="true"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className={styles.header}>
          <h3 className={styles.title}>{title}</h3>
          <Tooltip label="Close">
            <button type="button" className={styles.close} onClick={onClose} aria-label="Close">
              &times;
            </button>
          </Tooltip>
        </div>
        <div className={styles.body}>{children}</div>
        {footer && <div className={styles.footer}>{footer}</div>}
      </div>
    </div>
  );
};

export default Modal;
