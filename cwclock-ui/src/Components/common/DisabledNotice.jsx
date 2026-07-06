import React from "react";
import { useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";
import Button from "./Button";
import { logoutUser } from "../../Redux/Auth/Auth.actions";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/DisabledNotice.module.css";

// Blocks the whole app for a disabled account: per spec they "can't do
// anything" until the superuser confirms them, so this replaces the
// dashboard entirely rather than just disabling a few actions.
const DisabledNotice = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const navigate = useNavigate();

  const handleLogout = () => {
    dispatch(logoutUser());
    navigate("/login");
  };

  return (
    <div className={styles.page}>
      <div className={styles.card}>
        <h1 className={styles.title}>{t("disabled.title")}</h1>
        <p className={styles.body}>{t("disabled.body")}</p>
        <Button variant="secondary" onClick={handleLogout} title={t("disabled.signOut")}>
          {t("nav.logout")}
        </Button>
      </div>
    </div>
  );
};

export default DisabledNotice;
