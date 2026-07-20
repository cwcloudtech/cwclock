import React from "react";
import Footer from "../Components/Login/Footer";
import LoginNav from "../Components/Login/LoginNav";
import ContactForm from "../Components/Login/ContactForm";
import styles from "./Styles/Login.module.css";

const Contact = () => {
  return (
    <div className={styles.page}>
      <LoginNav />
      <div className={styles.main}>
        <ContactForm />
      </div>
      <Footer />
    </div>
  );
};

export default Contact;
