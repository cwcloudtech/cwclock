import React from "react";
import Footer from "../Components/Login/Footer";
import LoginForm from "../Components/Login/LoginForm";
import LoginNav from "../Components/Login/LoginNav";
import styles from "./Styles/Login.module.css";

const Login = () => {
  return (
    <div className={styles.page}>
      <LoginNav />
      <LoginForm label="Log In" checkBox="Stay Signed In" />
      <Footer />
    </div>
  );
};

export default Login;
