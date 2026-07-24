import React, { useEffect, useMemo, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FiChevronLeft, FiChevronRight } from "react-icons/fi";
import { FaPlug } from "react-icons/fa";
import Spinner from "../spinner/Spinner";
import NeedOrganizationEmptyState from "../common/NeedOrganizationEmptyState";
import Tooltip from "../common/Tooltip";
import CalendarDayCell from "./CalendarDayCell";
import CalendarEventModal from "./CalendarEventModal";
import CalendarShareModal from "./CalendarShareModal";
import {
  getCalendarEntriesApi,
  createCalendarEntryApi,
  updateCalendarEntryApi,
  deleteCalendarEntryApi,
} from "../../Redux/Calendar/Calendar.actions";
import { listClientsApi } from "../../Redux/Clients/Client.actions";
import { listProjectsApi } from "../../Redux/Projects/Project.actions";
import { toISODate } from "../common/dateRangeShortcuts";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/Calendar.module.css";

const startOfDay = (d) => {
  const x = new Date(d);
  x.setHours(0, 0, 0, 0);
  return x;
};
const addDays = (d, n) => {
  const x = new Date(d);
  x.setDate(x.getDate() + n);
  return x;
};
// Weeks are Sunday-first, matching this app's day-of-week indexing (see the
// i18n "days" dictionary and dateRangeShortcuts.js, both 0 = Sunday).
const startOfWeek = (d) => addDays(startOfDay(d), -d.getDay());

// Moves d by delta months, clamping the day-of-month to the target month's
// last day instead of letting native Date overflow into the month after
// (e.g. Jan 31 minus one month would otherwise silently land on Mar 3, since
// February has no 31st).
const shiftMonth = (d, delta) => {
  const target = new Date(d.getFullYear(), d.getMonth() + delta, 1);
  const lastDayOfTarget = new Date(target.getFullYear(), target.getMonth() + 1, 0).getDate();
  target.setDate(Math.min(d.getDate(), lastDayOfTarget));
  return target;
};

