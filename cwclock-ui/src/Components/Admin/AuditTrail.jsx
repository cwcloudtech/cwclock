import React, { useCallback, useEffect, useRef, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaSearch, FaChevronLeft, FaChevronRight, FaRegCopy, FaCheck } from "react-icons/fa";
import SingleSelect from "../common/SingleSelect";
import { listAuditApi } from "../../Redux/Admin/Admin.actions";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/AuditTrail.module.css";

const PAGE_SIZE = 25;

// A per-column value + copy button, mirroring uprodit's AdminOperations CopyCell:
// the button reveals on row hover and flips from a copy icon to a check for a
// moment on success. Renders a muted dash (and no button) when there's nothing
// to copy, so every column reads consistently.
const CopyCell = ({ className = "", value, title, children }) => {
  const { t } = useI18n();
  const [copied, setCopied] = useState(false);
  const timer = useRef(null);

  useEffect(() => () => timer.current && clearTimeout(timer.current), []);

  const text = value == null ? "" : String(value);

  const copy = () => {
    if (!navigator.clipboard?.writeText) return;
    navigator.clipboard
      .writeText(text)
      .then(() => {
        setCopied(true);
        if (timer.current) clearTimeout(timer.current);
        timer.current = setTimeout(() => setCopied(false), 1500);
      })
      .catch(() => {});
  };

  return (
    <span className={`${styles.copyCell} ${className}`}>
      <span className={styles.copyValue} title={title}>
        {children != null && children !== "" ? children : <span className={styles.cellEmpty}>—</span>}
      </span>
      {text && (
        <button
          type="button"
          className={styles.copyBtn}
          onClick={copy}
          title={t("common.copy")}
          aria-label={t("common.copy")}
        >
          {copied ? <FaCheck /> : <FaRegCopy />}
        </button>
      )}
    </span>
  );
};

