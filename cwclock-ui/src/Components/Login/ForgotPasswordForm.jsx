import React, { useState } from "react";
import { Link } from "react-router-dom";
import styles from "./Styles/Form.module.css";
import { useDispatch } from "react-redux";
import Spinner from "../spinner/Spinner";
import { forgotPasswordApi } from "../../Redux/Auth/Auth.actions";
import { useI18n } from "../../i18n/I18nContext";

const ForgotPasswordForm = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const [email, setEmail] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    try {
      await dispatch(forgotPasswordApi(email));
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return <Spinner />;
  }

  return (
    <div className={styles.form}>
      <h1 className={styles.heading}>{t("auth.forgotPasswordTitle")}</h1>
      <p>{t("auth.forgotPasswordBody")}</p>
      <form onSubmit={handleSubmit}>
        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={(e) => setEmail(e.target.value)}
            name="email"
            value={email}
            type="email"
            placeholder={t("auth.enterEmail")}
            title={t("auth.emailAddress")}
          />
        </div>

        <button type="submit" className={styles.btn} title={t("auth.sendResetLink")}>
          {t("auth.sendResetLink")}
        </button>
      </form>
      <div className={styles.footer}>
        <Link to="/login">{t("auth.backToLogin")}</Link>
      </div>
    </div>
  );
};

export default ForgotPasswordForm;
