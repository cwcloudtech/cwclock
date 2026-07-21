import React from "react";
import styles from "./Styles/Switch.module.css";

// Switch is a toggle-styled checkbox, used where an on/off action reads
// better as a switch than as an Enable/Disable button pair (eg. MfaSettings).
const Switch = ({ checked, onChange, disabled, id, ...props }) => (
  <span className={styles.track}>
    <input
      id={id}
      type="checkbox"
      className={styles.input}
      checked={checked}
      onChange={onChange}
      disabled={disabled}
      {...props}
    />
    <span className={styles.thumb} />
  </span>
);

export default Switch;
