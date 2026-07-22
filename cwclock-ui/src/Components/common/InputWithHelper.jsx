import React, { useMemo, useState } from "react";
import { FaChevronDown } from "react-icons/fa";
import Dropdown, { DropdownItem, DropdownText } from "./Dropdown";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/InputWithHelper.module.css";

// A free-text cw-input paired with a searchable dropdown of common presets
// (e.g. cron expressions, relative time periods) - picking a suggestion
// fills the input, the same autocomplete UX as AutocompleteSelect/
// MultiSelect, but the input itself stays directly editable for values
// that aren't in the preset list.
const InputWithHelper = ({ id, value, onChange, options, placeholder, error, disabled }) => {
  const { t } = useI18n();
  const [query, setQuery] = useState("");

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return options;
    return options.filter(
      (o) => o.label.toLowerCase().includes(q) || o.value.toLowerCase().includes(q)
    );
  }, [options, query]);

  return (
    <div className={styles.group}>
      <input
        id={id}
        type="text"
        className={`cw-input ${styles.input} ${error ? styles.inputError : ""}`}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        disabled={disabled}
      />
      <Dropdown
        title={t("common.suggestions")}
        align="end"
        triggerClassName={styles.trigger}
        disabled={disabled}
        trigger={<FaChevronDown aria-hidden="true" />}
      >
        {(close) => (
          <div className={styles.panel}>
            <input
              className="cw-input"
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
          </div>
        )}
      </Dropdown>
    </div>
  );
};

export default InputWithHelper;
