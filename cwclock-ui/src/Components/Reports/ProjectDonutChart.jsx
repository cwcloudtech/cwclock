import React, { useState } from "react";
import { useI18n } from "../../i18n/I18nContext";
import { formatHours } from "./reportFormat";
import styles from "./Styles/Reports.module.css";
import tooltipStyles from "../common/Styles/Tooltip.module.css";

const RADIUS = 70;
const STROKE_WIDTH = 34;
const CIRCUMFERENCE = 2 * Math.PI * RADIUS;

// Hours-per-project donut chart. SVG <circle> children can't be wrapped by
// the shared Tooltip component (it renders an HTML <span>, which doesn't
// lay out correctly inside <svg>), so hover state is tracked manually and a
// single floating bubble is rendered using Tooltip.module.css's own .bubble
// class - the exact same design every other tooltip in the app uses.
const ProjectDonutChart = ({ projectDurations = [] }) => {
  const { t } = useI18n();
  const [hovered, setHovered] = useState(null);
  const [pos, setPos] = useState({ x: 0, y: 0 });

  const total = projectDurations.reduce((sum, p) => sum + p.durationSecs, 0);
  if (total === 0) return null;

  let offset = 0;
  const segments = projectDurations.map((p) => {
    const length = (p.durationSecs / total) * CIRCUMFERENCE;
    const segment = { ...p, length, offset };
    offset += length;
    return segment;
  });

  const handleMouseMove = (e) => {
    const rect = e.currentTarget.getBoundingClientRect();
    setPos({ x: e.clientX - rect.left, y: e.clientY - rect.top });
  };

  return (
    <div className={styles.chart}>
      <div className={styles.donutWrapper} onMouseMove={handleMouseMove}>
        <svg viewBox="0 0 200 200" className={styles.donutSvg}>
          <g transform="rotate(-90 100 100)">
            {segments.map((seg) => (
              <circle
                key={seg.projectId}
                cx="100"
                cy="100"
                r={RADIUS}
                fill="none"
                stroke={seg.color || "#1cb9f7"}
                strokeWidth={STROKE_WIDTH}
                strokeDasharray={`${seg.length} ${CIRCUMFERENCE - seg.length}`}
                strokeDashoffset={-seg.offset}
                className={styles.donutSegment}
                tabIndex={0}
                onMouseEnter={() => setHovered(seg)}
                onMouseLeave={() => setHovered(null)}
                onFocus={() => setHovered(seg)}
                onBlur={() => setHovered(null)}
              />
            ))}
          </g>
          <text x="100" y="96" textAnchor="middle" className={styles.donutTotalValue}>
            {formatHours(total)}
          </text>
          <text x="100" y="116" textAnchor="middle" className={styles.donutTotalLabel}>
            {t("reports.total")}
          </text>
        </svg>
        {hovered && (
          <span
            className={tooltipStyles.bubble}
            role="tooltip"
            style={{ position: "absolute", opacity: 1, left: pos.x + 12, top: pos.y + 12 }}
          >
            {`${hovered.projectName}: ${formatHours(hovered.durationSecs)} (${Math.round((hovered.durationSecs / total) * 100)}%)`}
          </span>
        )}
      </div>
      <p className={styles.chartHint}>{t("reports.donutHint")}</p>
    </div>
  );
};

export default ProjectDonutChart;
