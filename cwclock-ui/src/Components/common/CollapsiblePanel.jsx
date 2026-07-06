import React, { useState } from "react";
import { FiChevronDown, FiChevronUp } from "react-icons/fi";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/CollapsiblePanel.module.css";

const CollapsiblePanel = ({ title, defaultOpen = false, children }) => {
  const { t } = useI18n();
  const [open, setOpen] = useState(defaultOpen);

  return (
    <div className={styles.panel}>
      <button
        type="button"
        className={styles.header}
        onClick={() => setOpen(!open)}
        aria-expanded={open}
        title={`${open ? t("common.collapse") : t("common.expand")} ${title}`}
      >
        <span className={styles.title}>{title}</span>
        {open ? <FiChevronUp /> : <FiChevronDown />}
      </button>
      {open && <div className={styles.body}>{children}</div>}
    </div>
  );
};

export default CollapsiblePanel;
