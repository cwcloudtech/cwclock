import React, { useEffect, useState } from "react";
import styles from "./Styles/Form.module.css";
import { useNavigate } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import Spinner from "../spinner/Spinner";
import { loginApi } from "../../Redux/Auth/Auth.actions";
import { useI18n } from "../../i18n/I18nContext";

const LoginForm = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { user, isLoading } = useSelector((state) => state.auth);
  const [formData, setFormData] = useState({
    email: "",
    password: "",
  });
  const { email, password } = formData;

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
    </div>
  );
};

export default LoginForm;
