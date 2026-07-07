import React from "react";
import { NavLink } from "react-router-dom";
import { FiClock } from "react-icons/fi";
import { FaFileAlt, FaRegUserCircle, FaBuilding, FaUserShield, FaChartBar } from "react-icons/fa";
import { FaBuildingShield } from "react-icons/fa6";
import Tooltip from "../../common/Tooltip";
import { useI18n } from "../../../i18n/I18nContext";
import styles from "./STYLE/SidebarNav.module.css";

const items = [
  { to: "/dashboard/timetracker", labelKey: "nav.timeTracker", Icon: FiClock },
  { to: "/dashboard/reports", labelKey: "nav.reports", Icon: FaChartBar },
  { to: "/dashboard/organizations", labelKey: "nav.organizations", Icon: FaBuilding },
  { to: "/dashboard/clients", labelKey: "nav.clients", Icon: FaRegUserCircle },
  { to: "/dashboard/projects", labelKey: "nav.projects", Icon: FaFileAlt },
];

const adminItems = [
  { to: "/dashboard/admin", labelKey: "nav.users", Icon: FaUserShield },
  { to: "/dashboard/admin/organizations", labelKey: "nav.organizationsAdmin", Icon: FaBuildingShield },
];

// Single sidebar nav that renders either as an icon rail or an expanded
// rail with labels, replacing the previously separate Slideopen/Slideclose
// components that duplicated the same link list.
const SidebarNav = ({ expanded, isSuperuser }) => {
  const { t } = useI18n();
  const links = isSuperuser ? [...items, ...adminItems] : items;
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
    </nav>
  );
};

export default SidebarNav;
