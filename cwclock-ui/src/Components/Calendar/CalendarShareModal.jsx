import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { toast } from "react-toastify";
import { FaRegCopy } from "react-icons/fa";
import Modal from "../common/Modal";
import Button from "../common/Button";
import Switch from "../common/Switch";
import Tooltip from "../common/Tooltip";
import {
  calendarFeedStatusApi,
  calendarFeedEnableApi,
  calendarFeedDisableApi,
  calendarFeedRegenerateApi,
} from "../../Redux/Users/User.actions";
import toastOptions from "../../Redux/toastOptions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import styles from "./Styles/CalendarShareModal.module.css";

// CalendarShareModal is the Calendar view's "external connections" panel
// (ai-instruct-85): a subscribable ICS feed URL a user can add to Outlook or
// Google Calendar, gated by an enable/disable switch and backed by a token
// stored on the user's own data payload rather than a per-organization
// connection (see internal/handlers/calendar_feed_handler.go).
const CalendarShareModal = ({ show, onClose }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const [status, setStatus] = useState(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const refresh = async () => {
    try {
      setStatus(await dispatch(calendarFeedStatusApi(user.token)));
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    }
  };

  useEffect(() => {
    if (show) {
      setError("");
      refresh();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [show]);

  const handleToggle = async (e) => {
    const enabled = e.target.checked;
    setError("");
    setBusy(true);
    try {
      setStatus(await dispatch(enabled ? calendarFeedEnableApi(user.token) : calendarFeedDisableApi(user.token)));
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    } finally {
      setBusy(false);
    }
  };

  const handleRegenerate = async () => {
    setError("");
    setBusy(true);
    try {
      setStatus(await dispatch(calendarFeedRegenerateApi(user.token)));
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    } finally {
      setBusy(false);
    }
  };

  const handleCopy = () => {
    if (!status?.url) return;
    navigator.clipboard.writeText(status.url).then(() => {
      toast.success(t("common.copied"), toastOptions);
    });
  };

  return (
    <Modal show={show} title={t("calendar.shareTitle")} onClose={onClose}>
      <p className={styles.intro}>{t("calendar.shareIntro")}</p>

      <div className={styles.switchField}>
        <Switch
          checked={!!status?.enabled}
          onChange={handleToggle}
          disabled={busy || !status}
          aria-label={t("calendar.shareEnable")}
        />
        <span className="cw-label">{t("calendar.shareEnable")}</span>
      </div>

      {status?.enabled && status?.url && (
        <>
          <div className={styles.urlBox}>
            <code className={styles.url}>{status.url}</code>
            <Tooltip label={t("common.copy")}>
              <button type="button" className={styles.iconBtn} onClick={handleCopy}>
                <FaRegCopy style={{ fontSize: "16px" }} />
              </button>
            </Tooltip>
          </div>
          <Button variant="secondary" size="sm" onClick={handleRegenerate} disabled={busy} title={t("calendar.shareRegenerate")}>
            {t("calendar.shareRegenerate")}
          </Button>
          <p className={styles.hint}>{t("calendar.shareHint")}</p>
        </>
      )}

      {error && <p className="cw-error">{error}</p>}
    </Modal>
  );
};

export default CalendarShareModal;
