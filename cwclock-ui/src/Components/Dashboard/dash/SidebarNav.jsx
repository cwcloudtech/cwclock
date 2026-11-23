import React from "react";
import { NavLink } from "react-router-dom";
import { FiClock, FiKey, FiCalendar, FiMessageSquare } from "react-icons/fi";
import { FaFileAlt, FaRegUserCircle, FaBuilding, FaUserShield, FaChartBar, FaFileInvoiceDollar, FaDownload, FaClipboardList, FaNetworkWired, FaRegCalendarCheck } from "react-icons/fa";
import { FaBuildingShield } from "react-icons/fa6";
import Tooltip from "../../common/Tooltip";
import useAppVersion from "../../common/useAppVersion";
import { useI18n } from "../../../i18n/I18nContext";
import styles from "./STYLE/SidebarNav.module.css";

const items = [
  { to: "/dashboard/timetracker", labelKey: "nav.timeTracker", Icon: FiClock },
  { to: "/dashboard/calendar", labelKey: "nav.calendar", Icon: FiCalendar },
  { to: "/dashboard/reports", labelKey: "nav.reports", Icon: FaChartBar },
  { to: "/dashboard/organizations", labelKey: "nav.organizations", Icon: FaBuilding },
  { to: "/dashboard/clients", labelKey: "nav.clients", Icon: FaRegUserCircle },
  { to: "/dashboard/projects", labelKey: "nav.projects", Icon: FaFileAlt },
  { to: "/dashboard/api-keys", labelKey: "nav.apiKeys", Icon: FiKey },
];

const invoicesItem = { to: "/dashboard/invoices", labelKey: "nav.invoices", Icon: FaFileInvoiceDollar };
const exportJobsItem = { to: "/dashboard/export-jobs", labelKey: "nav.exportJobs", Icon: FaDownload };
const assistantItem = { to: "/dashboard/assistant", labelKey: "nav.assistant", Icon: FiMessageSquare };

const adminItems = [
  { to: "/dashboard/admin", labelKey: "nav.users", Icon: FaUserShield },
  { to: "/dashboard/admin/organizations", labelKey: "nav.organizationsAdmin", Icon: FaBuildingShield },
  { to: "/dashboard/admin/ip-bans", labelKey: "nav.ipManagement", Icon: FaNetworkWired },
  { to: "/dashboard/admin/audit", labelKey: "nav.auditTrail", Icon: FaClipboardList },
  { to: "/dashboard/admin/holidays", labelKey: "nav.holidays", Icon: FaRegCalendarCheck },
];

// Single sidebar nav that renders either as an icon rail or an expanded
// rail with labels, replacing the previously separate Slideopen/Slideclose
// components that duplicated the same link list. Invoices and Export Jobs
// only show for an admin/owner of the current organization (showInvoices),
// unlike the admin section below which is global-superuser-only (isSuperuser).
const SidebarNav = ({ expanded, isSuperuser, showInvoices, aiEnabled }) => {
  const { t } = useI18n();
  const appVersion = useAppVersion();
  let links = showInvoices
    ? [...items.slice(0, 3), invoicesItem, exportJobsItem, ...items.slice(3)]
    : items;
  // The assistant sits just after the time tracker once a subscription exists.
  if (aiEnabled) links = [links[0], assistantItem, ...links.slice(1)];
  if (isSuperuser) links = [...links, ...adminItems];
  return (
    <nav className={styles.nav}>
      {links.map(({ to, labelKey, Icon }) => {
        const label = t(labelKey);
        return (
          <Tooltip key={to} label={expanded ? null : label} position="right" className={styles.tooltipWrapper}>
            <NavLink
              to={to}
              end
              className={({ isActive }) => `${styles.link} ${isActive ? styles.linkActive : ""}`}
            >
              <Icon className={styles.icon} />
              {expanded && <span className={styles.label}>{label}</span>}
            </NavLink>
          </Tooltip>
        );
      })}
      {appVersion && (
        <div
          className={styles.version}
          style={{ marginTop: "auto", padding: "10px 6px", textAlign: "center", color: "var(--cw-text-muted)", fontSize: "0.72rem", fontStyle: "italic" }}
          title={`v${appVersion}`}
        >
          v{appVersion}
        </div>
      )}
    </nav>
  );
};

export default SidebarNav;
