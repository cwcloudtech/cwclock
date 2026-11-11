import React, { useCallback, useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaSearch, FaChevronLeft, FaChevronRight } from "react-icons/fa";
import CopyIdButton from "../common/CopyIdButton";
import { listAuditApi } from "../../Redux/Admin/Admin.actions";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/AuditTrail.module.css";

const PAGE_SIZE = 25;

// AuditTrail is the superuser's view of what has been done across the whole
// deployment - who signed in, who changed a role, who touched an
// organization's credentials. It only ever reads: the trail is append-only
// server-side, so there is nothing here to edit or delete.
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
  // subjectInput is what's typed; subject is what's been submitted. Keeping
  // them apart means the trail isn't re-queried on every keystroke.
  const [subjectInput, setSubjectInput] = useState("");
  const [subject, setSubject] = useState("");

  const load = useCallback(() => {
    dispatch(listAuditApi({ page, size: PAGE_SIZE, action, subject }, user.token));
  }, [dispatch, page, action, subject, user.token]);

  useEffect(() => {
    load();
  }, [load]);

  // Changing a filter has to go back to page 1: page 3 of the old filter is
  // rarely a page of the new one, and would often come back empty.
  const changeAction = (next) => {
    setAction(next);
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

  const formatDetail = (detail) =>
    Object.entries(detail || {})
      .map(([key, value]) => `${t(`audit.details.${key}`)}: ${formatValue(value)}`)
      .join(" · ");

  return (
    <div className={styles.main}>
      <h1 className="cw-title">{t("audit.title")}</h1>
      <p className="cw-subtitle">{t("audit.intro")}</p>

      <div className={styles.filters}>
        <select
          className="cw-select"
          value={action}
          onChange={(e) => changeAction(e.target.value)}
          title={t("audit.filterByAction")}
        >
          <option value="">{t("audit.allActions")}</option>
          {audit.actions.map((name) => (
            <option key={name} value={name}>
              {actionLabel(name)}
            </option>
          ))}
        </select>

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

      {audit.entries.length === 0 ? (
        <div className={styles.empty}>
          <strong>{t("audit.emptyTitle")}</strong>
          <span>{t("audit.emptyBody")}</span>
        </div>
      ) : (
        <ul className={`cw-list ${styles.entryList}`}>
          {audit.entries.map((entry) => (
            <li className={`cw-list-item ${styles.entryRow}`} key={entry.id}>
              <span className={styles.when} title={new Date(entry.createdAt).toISOString()}>
                {new Date(entry.createdAt).toLocaleString(locale)}
              </span>
              <span className={styles.action}>{actionLabel(entry.action)}</span>
              <span className={styles.actor}>{entry.userEmail || t("audit.noActor")}</span>
              <span className={styles.subject} title={entry.subjectId}>
                {t(`audit.subjects.${entry.subjectType}`)}
                {entry.subjectId && <code className={styles.subjectId}>{entry.subjectId}</code>}
              </span>
              {/* Cells are ellipsised to keep one entry on one row, so the
                  full text has to be reachable - hence the title here and on
                  the subject above. */}
              <span className={styles.detail} title={formatDetail(entry.detail)}>
                {formatDetail(entry.detail)}
              </span>
              <div className={styles.rowActions}>
                <CopyIdButton id={entry.subjectId} className={styles.iconBtn} />
              </div>
            </li>
          ))}
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
