import React from "react";
import styles from "./Styles/Tooltip.module.css";

// CSS-driven tooltip: shown on hover via :hover, and on keyboard focus via
// :has(:focus-visible) - deliberately not :focus-within, which also matches
// the lingering focus a mouse click leaves on a button, keeping the tooltip
// stuck open after the pointer has moved away. Replaces bare `title`
// attributes, whose native browser tooltip is slow to appear and
// inconsistently styled (this was especially noticeable on the icon-only
// sidebar links).
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
