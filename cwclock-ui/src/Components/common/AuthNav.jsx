import React from "react";
import { Link } from "react-router-dom";
import logo from "../../assets/images/cwclock-logo.svg";
import styles from "./Styles/AuthNav.module.css";

const AuthNav = ({ prompt, linkTo, linkLabel }) => (
  <div className={styles.nav}>
    <Link to="/login" className={styles.logo} title="CWClock home">
      <img src={logo} alt="CWClock" />
    </Link>
    <div className={styles.actions}>
      <span className={styles.prompt}>{prompt}</span>
      <Link to={linkTo} className={styles.link} title={linkLabel}>
        {linkLabel}
      </Link>
    </div>
  </div>
);

export default AuthNav;
