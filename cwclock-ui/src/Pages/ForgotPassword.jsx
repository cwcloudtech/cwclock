import React from "react";
import Footer from "../Components/Login/Footer";
import LoginNav from "../Components/Login/LoginNav";
import ForgotPasswordForm from "../Components/Login/ForgotPasswordForm";
import styles from "./Styles/Login.module.css";

const ForgotPassword = () => {
  return (
    <div className={styles.page}>
      <LoginNav />
      <div className={styles.main}>
        <ForgotPasswordForm />
      </div>
      <Footer />
    </div>
  );
};

export default ForgotPassword;
