import React from "react";
import CalendarEventChip from "./CalendarEventChip";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/CalendarDayCell.module.css";

// Every cell renders at the same fixed height regardless of content (even an
// empty day) - entries beyond what fits are collapsed into a "+N more"
// indicator rather than growing the box or scrolling it (ai-instruct-86).
const MAX_VISIBLE_ENTRIES = 3;

// One month-grid cell: clicking empty space opens the "add a time record"
// modal for this day (defaulting to 9am-10am, ai-instruct-86), clicking an
// existing entry chip opens it for editing, and dragging a chip from another
// cell onto this one moves that time record here (ai-instruct-84). Weekend
// days get a subtly different background so they read as "not a usual
// working day" (ai-instruct-86).
const CalendarDayCell = ({ date, isCurrentMonth, isToday, isWeekend, entries, projects, onAddEntry, onEditEntry, onDropEntry }) => {
  const { t } = useI18n();

  const handleDrop = (e) => {
    e.preventDefault();
    const entryId = e.dataTransfer.getData("text/plain");
    if (entryId) onDropEntry(entryId, date);
  };

  const visibleEntries = entries.slice(0, MAX_VISIBLE_ENTRIES);
  const hiddenCount = entries.length - visibleEntries.length;

  const variantClass = !isCurrentMonth ? styles.outsideMonth : isWeekend ? styles.weekend : "";

  return (
    <div
      className={`${styles.cell} ${variantClass}`}
      onClick={() => onAddEntry(date)}
      onDragOver={(e) => e.preventDefault()}
      onDrop={handleDrop}
    >
      <span className={`${styles.dateNumber} ${isToday ? styles.today : ""}`}>{date.getDate()}</span>
      <div className={styles.events}>
        {visibleEntries.map((entry) => {
          const project = projects.find((p) => p.id === entry.projectId);
          const color = project?.color || "#1cb9f7";
          const label = entry.allDay ? entry.text : `${(entry.start || "").slice(0, 5)} ${entry.text}`;
          return (
            <CalendarEventChip
              key={entry.id}
              label={label}
              color={color}
              draggable
              onDragStart={(e) => e.dataTransfer.setData("text/plain", entry.id)}
              onClick={(e) => {
                e.stopPropagation();
                onEditEntry(entry);
              }}
            />
          );
        })}
        {hiddenCount > 0 && <span className={styles.moreLabel}>{t("calendar.moreEntries", { count: hiddenCount })}</span>}
      </div>
    </div>
  );
};

export default CalendarDayCell;
