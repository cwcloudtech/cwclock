import React from "react";
import { Link } from "react-router-dom";
import logo from "../../assets/images/cwclock-logo.svg";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/AuthNav.module.css";

const AuthNav = ({ prompt, linkTo, linkLabel }) => {
  const { t } = useI18n();
  return (
  <div className={styles.nav}>
    <Link to="/login" className={styles.logo} title={t("nav.cwclockHome")}>
      <img src={logo} alt={t("nav.cwclock")} />
    </Link>
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
