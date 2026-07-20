import React from "react";
import { useI18n } from "../../i18n/I18nContext";
import ReportEntryRow from "./ReportEntryRow";
import { formatHMS } from "./reportFormat";
import styles from "./Styles/Reports.module.css";

const DetailedReportView = ({ report, orgId, isAdminOrOwner, onChanged }) => {
  const { t } = useI18n();
  const { totals, entries = [] } = report;
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
        <span>
          <strong>{t("reports.days")}:</strong> {(totals.days ?? 0).toFixed(2)}
        </span>
        {showAmount && (
          <span>
            <strong>{t("reports.amount")}:</strong> {totals.amount.toFixed(2)} {totals.currency}
          </span>
        )}
      </div>

      <div className={styles.table}>
        <div className={styles.tableHeader}>
          <span>{t("timeTracker.day")}</span>
          <span>{t("timeTracker.taskDescription")}</span>
          <span>{t("reports.time")}</span>
          <span>{t("reports.duration")}</span>
          <span>{t("nav.users")}</span>
          <span>{t("common.email")}</span>
          {showAmount && <span>{t("reports.amount")}</span>}
          <span />
        </div>
        {entries.length === 0 && <p className={styles.empty}>{t("reports.noEntries")}</p>}
        {entries.map((entry) => (
          <ReportEntryRow
            key={entry.id}
            entry={entry}
            orgId={orgId}
            currency={totals.currency}
            isAdminOrOwner={isAdminOrOwner}
            showAmount={showAmount}
            onChanged={onChanged}
          />
        ))}
      </div>
    </div>
  );
};

export default DetailedReportView;
