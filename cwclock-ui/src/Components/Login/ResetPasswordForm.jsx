import React, { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import styles from "./Styles/Form.module.css";
import { useDispatch } from "react-redux";
import { toast } from "react-toastify";
import Spinner from "../spinner/Spinner";
import { resetPasswordApi } from "../../Redux/Auth/Auth.actions";
import { useI18n } from "../../i18n/I18nContext";
import toastOptions from "../../Redux/toastOptions";

const ResetPasswordForm = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token") || "";
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!token) {
      toast.error(t("auth.missingResetToken"), toastOptions);
      return;
    }
    if (password !== confirmPassword) {
      toast.error(t("auth.passwordsDontMatch"), toastOptions);
      return;
    }
    setIsLoading(true);
    try {
      await dispatch(resetPasswordApi(token, password, confirmPassword));
      navigate("/login");
    } catch (e) {
      // apiErrorMessage already toasted by resetPasswordApi.
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return <Spinner />;
  }

  return (
    <div className={styles.form}>
      <h1 className={styles.heading}>{t("auth.resetPasswordTitle")}</h1>
      <form onSubmit={handleSubmit}>
        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={(e) => setPassword(e.target.value)}
            name="password"
            value={password}
            type="password"
            placeholder={t("auth.newPassword")}
            title={t("auth.newPassword")}
          />
        </div>

        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={(e) => setConfirmPassword(e.target.value)}
            name="confirmPassword"
            value={confirmPassword}
            type="password"
            placeholder={t("auth.confirmNewPassword")}
            title={t("auth.confirmNewPassword")}
          />
        </div>

        <button type="submit" className={styles.btn} title={t("auth.resetPassword")}>
          {t("auth.resetPassword")}
        </button>
      </form>
      <div className={styles.footer}>
        <Link to="/login">{t("auth.backToLogin")}</Link>
      </div>
    </div>
  );
};

export default ResetPasswordForm;
