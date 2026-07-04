import React, { useEffect, useRef, useState } from "react";
import styles from "./Styles/Dropdown.module.css";

// Minimal replacement for react-bootstrap's Dropdown: a trigger button that
// opens an absolutely-positioned menu, closed on outside click or Escape.
const Dropdown = ({ trigger, align = "start", className = "", triggerClassName = "", title, children }) => {
  const [open, setOpen] = useState(false);
  const ref = useRef(null);

  useEffect(() => {
    if (!open) return;
    const onClick = (e) => {
      if (ref.current && !ref.current.contains(e.target)) setOpen(false);
    };
    const onKeyDown = (e) => e.key === "Escape" && setOpen(false);
    document.addEventListener("mousedown", onClick);
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("mousedown", onClick);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open]);

  return (
    <div className={`${styles.dropdown} ${className}`} ref={ref}>
      <button
        type="button"
        className={`${styles.trigger} ${triggerClassName}`}
        onClick={() => setOpen(!open)}
        title={title}
      >
        {trigger}
      </button>
      {open && (
        <div className={`${styles.menu} ${align === "end" ? styles.alignEnd : ""}`}>
          {typeof children === "function" ? children(() => setOpen(false)) : children}
        </div>
      )}
    </div>
  );
};

export const DropdownItem = ({ active, className = "", ...props }) => (
  <button
    type="button"
    className={`${styles.item} ${active ? styles.itemActive : ""} ${className}`}
    {...props}
  />
);

export const DropdownText = ({ className = "", ...props }) => (
  <div className={`${styles.text} ${className}`} {...props} />
);

export const DropdownDivider = () => <div className={styles.divider} />;

export default Dropdown;
