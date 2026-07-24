import React from "react";
import contrastColor from "../common/contrastColor";
import styles from "./Styles/CalendarDayCell.module.css";

// One month-grid cell: clicking empty space opens the "add a time record"
// modal for this day, clicking an existing entry chip opens it for editing,
// and dragging a chip from another cell onto this one moves that time
// record here - the drag-and-drop "add/reschedule a meeting" UX from the
// spec (ai-instruct-84). Each chip is colored with its project's own color.
const CalendarDayCell = ({ date, isCurrentMonth, isToday, entries, projects, onAddEntry, onEditEntry, onDropEntry }) => {
  const handleDrop = (e) => {
    e.preventDefault();
    const entryId = e.dataTransfer.getData("text/plain");
    if (entryId) onDropEntry(entryId, date);
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
            <div
              key={entry.id}
              className={styles.eventChip}
              style={{ backgroundColor: color, color: contrastColor(color) }}
              draggable
              title={label}
              onDragStart={(e) => e.dataTransfer.setData("text/plain", entry.id)}
              onClick={(e) => {
                e.stopPropagation();
                onEditEntry(entry);
              }}
            >
              {label}
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default CalendarDayCell;
