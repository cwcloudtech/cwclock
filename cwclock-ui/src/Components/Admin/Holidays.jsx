import React, { useCallback, useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FiPlus, FiEdit2, FiTrash2 } from "react-icons/fi";
import Modal from "../common/Modal";
import ConfirmModal from "../common/ConfirmModal";
import {
  listHolidaysApi,
  createHolidayApi,
  updateHolidayApi,
  deleteHolidayApi,
} from "../../Redux/Admin/Admin.actions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import { toast } from "react-toastify";
import toastOptions from "../../Redux/toastOptions";
import styles from "./Styles/Holidays.module.css";

const TYPES = ["fixed", "easter", "weekend", "nthWeekday"];
// Monday-first for display; values are the lowercase English the API expects.
const WEEKDAYS = ["monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"];
// Which occurrence of a weekday in the month; -1 is the last one.
const NTHS = [1, 2, 3, 4, 5, -1];
const nthKey = (n) => (n === -1 ? "last" : String(n));

// Holidays is the superuser-only public-holiday (and weekend) decision table,
// keyed by ISO country code. The working-day fill reads these rows to decide
// which days are non-working: fixed dates, easter-relative movable feasts, and
// the configurable weekend. Ported alongside AuditTrail as an admin screen.
const Holidays = () => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);

  const [holidays, setHolidays] = useState([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(null); // null closed; {} new; row = edit
  const [confirm, setConfirm] = useState(null); // the row pending deletion

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await dispatch(listHolidaysApi(user.token));
      setHolidays(data || []);
    } catch (e) {
      toast.error(apiErrorMessage(e, locale), toastOptions);
    } finally {
      setLoading(false);
    }
  }, [dispatch, user.token, locale]);

  useEffect(() => {
    load();
  }, [load]);

  // Localised month name ("November") from a 1-based month number.
  const monthName = (m) =>
    m >= 1 && m <= 12 ? new Date(2000, m - 1, 1).toLocaleDateString(locale, { month: "long" }) : "";

  // A human-readable "when" for a rule row.
  const ruleWhen = (h) => {
    if (h.type === "fixed") {
      return `${String(h.day).padStart(2, "0")}/${String(h.month).padStart(2, "0")}`;
    }
    if (h.type === "easter") {
      return `${t("admin.holidays.easter")} ${h.easterOffset >= 0 ? "+" : ""}${h.easterOffset}`;
    }
    if (h.type === "nthWeekday") {
      return `${t(`admin.holidays.nth.${nthKey(h.nth)}`)} ${t(`admin.holidays.weekdays.${h.weekday}`)} · ${monthName(h.month)}`;
    }
    return t(`admin.holidays.weekdays.${h.weekday}`);
  };

  const handleDelete = async () => {
    const row = confirm;
    setConfirm(null);
    if (!row) return;
    try {
      await dispatch(deleteHolidayApi(row.id, user.token));
      setHolidays((hs) => hs.filter((h) => h.id !== row.id));
    } catch (e) {
      toast.error(apiErrorMessage(e, locale), toastOptions);
    }
  };

  const handleSaved = () => {
    setEditing(null);
    load();
  };

  // Group by country, countries alphabetical.
  const byCountry = {};
  holidays.forEach((h) => {
    (byCountry[h.country] = byCountry[h.country] || []).push(h);
  });
  const countries = Object.keys(byCountry).sort();

  return (
    <div className={styles.main}>
      <div className={styles.header}>
        <div>
          <h1 className="cw-title">{t("admin.holidays.title")}</h1>
          <p className="cw-subtitle">{t("admin.holidays.intro")}</p>
        </div>
        <button type="button" className="cw-button" onClick={() => setEditing({})}>
          <FiPlus style={{ marginRight: 6 }} />
          {t("admin.holidays.add")}
        </button>
      </div>

      {loading && holidays.length === 0 ? (
        <p className={styles.muted}>{t("common.loading")}</p>
      ) : countries.length === 0 ? (
        <div className={styles.empty}>
          <strong>{t("admin.holidays.emptyTitle")}</strong>
          <span>{t("admin.holidays.emptyBody")}</span>
        </div>
      ) : (
        countries.map((country) => (
          <section key={country} className={styles.country}>
            <h2 className={styles.countryTitle}>{country}</h2>
            <ul className={`cw-list ${styles.list}`}>
              {byCountry[country].map((h) => (
                <li className={`cw-list-item ${styles.row}`} key={h.id}>
                  <span className={styles.when}>{ruleWhen(h)}</span>
                  <span className={styles.name}>{h.name}</span>
                  <span className={`${styles.badge} ${styles[`badge_${h.type}`]}`}>
                    {t(`admin.holidays.types.${h.type}`)}
                  </span>
                  <div className={styles.actions}>
                    <button
                      type="button"
                      className={styles.iconBtn}
                      onClick={() => setEditing(h)}
                      title={t("common.edit")}
                      aria-label={t("common.edit")}
                    >
                      <FiEdit2 />
                    </button>
                    <button
                      type="button"
                      className={`${styles.iconBtn} ${styles.danger}`}
                      onClick={() => setConfirm(h)}
                      title={t("common.delete")}
                      aria-label={t("common.delete")}
                    >
                      <FiTrash2 />
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          </section>
        ))
      )}

      {editing !== null && (
        <HolidayModal
          holiday={editing}
          token={user.token}
          onClose={() => setEditing(null)}
          onSaved={handleSaved}
        />
      )}

      <ConfirmModal
        show={confirm !== null}
        title={t("admin.holidays.deleteTitle")}
        body={t("admin.holidays.deleteBody", { name: confirm?.name || "" })}
        confirmLabel={t("common.delete")}
        onConfirm={handleDelete}
        onCancel={() => setConfirm(null)}
      />
    </div>
  );
};

// HolidayModal creates or edits one rule. The visible fields follow the type:
// a fixed date needs month/day, an easter rule its offset, a weekend rule its
// weekday.
const HolidayModal = ({ holiday, token, onClose, onSaved }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const isEdit = !!holiday.id;

  const [country, setCountry] = useState(holiday.country || "");
  const [name, setName] = useState(holiday.name || "");
  const [type, setType] = useState(holiday.type || "fixed");
  const [month, setMonth] = useState(holiday.month || 1);
  const [day, setDay] = useState(holiday.day || 1);
  const [easterOffset, setEasterOffset] = useState(holiday.easterOffset ?? 0);
  const [weekday, setWeekday] = useState(holiday.weekday || "saturday");
  const [nth, setNth] = useState(holiday.nth ?? 1);
  const [saving, setSaving] = useState(false);

  const handleSave = async () => {
    if (country.trim().length !== 2 || !name.trim()) return;
    const body = { country: country.trim().toUpperCase(), name: name.trim(), type };
    if (type === "fixed") {
      body.month = Number(month);
      body.day = Number(day);
    } else if (type === "easter") {
      body.easterOffset = Number(easterOffset);
    } else if (type === "nthWeekday") {
      body.month = Number(month);
      body.weekday = weekday;
      body.nth = Number(nth);
    } else {
      body.weekday = weekday;
    }
    setSaving(true);
    try {
      await dispatch(isEdit ? updateHolidayApi(holiday.id, body, token) : createHolidayApi(body, token));
      onSaved();
    } catch (e) {
      toast.error(apiErrorMessage(e, locale), toastOptions);
    } finally {
      setSaving(false);
    }
  };

  const footer = (
    <>
      <button type="button" className="cw-button cw-button--secondary" onClick={onClose} disabled={saving}>
        {t("common.cancel")}
      </button>
      <button
        type="button"
        className="cw-button"
        onClick={handleSave}
        disabled={saving || country.trim().length !== 2 || !name.trim()}
      >
        {t("common.save")}
      </button>
    </>
  );

  return (
    <Modal
      show
      title={isEdit ? t("admin.holidays.editTitle") : t("admin.holidays.addTitle")}
      onClose={onClose}
      footer={footer}
    >
      <div className={styles.form}>
        <label className={styles.field}>
          <span>{t("admin.holidays.country")}</span>
          <input
            className="cw-input"
            type="text"
            maxLength={2}
            value={country}
            placeholder="FR"
            onChange={(e) => setCountry(e.target.value.toUpperCase())}
          />
        </label>
        <label className={styles.field}>
          <span>{t("admin.holidays.name")}</span>
          <input className="cw-input" type="text" value={name} onChange={(e) => setName(e.target.value)} />
        </label>
        <label className={styles.field}>
          <span>{t("admin.holidays.type")}</span>
          <select className="cw-input" value={type} onChange={(e) => setType(e.target.value)}>
            {TYPES.map((ty) => (
              <option key={ty} value={ty}>
                {t(`admin.holidays.types.${ty}`)}
              </option>
            ))}
          </select>
        </label>

        {type === "fixed" && (
          <div className={styles.row2}>
            <label className={styles.field}>
              <span>{t("admin.holidays.day")}</span>
              <input className="cw-input" type="number" min={1} max={31} value={day} onChange={(e) => setDay(e.target.value)} />
            </label>
            <label className={styles.field}>
              <span>{t("admin.holidays.month")}</span>
              <input className="cw-input" type="number" min={1} max={12} value={month} onChange={(e) => setMonth(e.target.value)} />
            </label>
          </div>
        )}
        {type === "easter" && (
          <label className={styles.field}>
            <span>{t("admin.holidays.easterOffset")}</span>
            <input
              className="cw-input"
              type="number"
              min={-60}
              max={60}
              value={easterOffset}
              onChange={(e) => setEasterOffset(e.target.value)}
            />
            <small className={styles.hint}>{t("admin.holidays.easterHint")}</small>
          </label>
        )}
        {type === "weekend" && (
          <label className={styles.field}>
            <span>{t("admin.holidays.weekday")}</span>
            <select className="cw-input" value={weekday} onChange={(e) => setWeekday(e.target.value)}>
              {WEEKDAYS.map((wd) => (
                <option key={wd} value={wd}>
                  {t(`admin.holidays.weekdays.${wd}`)}
                </option>
              ))}
            </select>
          </label>
        )}
        {type === "nthWeekday" && (
          <div className={styles.row2}>
            <label className={styles.field}>
              <span>{t("admin.holidays.occurrence")}</span>
              <select className="cw-input" value={nth} onChange={(e) => setNth(e.target.value)}>
                {NTHS.map((n) => (
                  <option key={n} value={n}>
                    {t(`admin.holidays.nth.${nthKey(n)}`)}
                  </option>
                ))}
              </select>
            </label>
            <label className={styles.field}>
              <span>{t("admin.holidays.weekday")}</span>
              <select className="cw-input" value={weekday} onChange={(e) => setWeekday(e.target.value)}>
                {WEEKDAYS.map((wd) => (
                  <option key={wd} value={wd}>
                    {t(`admin.holidays.weekdays.${wd}`)}
                  </option>
                ))}
              </select>
            </label>
            <label className={styles.field}>
              <span>{t("admin.holidays.month")}</span>
              <input className="cw-input" type="number" min={1} max={12} value={month} onChange={(e) => setMonth(e.target.value)} />
            </label>
          </div>
        )}
      </div>
    </Modal>
  );
};

export default Holidays;
