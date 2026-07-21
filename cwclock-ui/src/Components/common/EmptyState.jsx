import React from "react";
import styles from "./Styles/EmptyState.module.css";
import logo from "../../assets/images/octopus.png";

const EmptyState = ({ title, body, action }) => (
  <div className={styles.Empty}>
    <img src={logo} alt="" />
    {title && <h4>{title}</h4>}
    {body && <p>{body}</p>}
    {action && <div className={styles.action}>{action}</div>}
  </div>
);

export default EmptyState;
