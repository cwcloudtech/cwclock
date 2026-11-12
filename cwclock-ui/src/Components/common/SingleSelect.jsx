import React from "react";
import Dropdown, { DropdownItem } from "./Dropdown";
import styles from "./Styles/SingleSelect.module.css";

// Single-choice filter dropdown, the one-value counterpart to MultiSelect:
// same trigger (muted label + bold summary) and the same Dropdown web
// component underneath, so a filter toolbar mixing the two reads as one
// control. A native <select> would not - it is styled by the OS, not by the
// app's theme.
//
// allLabel is the entry that clears the filter; it's always first, and
// selecting anything closes the panel (unlike MultiSelect, where the point
// is to tick several in a row).
const SingleSelect = ({ label, options, value, onChange, allLabel }) => {
  const summary = options.find((o) => o.value === value)?.label || allLabel;

  return (
    <Dropdown
      title={label}
      triggerClassName={styles.trigger}
      trigger={
        <>
          <span className={styles.triggerLabel}>{label}</span>
          <span className={styles.triggerSummary}>{summary}</span>
        </>
      }
    >
      {(close) => (
        <div className={styles.panel}>
          <div className={styles.list}>
            <DropdownItem
              active={!value}
              onClick={() => {
                onChange("");
                close();
              }}
            >
              {allLabel}
            </DropdownItem>
            {options.map((o) => (
              <DropdownItem
                key={o.value}
                active={o.value === value}
                onClick={() => {
                  onChange(o.value);
                  close();
                }}
              >
                {o.label}
              </DropdownItem>
            ))}
          </div>
        </div>
      )}
    </Dropdown>
  );
};

export default SingleSelect;