// AuditTrail is the superuser's view of what has been done across the whole
// deployment - who signed in, who changed a role, who touched an
// organization's credentials, and who added, moved or deleted the hours that
// get invoiced. It only ever reads: the trail is append-only server-side, so
// there is nothing here to edit or delete.
//
// There are two dropdown filters rather than one because time records are
// the highest-volume entries by far: without a subject-type filter, finding
// the handful of access changes in a busy month would mean paging past
// everyone's timesheets.
//
// Filtering and paging happen in the backend (GET /v1/admin/audit), not in
// this component: the trail only grows, so it can't be loaded whole and
// filtered client-side the way the users and organizations screens do.
const AuditTrail = () => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { audit, isLoading } = useSelector((state) => state.admin);

  const [page, setPage] = useState(1);
  const [action, setAction] = useState("");
  const [subjectType, setSubjectType] = useState("");
  // subjectInput is what's typed; subject is what's been submitted. Keeping
  // them apart means the trail isn't re-queried on every keystroke.
  const [subjectInput, setSubjectInput] = useState("");
  const [subject, setSubject] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");

  const load = useCallback(() => {
    dispatch(listAuditApi({ page, size: PAGE_SIZE, action, subjectType, subject, from, to }, user.token));
  }, [dispatch, page, action, subjectType, subject, from, to, user.token]);

  useEffect(() => {
    load();
  }, [load]);

  // Changing a filter has to go back to page 1: page 3 of the old filter is
  // rarely a page of the new one, and would often come back empty.
  const changeAction = (next) => {
    setAction(next);
    setPage(1);
  };

  const changeSubjectType = (next) => {
    setSubjectType(next);
    setPage(1);
  };

  const changeFrom = (next) => {
    setFrom(next);
    setPage(1);
  };

  const changeTo = (next) => {
    setTo(next);
    setPage(1);
  };

  const submitSubject = (e) => {
    e.preventDefault();
    setSubject(subjectInput.trim());
    setPage(1);
  };

  // An action with no translation yet (an older UI against a newer API)
  // falls back to the raw name rather than showing nothing - that's what
  // translate() does with an unknown key, and here it is the right answer:
  // the raw name is still readable.
  const actionLabel = (name) => t(`audit.actions.${name}`);
  const subjectLabel = (name) => t(`audit.subjects.${name}`);

  const formatDetail = (detail) =>
    Object.entries(detail || {})
      .map(([key, value]) => `${t(`audit.details.${key}`)}: ${formatValue(value)}`)
      .join(" · ");

  return (
    <div className={styles.main}>
      <h1 className="cw-title">{t("audit.title")}</h1>
      <p className="cw-subtitle">{t("audit.intro")}</p>

      <div className={styles.filters}>
        <div className={styles.filterRow}>
          <SingleSelect
            label={t("audit.action")}
            allLabel={t("audit.allActions")}
            value={action}
            onChange={changeAction}
            options={audit.actions.map((name) => ({ value: name, label: actionLabel(name) }))}
          />

          <SingleSelect
            label={t("audit.subjectType")}
            allLabel={t("audit.allSubjects")}
            value={subjectType}
            onChange={changeSubjectType}
            options={audit.subjectTypes.map((name) => ({ value: name, label: subjectLabel(name) }))}
          />

          <form className={styles.subjectForm} onSubmit={submitSubject}>
            <input
              className="cw-input"
              type="text"
              value={subjectInput}
              placeholder={t("audit.subjectPlaceholder")}
              onChange={(e) => setSubjectInput(e.target.value)}
            />
            <button type="submit" className="cw-button cw-button--sm" title={t("audit.filterBySubject")}>
              <FaSearch style={{ fontSize: "14px" }} />
            </button>
          </form>
        </div>

        {/* Date range on its own second row. */}
        <div className={styles.filterRow}>
          <label className={styles.dateField}>
            <span>{t("audit.from")}</span>
            <input
              className="cw-input"
              type="date"
              value={from}
              max={to || undefined}
              onChange={(e) => changeFrom(e.target.value)}
            />
          </label>
          <label className={styles.dateField}>
            <span>{t("audit.to")}</span>
            <input
              className="cw-input"
              type="date"
              value={to}
              min={from || undefined}
              onChange={(e) => changeTo(e.target.value)}
            />
          </label>
        </div>
      </div>

      {audit.entries.length === 0 ? (
        <div className={styles.empty}>
          <strong>{t("audit.emptyTitle")}</strong>
          <span>{t("audit.emptyBody")}</span>
        </div>
      ) : (
        <ul className={`cw-list ${styles.entryList}`}>
          {audit.entries.map((entry) => {
            const when = new Date(entry.createdAt).toLocaleString(locale);
            const actor = entry.userEmail || "";
            const detail = formatDetail(entry.detail);
            return (
              <li className={`cw-list-item ${styles.entryRow}`} key={entry.id}>
                <CopyCell
                  className={styles.when}
                  value={when}
                  title={new Date(entry.createdAt).toISOString()}
                >
                  {when}
                </CopyCell>
                <CopyCell className={styles.action} value={actionLabel(entry.action)}>
                  {actionLabel(entry.action)}
                </CopyCell>
                <CopyCell className={styles.actor} value={actor} title={actor}>
                  {entry.userEmail || <span className={styles.cellEmpty}>{t("audit.noActor")}</span>}
                </CopyCell>
                {/* The subject copies its id (what anyone actually pastes
                    elsewhere), falling back to the subject label. */}
                <CopyCell
                  className={styles.subject}
                  value={entry.subjectId || subjectLabel(entry.subjectType)}
                  title={entry.subjectId}
                >
                  {subjectLabel(entry.subjectType)}
                  {entry.subjectId && <code className={styles.subjectId}>{entry.subjectId}</code>}
                </CopyCell>
                <CopyCell className={styles.detail} value={detail} title={detail}>
                  {detail}
                </CopyCell>
              </li>
            );
          })}
        </ul>
      )}

      <div className={styles.pager}>
        <button
          type="button"
          className="cw-button cw-button--secondary cw-button--sm"
          onClick={() => setPage((p) => Math.max(1, p - 1))}
          disabled={page <= 1 || isLoading}
          title={t("audit.previousPage")}
        >
          <FaChevronLeft style={{ fontSize: "12px" }} /> {t("audit.previous")}
        </button>
        <span className={styles.pageLabel}>{t("audit.page", { page })}</span>
        <button
          type="button"
          className="cw-button cw-button--secondary cw-button--sm"
          onClick={() => setPage((p) => p + 1)}
          disabled={!audit.hasMore || isLoading}
          title={t("audit.nextPage")}
        >
          {t("audit.next")} <FaChevronRight style={{ fontSize: "12px" }} />
        </button>
      </div>
    </div>
  );
};

// formatValue keeps a detail value on one line: booleans as yes/no-ish text
// the dictionary owns, arrays joined, everything else as-is.
const formatValue = (value) => {
  if (Array.isArray(value)) return value.join(", ");
  if (typeof value === "boolean") return value ? "✓" : "✗";
  return String(value ?? "");
};

export default AuditTrail;
