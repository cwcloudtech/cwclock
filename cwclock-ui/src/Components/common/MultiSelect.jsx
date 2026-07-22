import React, { useMemo, useState } from "react";
import Dropdown, { DropdownText } from "./Dropdown";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/MultiSelect.module.css";

// Autocomplete + multiselect filter dropdown (client/project/member): a
// search box narrows the option list, checkboxes toggle selection, and the
// panel stays open across toggles so several options can be picked in a row.
const MultiSelect = ({ label, options, selected, onChange }) => {
  const { t } = useI18n();
  const [query, setQuery] = useState("");

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return options;
    return options.filter((o) => o.label.toLowerCase().includes(q));
  }, [options, query]);

  const toggle = (value) => {
    onChange(selected.includes(value) ? selected.filter((v) => v !== value) : [...selected, value]);
  };

  // Selects every currently filtered option (merged with whatever's already
  // selected outside the search), so typing a query first then "select all"
  // picks just that subset instead of everything.
  const selectAllFiltered = () => {
    onChange(Array.from(new Set([...selected, ...filtered.map((o) => o.value)])));
  };
  const allFilteredSelected = filtered.length > 0 && filtered.every((o) => selected.includes(o.value));

  const summary =
    selected.length === 0
      ? t("common.all")
      : selected.length === 1
      ? options.find((o) => o.value === selected[0])?.label || t("common.selectedOne")
      : t("common.selectedCount", { count: selected.length });

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
      {() => (
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
              <label key={o.value} className={styles.option}>
                <input type="checkbox" checked={selected.includes(o.value)} onChange={() => toggle(o.value)} />
                {o.label}
              </label>
            ))}
          </div>
          {(!allFilteredSelected && filtered.length > 0) || selected.length > 0 ? (
            <div className={styles.actions}>
              {!allFilteredSelected && filtered.length > 0 && (
                <button type="button" className={styles.clear} onClick={selectAllFiltered}>
                  {t("common.selectAll")}
                </button>
              )}
              {selected.length > 0 && (
                <button type="button" className={styles.clear} onClick={() => onChange([])}>
                  {t("common.clearFilter")}
                </button>
              )}
            </div>
          ) : null}
        </div>
      )}
    </Dropdown>
  );
};

export default MultiSelect;