// Calendar is the month-grid time tracking view (ai-instruct-84): same idea
// as Google Calendar/Teams calendar - click an empty day to add a time
// record, drag an existing one onto another day to reschedule it, and each
// entry is colored with its project's color.
const Calendar = () => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const [anchorDate, setAnchorDate] = useState(() => startOfDay(new Date()));
  const [viewMode, setViewMode] = useState("month"); // "month" | "week"
  const [modalState, setModalState] = useState(null); // null | { day } | { entry }
  const [showShareModal, setShowShareModal] = useState(false);

  const { entries, isLoading } = useSelector((state) => state.calendar);
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId } = useSelector((state) => state.organizations);
  const { projects } = useSelector((state) => state.projects);
  const { clients } = useSelector((state) => state.clients);

  useEffect(() => {
    if (currentOrgId) {
      dispatch(listClientsApi(currentOrgId, user.token));
      dispatch(listProjectsApi(currentOrgId, user.token));
    }
  }, [dispatch, currentOrgId, user.token]);

  const monthStart = useMemo(
    () => new Date(anchorDate.getFullYear(), anchorDate.getMonth(), 1),
    [anchorDate]
  );
  const monthEnd = useMemo(
    () => new Date(anchorDate.getFullYear(), anchorDate.getMonth() + 1, 0),
    [anchorDate]
  );
  // Week view shows just the 7-day week containing anchorDate, month view
  // shows every full week overlapping the month - both render through the
  // same grid/cell components at the same size (ai-instruct-85).
  const gridStart = useMemo(
    () => (viewMode === "week" ? startOfWeek(anchorDate) : startOfWeek(monthStart)),
    [viewMode, anchorDate, monthStart]
  );
  const gridEnd = useMemo(
    () => (viewMode === "week" ? addDays(gridStart, 6) : addDays(startOfWeek(monthEnd), 6)),
    [viewMode, gridStart, monthEnd]
  );

  const weeks = useMemo(() => {
    const days = [];
    for (let d = new Date(gridStart); d <= gridEnd; d = addDays(d, 1)) {
      days.push(d);
    }
    const rows = [];
    for (let i = 0; i < days.length; i += 7) rows.push(days.slice(i, i + 7));
    return rows;
  }, [gridStart, gridEnd]);

  useEffect(() => {
    if (currentOrgId) {
      dispatch(getCalendarEntriesApi(currentOrgId, user.token, toISODate(gridStart), toISODate(gridEnd)));
    }
  }, [currentOrgId, user.token, dispatch, gridStart, gridEnd]);

  const entriesByDay = useMemo(() => {
    const map = {};
    for (const entry of entries) {
      (map[entry.day] ||= []).push(entry);
    }
    return map;
  }, [entries]);

  const todayIso = toISODate(startOfDay(new Date()));

  const rangeLabel = useMemo(() => {
    const intlLocale = locale === "fr" ? "fr-FR" : "en-US";
    if (viewMode === "week") {
      const dayFmt = new Intl.DateTimeFormat(intlLocale, { day: "numeric", month: "short" });
      const yearFmt = new Intl.DateTimeFormat(intlLocale, { year: "numeric" });
      return `${dayFmt.format(gridStart)} - ${dayFmt.format(gridEnd)} ${yearFmt.format(gridEnd)}`;
    }
    return new Intl.DateTimeFormat(intlLocale, { month: "long", year: "numeric" }).format(monthStart);
  }, [viewMode, gridStart, gridEnd, monthStart, locale]);

  const weekdayLabels = useMemo(() => [0, 1, 2, 3, 4, 5, 6].map((i) => t(`days.${i}`).slice(0, 3)), [t]);

  const handleAddEntry = (date) => setModalState({ day: toISODate(date) });
  const handleEditEntry = (entry) => setModalState({ entry });
  const handleCloseModal = () => setModalState(null);

  const handleDropEntry = (entryId, date) => {
    const entry = entries.find((e) => e.id === entryId);
    const day = toISODate(date);
    if (!entry || entry.day === day) return;
    dispatch(updateCalendarEntryApi({ ...entry, day }, currentOrgId, user.token)).catch(() => {});
  };

  // Errors are already toasted by the Calendar redux actions themselves; the
  // .catch here only prevents an unhandled-rejection warning, it doesn't
  // close the modal, so the user can fix the form and retry.
  const handleSave = (payload) => {
    const action = payload.id
      ? updateCalendarEntryApi(payload, currentOrgId, user.token)
      : createCalendarEntryApi(payload, currentOrgId, user.token);
    return dispatch(action)
      .then(() => setModalState(null))
      .catch(() => {});
  };

  const handleDelete = (id) =>
    dispatch(deleteCalendarEntryApi(id, currentOrgId, user.token))
      .then(() => setModalState(null))
      .catch(() => {});

  if (!currentOrgId) {
    return (
      <div className={styles.body}>
        <NeedOrganizationEmptyState body={t("timeTracker.needOrganization")} />
      </div>
    );
  }

  return (
    <div className={styles.body}>
      <div className={styles.header}>
        <h2 className={styles.monthLabel}>{rangeLabel}</h2>
        <div className={styles.navButtons}>
          <div className={styles.viewToggle}>
            <button
              type="button"
              className={`${styles.viewButton} ${viewMode === "month" ? styles.viewButtonActive : ""}`}
              onClick={() => setViewMode("month")}
            >
              {t("calendar.monthView")}
            </button>
            <button
              type="button"
              className={`${styles.viewButton} ${viewMode === "week" ? styles.viewButtonActive : ""}`}
              onClick={() => setViewMode("week")}
            >
              {t("calendar.weekView")}
            </button>
          </div>
          <button
            type="button"
            className="cw-button cw-button--secondary cw-button--sm"
            onClick={() => setAnchorDate(startOfDay(new Date()))}
          >
            {t("calendar.today")}
          </button>
          <button
            type="button"
            className={styles.navButton}
            onClick={() => setAnchorDate((d) => (viewMode === "week" ? addDays(d, -7) : shiftMonth(d, -1)))}
            aria-label={t("calendar.previousMonth")}
            title={t("calendar.previousMonth")}
          >
            <FiChevronLeft />
          </button>
          <button
            type="button"
            className={styles.navButton}
            onClick={() => setAnchorDate((d) => (viewMode === "week" ? addDays(d, 7) : shiftMonth(d, 1)))}
            aria-label={t("calendar.nextMonth")}
            title={t("calendar.nextMonth")}
          >
            <FiChevronRight />
          </button>
          <Tooltip label={t("calendar.shareTitle")} position="bottom">
            <button
              type="button"
              className={styles.navButton}
              onClick={() => setShowShareModal(true)}
              aria-label={t("calendar.shareTitle")}
            >
              <FaPlug />
            </button>
          </Tooltip>
        </div>
      </div>

      <div className={styles.weekdayRow}>
        {weekdayLabels.map((label, i) => (
          <div key={i} className={styles.weekdayCell}>
            {label}
          </div>
        ))}
      </div>

      <div className={styles.grid}>
        {weeks.map((week, wi) => (
          <div key={wi} className={styles.weekRow}>
            {week.map((date) => {
              const iso = toISODate(date);
              return (
                <CalendarDayCell
                  key={iso}
                  date={date}
                  isCurrentMonth={viewMode === "week" || date.getMonth() === monthStart.getMonth()}
                  isToday={iso === todayIso}
                  entries={entriesByDay[iso] || []}
                  projects={projects}
                  onAddEntry={handleAddEntry}
                  onEditEntry={handleEditEntry}
                  onDropEntry={handleDropEntry}
                />
              );
            })}
          </div>
        ))}
      </div>

      {isLoading && <Spinner />}

      <CalendarEventModal
        show={!!modalState}
        entry={modalState?.entry || null}
        defaultDay={modalState?.day || todayIso}
        projects={projects}
        clients={clients}
        onClose={handleCloseModal}
        onSave={handleSave}
        onDelete={handleDelete}
      />

      <CalendarShareModal show={showShareModal} onClose={() => setShowShareModal(false)} />
    </div>
  );
};

export default Calendar;
