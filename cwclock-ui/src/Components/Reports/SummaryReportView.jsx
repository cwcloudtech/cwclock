import React from "react";
import { useI18n } from "../../i18n/I18nContext";
import DailyBarChart from "./DailyBarChart";
import { formatHMS } from "./reportFormat";
import styles from "./Styles/Reports.module.css";

const SummaryReportView = ({ report }) => {
  const { t } = useI18n();
  const { totals, daily, rows } = report;
  const showAmount = totals.amount !== undefined && totals.amount !== null;

  return (
    <div className={styles.reportBody}>
      <div className={styles.totals}>
        <span>
          <strong>{t("reports.total")}:</strong> {formatHMS(totals.durationSecs)}
        </span>
        <span>
          <strong>{t("reports.billable")}:</strong> {formatHMS(totals.durationSecs)}
        </span>
        {showAmount && (
          <span>
            <strong>{t("reports.amount")}:</strong> {totals.amount.toFixed(2)} {totals.currency}
          </span>
        )}
      </div>

      <DailyBarChart daily={daily} />

      <div className={styles.table}>
        <div className={styles.tableHeader}>
          <span>{t("projects.title")}</span>
          <span>{t("common.client")}</span>
          <span>{t("timeTracker.taskDescription")}</span>
          <span>{t("reports.duration")}</span>
          {showAmount && <span>{t("reports.amount")}</span>}
        </div>
        {rows.length === 0 && <p className={styles.empty}>{t("reports.noEntries")}</p>}
        {rows.map((row, i) => (
          <div className={styles.tableRow} key={`${row.projectId}-${i}`}>
            <span>{row.projectName}</span>
            <span>{row.clientName}</span>
            <span>{row.description}</span>
            <span>{formatHMS(row.durationSecs)}</span>
            {showAmount && (
              <span>
                {(row.amount ?? 0).toFixed(2)} {totals.currency}
              </span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

export default SummaryReportView;
