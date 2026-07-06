import React from "react";
import { FiCalendar } from "react-icons/fi";
import Dropdown, { DropdownItem } from "./Dropdown";
import { useI18n } from "../../i18n/I18nContext";
import { dateRangeShortcuts, toISODate } from "./dateRangeShortcuts";
import styles from "./Styles/DateRangePicker.module.css";

// Grafana-style time selector: two date pickers plus a shortcuts dropdown
// that sets both at once.
const DateRangePicker = ({ start, end, onChange }) => {
  const { t } = useI18n();
  const shortcuts = dateRangeShortcuts(t);

  const applyShortcut = (range, close) => {
    const [s, e] = range();
    onChange(toISODate(s), toISODate(e));
    close();
  };

  return (
    <div className={styles.wrapper}>
      <Dropdown
        title={t("reports.dateRangeShortcuts")}
        triggerClassName={styles.shortcutTrigger}
        trigger={
          <>
            <FiCalendar />
            {t("reports.shortcuts")}
          </>
        }
      >
        {(close) => shortcuts.map((s) => (
          <DropdownItem key={s.key} onClick={() => applyShortcut(s.range, close)}>
            {s.label}
          </DropdownItem>
        ))}
      </Dropdown>
      <input
        className="cw-input"
        type="date"
        value={start}
        max={end}
        title={t("reports.startDate")}
        onChange={(e) => onChange(e.target.value, end)}
      />
      <span className={styles.sep}>-</span>
      <input
        className="cw-input"
        type="date"
        value={end}
        min={start}
        title={t("reports.endDate")}
        onChange={(e) => onChange(start, e.target.value)}
      />
    </div>
  );
};

export default DateRangePicker;
