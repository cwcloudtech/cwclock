import React, { useEffect, useState } from "react";
import styles from "../Login/Styles/Form.module.css";
import { useDispatch, useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import Spinner from "../spinner/Spinner";
import { registerApi } from "../../Redux/Auth/Auth.actions";
import { useI18n } from "../../i18n/I18nContext";

const SignUpForm = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { user, isLoading } = useSelector((state) => state.auth);
  const [formData, setFormData] = useState({
    name: "",
    surname: "",
    email: "",
    password: "",
    confirmPassword: "",
  });
  const [error, setError] = useState("");

  const { name, surname, email, password, confirmPassword } = formData;

  const onChange = (e) => {
    let { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value,
    });
  };
  const handleSubmit = async (e) => {
    e.preventDefault();
    if (password !== confirmPassword) {
      setError(t("auth.passwordsDontMatch"));
      return;
    }
    setError("");
    const userData = {
      name,
      surname,
      email,
      password,
    };
    dispatch(registerApi(userData));
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
      <h1 className={styles.heading}>{t("auth.signUp")}</h1>
      <form onSubmit={handleSubmit}>
        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={onChange}
            name="name"
            value={name}
            type="text"
            placeholder={t("common.firstName")}
            title={t("common.firstName")}
          />
        </div>

        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={onChange}
            name="surname"
            value={surname}
            type="text"
            placeholder={t("common.lastName")}
            title={t("common.lastName")}
          />
        </div>

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

        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={onChange}
            name="confirmPassword"
            value={confirmPassword}
            type="password"
            placeholder={t("common.confirmPassword")}
            title={t("common.confirmPassword")}
          />
        </div>

        {error && <p className="cw-error">{error}</p>}

        <button type="submit" className={styles.btn} title={t("auth.createAccountTitle")}>
          {t("auth.createAccount")}
        </button>
      </form>
    </div>
  );
};

export default SignUpForm;
