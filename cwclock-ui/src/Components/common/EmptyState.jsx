import React from "react";
import styles from "./Styles/EmptyState.module.css";
import logo from "../../assets/images/octopus.png";

const EmptyState = ({ title, body }) => (
  <div className={styles.Empty}>
    <img src={logo} alt="" />
    <h4>{title}</h4>
    <p>{body}</p>
  </div>
);

export default EmptyState;
