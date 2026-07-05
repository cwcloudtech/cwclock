import React from "react";
import { useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";
import Button from "./Button";
import { logoutUser } from "../../Redux/Auth/Auth.actions";
import styles from "./Styles/DisabledNotice.module.css";

// Blocks the whole app for a disabled account: per spec they "can't do
// anything" until the superuser confirms them, so this replaces the
// dashboard entirely rather than just disabling a few actions.
const DisabledNotice = () => {
  const dispatch = useDispatch();
  const navigate = useNavigate();

  const handleLogout = () => {
    dispatch(logoutUser());
    navigate("/login");
  };

  return (
    <div className={styles.page}>
      <div className={styles.card}>
        <h1 className={styles.title}>Account disabled</h1>
        <p className={styles.body}>
          Your account is disabled. Please contact an administrator to get access.
        </p>
        <Button variant="secondary" onClick={handleLogout} title="Sign out">
          Logout
        </Button>
      </div>
    </div>
  );
};

export default DisabledNotice;
