import React, { useState } from "react";
import { FiMaximize2, FiMinimize2 } from "react-icons/fi";
import { useI18n } from "../../i18n/I18nContext";
import Tooltip from "../common/Tooltip";
import DailyBarChart from "./DailyBarChart";
import ProjectDonutChart from "./ProjectDonutChart";
import { formatHMS } from "./reportFormat";
import styles from "./Styles/Reports.module.css";

const SummaryReportView = ({ report }) => {
  const { t } = useI18n();
  const { totals, daily = [], rows = [], projectDurations = [] } = report;
  const showAmount = totals.amount !== undefined && totals.amount !== null;
  // null | "donut" | "bar": which chart (if any) is maximized, masking the
  // other one entirely (unmounted, not just hidden).
  const [expandedChart, setExpandedChart] = useState(null);
  const toggleExpand = (chart) => setExpandedChart((current) => (current === chart ? null : chart));

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

      <div className={styles.chartsRow}>
        {expandedChart !== "bar" && projectDurations.length > 0 && (
          <div
            className={`${styles.chartPane} ${expandedChart === "donut" ? styles.chartPaneFull : styles.chartPaneDonut}`}
          >
            <div className={styles.expandBtnWrapper}>
              <Tooltip label={expandedChart === "donut" ? t("common.collapse") : t("common.expand")} position="bottom">
                <button
                  type="button"
                  className={styles.expandBtn}
                  onClick={() => toggleExpand("donut")}
                  aria-label={expandedChart === "donut" ? t("common.collapse") : t("common.expand")}
                >
                  {expandedChart === "donut" ? <FiMinimize2 /> : <FiMaximize2 />}
                </button>
              </Tooltip>
            </div>
            <ProjectDonutChart projectDurations={projectDurations} />
          </div>
        )}
        {expandedChart !== "donut" && daily.length > 0 && (
          <div
            className={`${styles.chartPane} ${expandedChart === "bar" ? styles.chartPaneFull : styles.chartPaneBar}`}
          >
            <div className={styles.expandBtnWrapper}>
              <Tooltip label={expandedChart === "bar" ? t("common.collapse") : t("common.expand")} position="bottom">
                <button
                  type="button"
                  className={styles.expandBtn}
                  onClick={() => toggleExpand("bar")}
                  aria-label={expandedChart === "bar" ? t("common.collapse") : t("common.expand")}
                >
                  {expandedChart === "bar" ? <FiMinimize2 /> : <FiMaximize2 />}
                </button>
              </Tooltip>
            </div>
            <DailyBarChart daily={daily} />
          </div>
        )}
      </div>

      <div className={styles.table}>
        <div className={styles.tableHeader}>
          <span>{t("projects.title")}</span>
          <span>{t("common.client")}</span>
          <span>{t("timeTracker.taskDescription")}</span>
          <span>{t("common.email")}</span>
          <span>{t("reports.duration")}</span>
          {showAmount && <span>{t("reports.amount")}</span>}
        </div>
        {rows.length === 0 && <p className={styles.empty}>{t("reports.noEntries")}</p>}
        {rows.map((row, i) => (
          <div className={styles.tableRow} key={`${row.projectId}-${i}`}>
            <span>{row.projectName}</span>
            <span>{row.clientName}</span>
            <span>{row.description}</span>
            <span>{row.userEmail}</span>
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
