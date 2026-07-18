import React from "react";
import { useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";
import Button from "./Button";
import { logoutUser } from "../../Redux/Auth/Auth.actions";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/DisabledNotice.module.css";

// Blocks the whole app for a disabled or banned account: per spec they
// "can't do anything" until the superuser confirms them (or, for "email"
// activation mode, until they follow their confirmation link), so this
// replaces the dashboard entirely rather than just disabling a few
// actions. The message shown depends on why the account is blocked: banned
// outright, or disabled - in which case the server's i18nCode says whether
// activation mode is "admin" (needs a superuser) or "email" (needs the
// user to click their confirmation link).
const DisabledNotice = ({ user }) => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const navigate = useNavigate();

  const handleLogout = () => {
    dispatch(logoutUser());
    navigate("/login");
  };

  const isBanned = user?.role === "ban";
  const isEmailMode = user?.i18nCode === "errors.accountDisabledEmail";
  const title = isBanned ? t("banned.title") : t(isEmailMode ? "disabled.emailMode.title" : "disabled.adminMode.title");
  const body = isBanned ? t("banned.body") : t(isEmailMode ? "disabled.emailMode.body" : "disabled.adminMode.body");

  return (
    <div className={styles.page}>
      <div className={styles.card}>
        <h1 className={styles.title}>{title}</h1>
        <p className={styles.body}>{body}</p>
        <Button variant="secondary" onClick={handleLogout} title={t("disabled.signOut")}>
          {t("nav.logout")}
        </Button>
      </div>
    </div>
  );
};

export default DisabledNotice;
