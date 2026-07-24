import React from "react";
import { FaRegCopy } from "react-icons/fa";
import { toast } from "react-toastify";
import contrastColor from "../common/contrastColor";
import Tooltip from "../common/Tooltip";
import toastOptions from "../../Redux/toastOptions";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/CalendarDayCell.module.css";

// One grid cell (shared by both the month and week views - ai-instruct-85
// keeps the same cell size in both, so this component doesn't need to know
// which view it's in). Clicking empty space opens the "add a time record"
// modal for this day, clicking an existing entry chip opens it for editing,
// and dragging a chip from another cell onto this one moves that time
// record here (ai-instruct-84). Each chip is colored with its project's own
// color; a long label is truncated, shown in full via a tooltip, and has its
// own copy button (ai-instruct-85).
const CalendarDayCell = ({ date, isCurrentMonth, isToday, entries, projects, onAddEntry, onEditEntry, onDropEntry }) => {
  const { t } = useI18n();

  const handleDrop = (e) => {
    e.preventDefault();
    const entryId = e.dataTransfer.getData("text/plain");
    if (entryId) onDropEntry(entryId, date);
  };

  const handleCopy = (e, label) => {
    e.stopPropagation();
    navigator.clipboard.writeText(label).then(() => {
      toast.success(t("common.copied"), toastOptions);
    });
  };

  return (
    <div
      className={`${styles.cell} ${isCurrentMonth ? "" : styles.outsideMonth}`}
      onClick={() => onAddEntry(date)}
      onDragOver={(e) => e.preventDefault()}
      onDrop={handleDrop}
    >
      <span className={`${styles.dateNumber} ${isToday ? styles.today : ""}`}>{date.getDate()}</span>
      <div className={styles.events}>
        {entries.map((entry) => {
          const project = projects.find((p) => p.id === entry.projectId);
          const color = project?.color || "#1cb9f7";
          const label = entry.allDay ? entry.text : `${(entry.start || "").slice(0, 5)} ${entry.text}`;
          return (
            <Tooltip key={entry.id} label={label} className={styles.chipTooltip}>
              <div
                className={styles.eventChip}
                style={{ backgroundColor: color, color: contrastColor(color) }}
                draggable
                onDragStart={(e) => e.dataTransfer.setData("text/plain", entry.id)}
                onClick={(e) => {
                  e.stopPropagation();
                  onEditEntry(entry);
                }}
              >
                <span className={styles.eventChipLabel}>{label}</span>
                <button
                  type="button"
                  className={styles.copyBtn}
                  style={{ color: contrastColor(color) }}
                  onClick={(e) => handleCopy(e, label)}
                  aria-label={t("common.copy")}
                >
                  <FaRegCopy />
                </button>
              </div>
            </Tooltip>
          );
        })}
      </div>
    </div>
  );
};

export default CalendarDayCell;
