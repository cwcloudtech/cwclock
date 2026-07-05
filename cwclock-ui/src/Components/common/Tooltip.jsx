import React from "react";
import styles from "./Styles/Tooltip.module.css";

// CSS-driven tooltip: shown on hover/focus via :hover/:focus-within, so it
// works for both mouse and keyboard navigation without any JS positioning
// logic. Replaces bare `title` attributes, whose native browser tooltip is
// slow to appear and inconsistently styled (this was especially noticeable
// on the icon-only sidebar links).
const Tooltip = ({ label, position = "top", className = "", children }) => {
  if (!label) return children;

  return (
    <span className={`${styles.wrapper} ${className}`}>
      {children}
      <span className={styles.bubble} data-position={position} role="tooltip">
        {label}
      </span>
    </span>
  );
};

export default Tooltip;
