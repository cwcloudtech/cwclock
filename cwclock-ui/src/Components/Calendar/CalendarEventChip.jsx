import React from "react";
import { FaRegCopy } from "react-icons/fa";
import { toast } from "react-toastify";
import contrastColor from "../common/contrastColor";
import Tooltip from "../common/Tooltip";
import toastOptions from "../../Redux/toastOptions";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/CalendarEventChip.module.css";

// CalendarEventChip renders one time entry as a project-colored block, shared
// by the month grid, the week view's all-day row and its hourly grid
// (ai-instruct-84/85/86): the label truncates with an ellipsis when it
// doesn't fit, the full text is always available via a tooltip and a copy
// button, and (when draggable) the chip can be dragged onto another day to
// reschedule the entry.
const CalendarEventChip = ({ label, color, draggable, onDragStart, onClick, className, style }) => {
  const { t } = useI18n();

  const handleCopy = (e) => {
    e.stopPropagation();
    navigator.clipboard.writeText(label).then(() => {
      toast.success(t("common.copied"), toastOptions);
    });
  };

  return (
    <Tooltip label={label} className={styles.chipTooltip}>
      <div
        className={`${styles.eventChip} ${className || ""}`}
        style={{ backgroundColor: color, color: contrastColor(color), ...style }}
        draggable={draggable}
        onDragStart={onDragStart}
        onClick={onClick}
      >
        <span className={styles.eventChipLabel}>{label}</span>
        <button
          type="button"
          className={styles.copyBtn}
          style={{ color: contrastColor(color) }}
          onClick={handleCopy}
          aria-label={t("common.copy")}
        >
          <FaRegCopy />
        </button>
      </div>
    </Tooltip>
  );
};

export default CalendarEventChip;
