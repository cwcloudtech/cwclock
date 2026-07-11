import React from "react";
import { useI18n } from "../../i18n/I18nContext";
import Tooltip from "../common/Tooltip";
import { formatHours } from "./reportFormat";
import styles from "./Styles/Reports.module.css";

// Single-series magnitude-over-time bar chart (hours worked per day). Only
// ~8 evenly spaced day labels are shown regardless of range length, matching
// how the reference report sparsely labels a several-week axis; every bar
// still carries its exact value in the shared Tooltip component (same bubble
// design used by the project donut chart next to it).
const DailyBarChart = ({ daily }) => {
  const { t } = useI18n();
  if (!daily || daily.length === 0) return null;

  const max = Math.max(...daily.map((d) => d.durationSecs), 1);
  const labelStep = Math.max(1, Math.ceil(daily.length / 8));

  return (
    <div className={styles.chart}>
      <div className={styles.chartBars}>
        {daily.map((d, i) => {
          const heightPct = (d.durationSecs / max) * 100;
          const date = new Date(`${d.day}T00:00:00`);
          const tooltip = `${date.toLocaleDateString(undefined, { weekday: "short", month: "short", day: "numeric" })}: ${formatHours(d.durationSecs)}`;
          return (
            <Tooltip key={d.day} label={tooltip} position="top" className={styles.chartBarColTooltip}>
              <div className={styles.chartBarCol}>
                <div className={styles.chartBarTrack}>
                  <div className={styles.chartBar} style={{ height: `${heightPct}%` }} />
                </div>
                <span className={styles.chartBarLabel}>
                  {i % labelStep === 0 ? date.toLocaleDateString(undefined, { month: "short", day: "numeric" }) : ""}
                </span>
              </div>
            </Tooltip>
          );
        })}
      </div>
      <p className={styles.chartHint}>{t("reports.chartHint")}</p>
    </div>
  );
};

export default DailyBarChart;
