import React, { useEffect, useRef, useState } from "react";
import CalendarEventChip from "./CalendarEventChip";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/CalendarWeekView.module.css";

const HOURS = Array.from({ length: 24 }, (_, i) => i);
const HOUR_HEIGHT = 48;
const HALF_HOUR_HEIGHT = HOUR_HEIGHT / 2;
// Selection granularity: each day column is 48 half-hour slots rather than
// 24 hour slots, so dragging can start/end on the half hour (ai-instruct-87).
const SLOTS = Array.from({ length: 24 * 2 }, (_, i) => i);
// Weekly views scroll to this hour on open so working hours are visible
// without the user having to scroll up from midnight (ai-instruct-86).
const INITIAL_SCROLL_HOUR = 7;

const minutesOfDay = (hms) => {
  if (!hms) return 0;
  const [h, m] = hms.split(":").map(Number);
  return (h || 0) * 60 + (m || 0);
};

// Converts a half-hour slot index back into a "HH:MM:SS" time string. Slot
// 48 (the end of a selection reaching all the way to midnight) has no valid
// 24:00 hour, so it's reported as the last second of the day instead.
const slotToTime = (slot) => {
  if (slot >= SLOTS.length) return "23:59:59";
  const hour = Math.floor(slot / 2);
  const minute = (slot % 2) * 30;
  return `${String(hour).padStart(2, "0")}:${String(minute).padStart(2, "0")}:00`;
};

// CalendarWeekView is the week view's hour-by-hour grid (ai-instruct-86):
// each day is a 24h column instead of the month view's small box. Dragging
// down across half-hour cells in a single column selects that exact time
// range (snapped to the half hour, ai-instruct-87) and opens the "add a time
// record" modal prefilled with it, instead of the month view's fixed
// 9am-10am default. Existing entries can still be dragged onto another
// day's column to reschedule them (same day-level move as the month view -
// the hour they're dropped on doesn't change their time).
const CalendarWeekView = ({ days, entriesByDay, projects, todayIso, onAddEntry, onEditEntry, onDropEntry }) => {
  const { t } = useI18n();
  const [dragSelection, setDragSelection] = useState(null); // { date, iso, startSlot, endSlot }
  const scrollRef = useRef(null);

  useEffect(() => {
    if (scrollRef.current) scrollRef.current.scrollTop = INITIAL_SCROLL_HOUR * HOUR_HEIGHT;
  }, []);

  // A plain click (mousedown+mouseup with no mouseenter into another cell)
  // leaves startSlot === endSlot, which finishDrag turns into a single
  // half-hour block - the same gesture handles both "click one slot" and
  // "drag across several".
  useEffect(() => {
    if (!dragSelection) return;
    const handleMouseUp = () => {
      setDragSelection((sel) => {
        if (sel) {
          const from = Math.min(sel.startSlot, sel.endSlot);
          const to = Math.max(sel.startSlot, sel.endSlot) + 1;
          onAddEntry(sel.date, slotToTime(from), slotToTime(to));
        }
        return null;
      });
    };
    window.addEventListener("mouseup", handleMouseUp);
    return () => window.removeEventListener("mouseup", handleMouseUp);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dragSelection]);

  const isSlotSelected = (iso, slot) => {
    if (!dragSelection || dragSelection.iso !== iso) return false;
    const lo = Math.min(dragSelection.startSlot, dragSelection.endSlot);
    const hi = Math.max(dragSelection.startSlot, dragSelection.endSlot);
    return slot >= lo && slot <= hi;
  };

  const handleDayDrop = (e, date) => {
    e.preventDefault();
    const entryId = e.dataTransfer.getData("text/plain");
    if (entryId) onDropEntry(entryId, date);
  };

  return (
    <div className={styles.weekGrid}>
      <div className={styles.headerRow}>
        <div className={styles.axisCell} />
        {days.map((d) => (
          <div
            key={d.iso}
            className={`${styles.dayHeader} ${d.isWeekend ? styles.weekend : ""} ${d.iso === todayIso ? styles.today : ""}`}
          >
            <span className={styles.dayHeaderName}>{d.label}</span>
            <span className={styles.dayHeaderNumber}>{d.date.getDate()}</span>
          </div>
        ))}
      </div>

      <div className={styles.allDayRow}>
        <div className={styles.axisAllDayLabel}>{t("calendar.allDayRowLabel")}</div>
        {days.map((d) => (
          <div key={d.iso} className={`${styles.allDayCell} ${d.isWeekend ? styles.weekend : ""}`}>
            {(entriesByDay[d.iso] || [])
              .filter((entry) => entry.allDay)
              .map((entry) => {
                const project = projects.find((p) => p.id === entry.projectId);
                return (
                  <CalendarEventChip
                    key={entry.id}
                    label={entry.text}
                    color={project?.color || "#1cb9f7"}
                    onClick={() => onEditEntry(entry)}
                  />
                );
              })}
          </div>
        ))}
      </div>

      <div className={styles.hourScroll} ref={scrollRef}>
        <div className={styles.hourBody} style={{ height: HOURS.length * HOUR_HEIGHT }}>
          <div className={styles.hourAxis}>
            {HOURS.map((h) => (
              <div key={h} className={styles.hourLabel} style={{ height: HOUR_HEIGHT }}>
                {String(h).padStart(2, "0")}:00
              </div>
            ))}
          </div>
          {days.map((d) => (
            <div
              key={d.iso}
              className={`${styles.dayColumn} ${d.isWeekend ? styles.weekend : ""}`}
              onDragOver={(e) => e.preventDefault()}
              onDrop={(e) => handleDayDrop(e, d.date)}
            >
              {SLOTS.map((slot) => (
                <div
                  key={slot}
                  className={`${styles.hourCell} ${slot % 2 === 0 ? styles.halfHourBoundary : styles.hourBoundary} ${
                    isSlotSelected(d.iso, slot) ? styles.hourCellSelected : ""
                  }`}
                  style={{ height: HALF_HOUR_HEIGHT }}
                  onMouseDown={() => setDragSelection({ date: d.date, iso: d.iso, startSlot: slot, endSlot: slot })}
                  onMouseEnter={() =>
                    setDragSelection((sel) => (sel && sel.iso === d.iso ? { ...sel, endSlot: slot } : sel))
                  }
                />
              ))}
              {(entriesByDay[d.iso] || [])
                .filter((entry) => !entry.allDay)
                .map((entry) => {
                  const project = projects.find((p) => p.id === entry.projectId);
                  const start = minutesOfDay(entry.start);
                  const end = Math.max(minutesOfDay(entry.end), start + 15);
                  const top = (start / 60) * HOUR_HEIGHT;
                  const height = Math.max(((end - start) / 60) * HOUR_HEIGHT, 18);
                  const label = `${(entry.start || "").slice(0, 5)} ${entry.text}`;
                  return (
                    <div key={entry.id} className={styles.timedEntryWrapper} style={{ top, height }}>
                      <CalendarEventChip
                        label={label}
                        color={project?.color || "#1cb9f7"}
                        style={{ height: "100%" }}
                        draggable
                        onDragStart={(e) => e.dataTransfer.setData("text/plain", entry.id)}
                        onClick={(e) => {
                          e.stopPropagation();
                          onEditEntry(entry);
                        }}
                      />
                    </div>
                  );
                })}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default CalendarWeekView;
