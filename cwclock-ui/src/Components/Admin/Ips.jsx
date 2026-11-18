import React, { useCallback, useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaSearch, FaChevronLeft, FaChevronRight, FaBan, FaUnlock, FaMapMarkerAlt } from "react-icons/fa";
import CopyIdButton from "../common/CopyIdButton";
import Tooltip from "../common/Tooltip";
import Modal from "../common/Modal";
import ConfirmModal from "../common/ConfirmModal";
import { listIpBansApi, ipStatsApi, putIpBanApi } from "../../Redux/Admin/Admin.actions";
import { useI18n } from "../../i18n/I18nContext";
import { geoLabel, networkLabel } from "./geoLabel";
import styles from "./Styles/Ips.module.css";

// IpManagement is the superuser's roster of every address seen (plus any banned
// before being seen), with its members, geolocation, connection count and ban
// state - and the ban/unban action. Filtering and paging happen in the backend
// (GET /v1/admin/ip-bans), which returns { results, page, pageSize, hasNext }.
const IpManagement = () => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);

  const [page, setPage] = useState(0);
  // termInput is what's typed; term is what's been submitted - so the roster
  // isn't re-queried on every keystroke.
  const [termInput, setTermInput] = useState("");
  const [term, setTerm] = useState("");
  const [data, setData] = useState({ results: [], page: 0, pageSize: 50, hasNext: false });
  const [stats, setStats] = useState({ total: 0, banned: 0 });
  const [loading, setLoading] = useState(false);
  const [busyIp, setBusyIp] = useState(null);

  const [banning, setBanning] = useState(false); // "ban an address" modal
  const [banIpValue, setBanIpValue] = useState("");
  const [banMotif, setBanMotif] = useState("");
  const [unbanning, setUnbanning] = useState(null); // ip pending unban confirm

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const d = await dispatch(listIpBansApi(term, page, user.token));
      setData(d || { results: [], page, pageSize: 50, hasNext: false });
    } catch {
      /* error toast shown in the thunk */
    } finally {
      setLoading(false);
    }
  }, [dispatch, term, page, user.token]);

  const loadStats = useCallback(async () => {
    try {
      setStats((await dispatch(ipStatsApi(user.token))) || { total: 0, banned: 0 });
    } catch {
      /* stats are a nicety, not worth a toast */
    }
  }, [dispatch, user.token]);

  useEffect(() => {
    load();
  }, [load]);
  useEffect(() => {
    loadStats();
  }, [loadStats]);

  const refresh = () => {
    load();
    loadStats();
  };

  const submitTerm = (e) => {
    e.preventDefault();
    setTerm(termInput.trim());
    setPage(0);
  };

  const confirmBan = async (e) => {
    e?.preventDefault?.();
    const ip = banIpValue.trim();
    if (!ip) return;
    try {
      await dispatch(putIpBanApi({ ip, action: "ban", motif: banMotif.trim() }, user.token));
      setBanning(false);
      setBanIpValue("");
      setBanMotif("");
      refresh();
    } catch {
      /* stays open so the admin can fix the address */
    }
  };

  const banRow = async (ip) => {
    setBusyIp(ip);
    try {
      await dispatch(putIpBanApi({ ip, action: "ban", motif: "" }, user.token));
      refresh();
    } catch {
      /* toast shown in the thunk */
    } finally {
      setBusyIp(null);
    }
  };

  const doUnban = async (ip) => {
    setUnbanning(null);
    setBusyIp(ip);
    try {
      await dispatch(putIpBanApi({ ip, action: "unban" }, user.token));
      refresh();
    } catch {
      /* toast shown in the thunk */
    } finally {
      setBusyIp(null);
    }
  };

  const memberEmails = (row) => (row.members || []).map((m) => m.email).join(", ");

  return (
    <div className={styles.main}>
      <h1 className="cw-title">{t("admin.ipTitle")}</h1>
      <p className="cw-subtitle">{t("admin.ipIntro")}</p>

      <div className={styles.statsRow}>
        <div className={styles.statCard}>
          <span className={styles.statValue}>{stats.total}</span>
          <span className={styles.statLabel}>{t("admin.ipTotal")}</span>
        </div>
        <div className={styles.statCard}>
          <span className={styles.statValue}>{stats.banned}</span>
          <span className={styles.statLabel}>{t("admin.ipBannedCount")}</span>
        </div>
      </div>

      <div className={styles.toolbar}>
        <form className={styles.subjectForm} onSubmit={submitTerm}>
          <input
            className="cw-input"
            type="text"
            value={termInput}
            placeholder={t("admin.ipFilterPlaceholder")}
            onChange={(e) => setTermInput(e.target.value)}
          />
          <button type="submit" className="cw-button cw-button--sm" title={t("admin.ipFilter")}>
            <FaSearch style={{ fontSize: "14px" }} />
          </button>
        </form>
        <button
          type="button"
          className="cw-button cw-button--danger cw-button--sm"
          onClick={() => setBanning(true)}
          title={t("admin.banAddress")}
        >
          <FaBan style={{ fontSize: "13px" }} /> {t("admin.banAddress")}
        </button>
      </div>

      {data.results.length === 0 ? (
        <div className={styles.empty}>
          <strong>{t("admin.ipEmptyTitle")}</strong>
          <span>{t("admin.ipEmptyBody")}</span>
        </div>
      ) : (
        <ul className={`cw-list ${styles.ipList}`}>
          {data.results.map((row) => (
            <li className={`cw-list-item ${styles.ipRow}`} key={row.ip}>
              <code className={styles.ipAddr}>{row.ip}</code>
              <span className={styles.members} title={memberEmails(row)}>
                {row.members && row.members.length ? memberEmails(row) : t("admin.anonymous")}
              </span>
              <span className={styles.geo}>
                {geoLabel(row.geo) ? (
                  <>
                    <FaMapMarkerAlt style={{ fontSize: "11px", marginRight: 4 }} />
                    {geoLabel(row.geo)}
                  </>
                ) : (
                  "—"
                )}
              </span>
              <span className={styles.network}>{networkLabel(row.geo) || "—"}</span>
              <span className={styles.count}>{row.count}</span>
              <span className={styles.seen} title={row.lastSeen ? new Date(row.lastSeen).toISOString() : ""}>
                {row.lastSeen ? new Date(row.lastSeen).toLocaleString(locale) : "—"}
              </span>
              <span>
                {row.banned ? (
                  <span className={styles.bannedBadge} title={row.reason}>
                    {t("admin.banned")}
                  </span>
                ) : (
                  <span className={styles.muted}>—</span>
                )}
              </span>
              <div className={styles.rowActions}>
                <CopyIdButton id={row.ip} className={styles.iconBtn} />
                {row.banned ? (
                  <Tooltip label={t("admin.unban")}>
                    <button
                      type="button"
                      className={styles.iconBtn}
                      onClick={() => setUnbanning(row.ip)}
                      disabled={busyIp === row.ip}
                    >
                      <FaUnlock style={{ fontSize: "15px" }} />
                    </button>
                  </Tooltip>
                ) : (
                  <Tooltip label={t("admin.ban")}>
                    <button
                      type="button"
                      className={`${styles.iconBtn} ${styles.iconBtnDanger}`}
                      onClick={() => banRow(row.ip)}
                      disabled={busyIp === row.ip}
                    >
                      <FaBan style={{ fontSize: "15px" }} />
                    </button>
                  </Tooltip>
                )}
              </div>
            </li>
          ))}
        </ul>
      )}

      <div className={styles.pager}>
        <button
          type="button"
          className="cw-button cw-button--secondary cw-button--sm"
          onClick={() => setPage((p) => Math.max(0, p - 1))}
          disabled={page <= 0 || loading}
          title={t("audit.previousPage")}
        >
          <FaChevronLeft style={{ fontSize: "12px" }} /> {t("audit.previous")}
        </button>
        <span className={styles.pageLabel}>{t("audit.page", { page: page + 1 })}</span>
        <button
          type="button"
          className="cw-button cw-button--secondary cw-button--sm"
          onClick={() => setPage((p) => p + 1)}
          disabled={!data.hasNext || loading}
          title={t("audit.nextPage")}
        >
          {t("audit.next")} <FaChevronRight style={{ fontSize: "12px" }} />
        </button>
      </div>

      <Modal
        show={banning}
        title={t("admin.banAddress")}
        onClose={() => setBanning(false)}
        footer={
          <>
            <button type="button" className="cw-button cw-button--secondary" onClick={() => setBanning(false)}>
              {t("common.cancel")}
            </button>
            <button type="button" className="cw-button cw-button--danger" onClick={confirmBan}>
              {t("admin.ban")}
            </button>
          </>
        }
      >
        <form onSubmit={confirmBan}>
          <div className="cw-field">
            <label className="cw-label">{t("admin.ipAddress")}</label>
            <input
              className="cw-input"
              type="text"
              value={banIpValue}
              onChange={(e) => setBanIpValue(e.target.value)}
              placeholder="203.0.113.7"
              autoFocus
            />
          </div>
          <div className="cw-field">
            <label className="cw-label">{t("admin.banReason")}</label>
            <input
              className="cw-input"
              type="text"
              value={banMotif}
              onChange={(e) => setBanMotif(e.target.value)}
              placeholder={t("admin.banReasonPlaceholder")}
            />
          </div>
        </form>
      </Modal>

      <ConfirmModal
        show={!!unbanning}
        title={t("admin.unban")}
        body={unbanning ? t("admin.confirmUnban", { ip: unbanning }) : ""}
        confirmLabel={t("admin.unban")}
        onConfirm={() => doUnban(unbanning)}
        onCancel={() => setUnbanning(null)}
      />
    </div>
  );
};

export default IpManagement;
