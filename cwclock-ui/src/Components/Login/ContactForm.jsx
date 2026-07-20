import React, { useState } from "react";
import { Link } from "react-router-dom";
import styles from "./Styles/Form.module.css";
import { useDispatch, useSelector } from "react-redux";
import Spinner from "../spinner/Spinner";
import { sendContactApi } from "../../Redux/Contact/Contact.actions";
import { useI18n } from "../../i18n/I18nContext";

const emptyFields = {
  firstname: "",
  name: "",
  email: "",
  subject: "",
  message: "",
};

// initialFields presets the email field with the connected user's own
// address (ai-instruct-57), so a logged-in user reaching this form from the
// navbar doesn't have to retype an email the app already knows.
const initialFields = (user) => ({
  ...emptyFields,
  email: user?.token ? user.email : "",
});

const ContactForm = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const [fields, setFields] = useState(() => initialFields(user));
  const [isLoading, setIsLoading] = useState(false);

  const setField = (key, value) => setFields((f) => ({ ...f, [key]: value }));

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    try {
      await dispatch(sendContactApi(fields));
      setFields(initialFields(user));
    } catch {
      // Error toast already shown by sendContactApi.
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return <Spinner />;
  }

  return (
    <div className={styles.form}>
      <h1 className={styles.heading}>{t("contact.title")}</h1>
      <p>{t("contact.body")}</p>
      <form onSubmit={handleSubmit}>
        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={(e) => setField("firstname", e.target.value)}
            name="firstname"
            value={fields.firstname}
            type="text"
            placeholder={t("common.firstName")}
            title={t("common.firstName")}
          />
        </div>

        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={(e) => setField("name", e.target.value)}
            name="name"
            value={fields.name}
            type="text"
            placeholder={t("common.lastName")}
            title={t("common.lastName")}
          />
        </div>

        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={(e) => setField("email", e.target.value)}
            name="email"
            value={fields.email}
            type="email"
            placeholder={t("auth.enterEmail")}
            title={t("auth.emailAddress")}
            required
          />
        </div>

        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={(e) => setField("subject", e.target.value)}
            name="subject"
            value={fields.subject}
            type="text"
            placeholder={t("contact.subject")}
            title={t("contact.subject")}
            required
          />
        </div>

        <div className={styles.field}>
          <textarea
            className={styles.input}
            style={{ minHeight: 120, resize: "vertical", fontFamily: "inherit" }}
            onChange={(e) => setField("message", e.target.value)}
            name="message"
            value={fields.message}
            rows={5}
            placeholder={t("contact.message")}
            title={t("contact.message")}
            required
          />
        </div>

        <button type="submit" className={styles.btn} title={t("contact.send")}>
          {t("contact.send")}
        </button>
      </form>
      <div className={styles.footer}>
        {user?.token ? (
          <Link to="/dashboard/timetracker">{t("contact.backToTimeTracking")}</Link>
        ) : (
          <Link to="/login">{t("auth.backToLogin")}</Link>
        )}
      </div>
    </div>
  );
};

export default ContactForm;
