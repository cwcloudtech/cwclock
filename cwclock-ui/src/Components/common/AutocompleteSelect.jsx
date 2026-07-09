import React, { useMemo, useState } from "react";
import Dropdown, { DropdownItem, DropdownText } from "./Dropdown";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/MultiSelect.module.css";

// Single-select autocomplete dropdown, styled like MultiSelect (search box
// narrows the option list) but picking an option selects it and closes the
// panel right away, instead of staying open for picking several.
const AutocompleteSelect = ({ label, placeholder, options, value, onChange, disabled, className = "" }) => {
  const { t } = useI18n();
  const [query, setQuery] = useState("");

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return options;
    return options.filter((o) => o.label.toLowerCase().includes(q));
  }, [options, query]);

  const selectedLabel = options.find((o) => o.value === value)?.label;

  return (
    <Dropdown
      title={label}
      className={className}
      triggerClassName={styles.trigger}
      disabled={disabled}
      trigger={
        <>
          <span className={styles.triggerLabel}>{label}</span>
          <span className={styles.triggerSummary}>{selectedLabel || placeholder}</span>
        </>
      }
    >
      {(close) => (
        <div className={styles.panel}>
          <input
            className={`cw-input ${styles.search}`}
            type="text"
            placeholder={t("common.search")}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            autoFocus
          />
          <div className={styles.list}>
            {filtered.length === 0 && <DropdownText>{t("common.noResults")}</DropdownText>}
            {filtered.map((o) => (
              <DropdownItem
                key={o.value}
                active={o.value === value}
                onClick={() => {
                  onChange(o.value);
                  setQuery("");
                  close();
                }}
              >
                {o.label}
              </DropdownItem>
            ))}
          </div>
          {value && (
            <button
              type="button"
              className={styles.clear}
              onClick={() => {
                onChange("");
                setQuery("");
                close();
              }}
            >
              {t("common.clearFilter")}
            </button>
          )}
        </div>
      )}
    </Dropdown>
  );
};

export default AutocompleteSelect;
