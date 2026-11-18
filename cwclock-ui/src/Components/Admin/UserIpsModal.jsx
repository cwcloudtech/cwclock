import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { FaBan, FaUnlock } from "react-icons/fa";
import Modal from "../common/Modal";
import Tooltip from "../common/Tooltip";
import memberLabel from "../common/memberLabel";
import { listUserIpsApi, putIpBanApi } from "../../Redux/Admin/Admin.actions";
import { useI18n } from "../../i18n/I18nContext";
import { geoLabel } from "./geoLabel";
import styles from "./Styles/Ips.module.css";

// UserIpsModal shows every address an account has connected from (the "eye"
// button on each row of the admin Users screen), with its geolocation,
// connection count and last-seen, and lets a superuser ban/unban the address
// inline. The ban is the SAME global IP ban the IP-management screen manages -
// banning here blocks the address for everyone, not just this user.
const UserIpsModal = ({ show, targetUser, token, onClose }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const [ips, setIps] = useState([]);
  const [loading, setLoading] = useState(false);
  const [busyIp, setBusyIp] = useState(null);

  useEffect(() => {
    if (!show || !targetUser) return;
    let cancelled = false;
    setLoading(true);
    dispatch(listUserIpsApi(targetUser.id, token))
      .then((data) => !cancelled && setIps(Array.isArray(data) ? data : []))
      .catch(() => !cancelled && setIps([]))
      .finally(() => !cancelled && setLoading(false));
    return () => {
      cancelled = true;
    };
  }, [show, targetUser, token, dispatch]);

  const toggleBan = async (entry) => {
    const action = entry.banned ? "unban" : "ban";
    setBusyIp(entry.ip);
    // Optimistic flip, rolled back if the call fails.
    setIps((prev) => prev.map((e) => (e.ip === entry.ip ? { ...e, banned: !e.banned } : e)));
    try {
      await dispatch(putIpBanApi({ ip: entry.ip, action, motif: "" }, token));
    } catch {
      setIps((prev) => prev.map((e) => (e.ip === entry.ip ? { ...e, banned: entry.banned } : e)));
    } finally {
      setBusyIp(null);
    }
  };

  if (!targetUser) return null;

  return (
    <Modal show={show} title={t("admin.userIpsTitle", { name: memberLabel(targetUser) })} onClose={onClose}>
      {loading ? (
        <p className="cw-subtitle">{t("common.loading")}</p>
      ) : ips.length === 0 ? (
        <p className="cw-subtitle">{t("admin.noIps")}</p>
      ) : (
        <ul className={`cw-list ${styles.ipList}`}>
          {ips.map((entry) => (
            <li className={`cw-list-item ${styles.userIpRow}`} key={entry.ip}>
              <code className={styles.ipAddr}>{entry.ip}</code>
              <span className={styles.geo}>{geoLabel(entry.geo) || "—"}</span>
              <span className={styles.count}>{entry.count}</span>
              <span className={styles.seen} title={entry.lastSeen ? new Date(entry.lastSeen).toISOString() : ""}>
                {entry.lastSeen ? new Date(entry.lastSeen).toLocaleString(locale) : "—"}
              </span>
              <span>
                {entry.banned ? (
                  <span className={styles.bannedBadge}>{t("admin.banned")}</span>
                ) : (
                  <span className={styles.muted}>—</span>
                )}
              </span>
              <div className={styles.rowActions}>
                <Tooltip label={entry.banned ? t("admin.unban") : t("admin.ban")}>
                  <button
                    type="button"
                    className={`${styles.iconBtn} ${entry.banned ? "" : styles.iconBtnDanger}`}
                    onClick={() => toggleBan(entry)}
                    disabled={busyIp === entry.ip}
                  >
                    {entry.banned ? (
                      <FaUnlock style={{ fontSize: "15px" }} />
                    ) : (
                      <FaBan style={{ fontSize: "15px" }} />
                    )}
                  </button>
                </Tooltip>
              </div>
            </li>
          ))}
        </ul>
      )}
    </Modal>
  );
};

export default UserIpsModal;
