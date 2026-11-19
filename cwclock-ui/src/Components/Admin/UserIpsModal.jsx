import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import Modal from "../common/Modal";
import Button from "../common/Button";
import Spinner from "../spinner/Spinner";
import EmptyState from "../common/EmptyState";
import memberLabel from "../common/memberLabel";
import { listUserIpsApi, putIpBanApi } from "../../Redux/Admin/Admin.actions";
import { useI18n } from "../../i18n/I18nContext";
import { geoLabel } from "./geoLabel";
import styles from "./Styles/Ips.module.css";

// UserIpsModal shows every address an account has connected from (the "eye"
// button on each row of the admin Users screen), as a clean table with its
// geolocation, connection count and first/last seen, and lets a superuser
// ban/unban the address inline. The ban is the SAME global IP ban the IP
// management screen manages - banning here blocks the address for everyone.
const UserIpsModal = ({ show, targetUser, token, onClose }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const [ips, setIps] = useState(null); // null = loading
  const [busyIp, setBusyIp] = useState(null);

  useEffect(() => {
    if (!show || !targetUser) return;
    let cancelled = false;
    setIps(null);
    dispatch(listUserIpsApi(targetUser.id, token))
      .then((data) => !cancelled && setIps(Array.isArray(data) ? data : []))
      .catch(() => !cancelled && setIps([]));
    return () => {
      cancelled = true;
    };
  }, [show, targetUser, token, dispatch]);

  const toggleBan = async (entry) => {
    const action = entry.banned ? "unban" : "ban";
    setBusyIp(entry.ip);
    // Optimistic flip, rolled back if the call fails.
    setIps((list) => list.map((x) => (x.ip === entry.ip ? { ...x, banned: !x.banned } : x)));
    try {
      await dispatch(putIpBanApi({ ip: entry.ip, action, motif: "" }, token));
    } catch {
      setIps((list) => list.map((x) => (x.ip === entry.ip ? { ...x, banned: entry.banned } : x)));
    } finally {
      setBusyIp(null);
    }
  };

  if (!targetUser) return null;

  const fmt = (d) => (d ? new Date(d).toLocaleString(locale) : "—");

  return (
    <Modal show={show} title={t("admin.userIpsTitle", { name: memberLabel(targetUser) })} onClose={onClose} wide>
      <p className={styles.modalHelp}>{t("admin.userIpsHelp")}</p>

      {ips === null ? (
        <Spinner />
      ) : ips.length === 0 ? (
        <EmptyState title={t("admin.noIps")} body={t("admin.noIpsBody")} />
      ) : (
        <div className={styles.tableWrap}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>{t("admin.ipAddress")}</th>
                <th>{t("admin.ipLocation")}</th>
                <th className={styles.num}>{t("admin.ipConnections")}</th>
                <th>{t("admin.ipFirstSeen")}</th>
                <th>{t("admin.ipLastSeen")}</th>
                <th>{t("admin.ipStatus")}</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {ips.map((entry) => (
                <tr key={entry.ip}>
                  <td>
                    <code>{entry.ip}</code>
                  </td>
                  <td className={styles.muted}>{geoLabel(entry.geo) || "—"}</td>
                  <td className={styles.num}>{entry.count}</td>
                  <td className={styles.muted}>{fmt(entry.firstSeen)}</td>
                  <td className={styles.muted}>{fmt(entry.lastSeen)}</td>
                  <td>
                    {entry.banned ? (
                      <span className={styles.bannedBadge}>{t("admin.banned")}</span>
                    ) : (
                      <span className={styles.muted}>—</span>
                    )}
                  </td>
                  <td className={styles.num}>
                    <Button
                      size="sm"
                      variant={entry.banned ? "secondary" : "danger"}
                      onClick={() => toggleBan(entry)}
                      disabled={busyIp === entry.ip}
                    >
                      {entry.banned ? t("admin.unban") : t("admin.ban")}
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Modal>
  );
};

export default UserIpsModal;
