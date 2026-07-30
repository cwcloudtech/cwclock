import React from "react";
import { Link } from "react-router-dom";
import { FaAndroid } from "react-icons/fa";
import logo from "../../assets/images/cwclock-logo.svg";
import { useI18n } from "../../i18n/I18nContext";
import useAppVersion from "./useAppVersion";
import isAndroidDevice from "./isAndroidDevice";
import styles from "./Styles/AuthNav.module.css";

// Shown on both the login and signup forms (ai-instruct-106), same
// Android-only gating as the navbar's download link (Slidebar.jsx).
const AuthNav = ({ prompt, linkTo, linkLabel }) => {
  const { t } = useI18n();
  const appVersion = useAppVersion();

  return (
  <div className={styles.nav}>
    <Link to="/login" className={styles.logo} title={t("nav.cwclockHome")}>
      <img src={logo} alt={t("nav.cwclock")} />
    </Link>
    {appVersion && isAndroidDevice() && (
      <a
        href={process.env.REACT_APP_MOBILE_URL_PATTERN.replace("{version}", appVersion)}
        className={styles.androidLink}
        title={t("nav.downloadAndroidApp")}
      >
        <FaAndroid fontSize="18px" />
        <span>{t("nav.downloadAndroidApp")}</span>
      </a>
    )}
    <div className={styles.actions}>
      <span className={styles.prompt}>{prompt}</span>
      <Link to={linkTo} className={styles.link} title={linkLabel}>
        {linkLabel}
      </Link>
    </div>
  </div>
  );
};

export default AuthNav;
