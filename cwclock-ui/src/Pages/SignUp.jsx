import React from "react";
import Heading from "../Components/SignUp/Heading";
import SignUpForm from "../Components/SignUp/SignUpForm";
import SignUpNav from "../Components/SignUp/SignUpNav";
import Footer from "../Components/Login/Footer";
import styles from "./Styles/Login.module.css";

const SignUp = () => {
  return (
    <div className={styles.page}>
      <SignUpNav />
      <div className={styles.main}>
        <div>
          <Heading />
          <SignUpForm />
        </div>
      </div>
      <Footer />
    </div>
  );
};

export default SignUp;
