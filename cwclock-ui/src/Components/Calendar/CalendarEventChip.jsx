import React, { useRef, useState } from "react";
import { createPortal } from "react-dom";
import { FaRegCopy } from "react-icons/fa";
import { toast } from "react-toastify";
import contrastColor from "../common/contrastColor";
import toastOptions from "../../Redux/toastOptions";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/CalendarEventChip.module.css";

// CalendarEventChip renders one time entry as a project-colored block, shared
// by the month grid, the week view's all-day row and its hourly grid
// (ai-instruct-84/85/86): the label truncates with an ellipsis when it
// doesn't fit, the full text is always available via a tooltip and a copy
// button, and (when draggable) the chip can be dragged onto another day to
// reschedule the entry.
//
// The tooltip is portaled to document.body and positioned from the chip's
// own getBoundingClientRect() (position: fixed), instead of using the app's
// shared Tooltip component's CSS-anchored bubble - month-view cells clip
// their own overflow to keep every day box the same fixed height regardless
// of content (ai-instruct-86), which would otherwise cut the bubble off (or
// hide it entirely) any time it needs to render outside its own cell
// (ai-instruct-87).
const CalendarEventChip = ({ label, color, draggable, onDragStart, onClick, className, style }) => {
  const { t } = useI18n();
  const chipRef = useRef(null);
  const [tooltipPos, setTooltipPos] = useState(null);

  const showTooltip = () => {
    const rect = chipRef.current?.getBoundingClientRect();
    if (rect) setTooltipPos({ top: rect.top, left: rect.left + rect.width / 2 });
  };
  const hideTooltip = () => setTooltipPos(null);

  const handleCopy = (e) => {
    e.stopPropagation();
    navigator.clipboard.writeText(label).then(() => {
      toast.success(t("common.copied"), toastOptions);
    });
  };

  return (
    <>
      <div
        ref={chipRef}
        className={`${styles.eventChip} ${className || ""}`}
        style={{ backgroundColor: color, color: contrastColor(color), ...style }}
        draggable={draggable}
        onDragStart={onDragStart}
        onClick={onClick}
        onMouseEnter={showTooltip}
        onMouseLeave={hideTooltip}
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
      {tooltipPos &&
        createPortal(
          <div className={styles.portalTooltip} style={{ top: tooltipPos.top, left: tooltipPos.left }} role="tooltip">
            {label}
          </div>,
          document.body
        )}
    </>
  );
};

export default CalendarEventChip;
