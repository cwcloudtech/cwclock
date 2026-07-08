import React, { useState } from "react";
import Button from "./Button";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/CookieBanner.module.css";

// CWClock only ever stores functional data (auth token, theme, locale,
// selected org) in localStorage - no third-party tracking/analytics cookies
// exist to consent to, so this is an informational, one-time notice rather
// than an accept/reject cookie-consent flow.
const STORAGE_KEY = "cwclock_cookie_notice_ack";

const CookieBanner = () => {
  const { t } = useI18n();
  const [dismissed, setDismissed] = useState(() => localStorage.getItem(STORAGE_KEY) === "1");

  if (dismissed) return null;

  const handleAck = () => {
    localStorage.setItem(STORAGE_KEY, "1");
    setDismissed(true);
  };

  return (
    <div className={styles.banner} role="region" aria-label={t("cookieBanner.title")}>
      <p className={styles.text}>{t("cookieBanner.message")}</p>
      <Button size="sm" onClick={handleAck} title={t("cookieBanner.understand")}>
        {t("cookieBanner.understand")}
      </Button>
    </div>
  );
};

export default CookieBanner;
