import React from "react";
import Footer from "../Components/Login/Footer";
import LoginNav from "../Components/Login/LoginNav";
import ResetPasswordForm from "../Components/Login/ResetPasswordForm";
import styles from "./Styles/Login.module.css";

const ResetPassword = () => {
  return (
    <div className={styles.page}>
      <LoginNav />
      <div className={styles.main}>
        <ResetPasswordForm />
      </div>
      <Footer />
    </div>
  );
};

export default ResetPassword;
