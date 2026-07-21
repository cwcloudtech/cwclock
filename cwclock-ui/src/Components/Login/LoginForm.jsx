import React, { useEffect, useState } from "react";
import styles from "./Styles/Form.module.css";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import { toast } from "react-toastify";
import Spinner from "../spinner/Spinner";
import { loginApi } from "../../Redux/Auth/Auth.actions";
import { useI18n } from "../../i18n/I18nContext";
import OidcButtons from "../common/OidcButtons";
import toastOptions from "../../Redux/toastOptions";
import MfaChallengeForm from "./MfaChallengeForm";

const LoginForm = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { user, isLoading, mfaChallenge } = useSelector((state) => state.auth);
  const [searchParams] = useSearchParams();
  const [formData, setFormData] = useState({
    email: "",
    password: "",
  });
  const { email, password } = formData;

  useEffect(() => {
    const confirmed = searchParams.get("confirmed");
    if (confirmed === "1") {
      toast.success(t("auth.confirmationSuccess"), toastOptions);
    } else if (confirmed === "0") {
      const reason = searchParams.get("reason");
      toast.error(reason === "banned" ? t("auth.confirmationBanned") : t("auth.confirmationInvalid"), toastOptions);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const onChange = (e) => {
    let { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value,
    });
  };
  const handleSubmit = (e) => {
    e.preventDefault();
    const userData = {
      email,
      password,
    };
    dispatch(loginApi(userData));
  };
  useEffect(() => {
    if (user.token) {
      navigate("/dashboard/timetracker");
    }
  }, [user, navigate]);

  if (isLoading) {
    return <Spinner />;
  }
  if (mfaChallenge) {
    return <MfaChallengeForm challenge={mfaChallenge} />;
  }
  return (
    <div className={styles.form}>
      <h1 className={styles.heading}>{t("auth.logIn")}</h1>
      <form onSubmit={handleSubmit}>
        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={onChange}
            name="email"
            value={email}
            type="email"
            placeholder={t("auth.enterEmail")}
            title={t("auth.emailAddress")}
          />
        </div>

        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={onChange}
            name="password"
            value={password}
            type="password"
            placeholder={t("auth.password")}
            title={t("auth.password")}
          />
        </div>

        <button type="submit" className={styles.btn} title={t("auth.logInToAccount")}>
          {t("auth.logIn")}
        </button>
      </form>
      <div className={styles.footer}>
        <Link to="/forgot-password">{t("auth.forgotPassword")}</Link>
      </div>
      <OidcButtons />
    </div>
  );
};

export default LoginForm;
