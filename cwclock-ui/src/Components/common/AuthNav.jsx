import React, { useState } from "react";
import { Link } from "react-router-dom";
import { FaAndroid } from "react-icons/fa";
import logo from "../../assets/images/cwclock-logo.svg";
import { useI18n } from "../../i18n/I18nContext";
import useAppVersion from "./useAppVersion";
import isAndroidDevice from "./isAndroidDevice";
import AndroidQrModal from "./AndroidQrModal";
import styles from "./Styles/AuthNav.module.css";

// Shown on both the login and signup forms (ai-instruct-106), same
// Android-gating as the navbar's download link (Slidebar.jsx).
const AuthNav = ({ prompt, linkTo, linkLabel }) => {
  const { t } = useI18n();
  const appVersion = useAppVersion();
  const [showQr, setShowQr] = useState(false);
  const downloadUrl = appVersion
    ? process.env.REACT_APP_MOBILE_URL_PATTERN.replace("{version}", appVersion)
    : null;

  return (
  <div className={styles.nav}>
    <Link to="/login" className={styles.logo} title={t("nav.cwclockHome")}>
      <img src={logo} alt={t("nav.cwclock")} />
    </Link>
    {downloadUrl && (
      isAndroidDevice() ? (
        <a
          href={downloadUrl}
          className={styles.androidLink}
          title={t("nav.downloadAndroidApp")}
        >
          <FaAndroid fontSize="18px" />
          <span>{t("nav.downloadAndroidApp")}</span>
        </a>
      ) : (
        <button
          type="button"
          className={`${styles.androidLink} ${styles.androidButton}`}
          title={t("nav.downloadAndroidApp")}
          onClick={() => setShowQr(true)}
        >
          <FaAndroid fontSize="18px" />
          <span>{t("nav.downloadAndroidApp")}</span>
        </button>
      )
    )}
    <div className={styles.actions}>
      <span className={styles.prompt}>{prompt}</span>
      <Link to={linkTo} className={styles.link} title={linkLabel}>
        {linkLabel}
      </Link>
    </div>
    <AndroidQrModal show={showQr} onClose={() => setShowQr(false)} url={downloadUrl} />
  </div>
  );
};

export default AuthNav;
