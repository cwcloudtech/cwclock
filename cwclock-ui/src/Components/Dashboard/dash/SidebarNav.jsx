import React from "react";
import { NavLink } from "react-router-dom";
import { FiClock } from "react-icons/fi";
import { FaFileAlt, FaRegUserCircle, FaBuilding } from "react-icons/fa";
import Tooltip from "../../common/Tooltip";
import styles from "./STYLE/SidebarNav.module.css";

const items = [
  { to: "/dashboard/timetracker", label: "Time Tracker", Icon: FiClock },
  { to: "/dashboard/organizations", label: "Organizations", Icon: FaBuilding },
  { to: "/dashboard/clients", label: "Clients", Icon: FaRegUserCircle },
  { to: "/dashboard/projects", label: "Projects", Icon: FaFileAlt },
];

// Single sidebar nav that renders either as an icon rail or an expanded
// rail with labels, replacing the previously separate Slideopen/Slideclose
// components that duplicated the same link list.
const SidebarNav = ({ expanded }) => {
  return (
    <nav className={styles.nav}>
      {items.map(({ to, label, Icon }) => (
        <Tooltip key={to} label={expanded ? null : label} position="right" className={styles.tooltipWrapper}>
          <NavLink
            to={to}
            className={({ isActive }) => `${styles.link} ${isActive ? styles.linkActive : ""}`}
          >
            <Icon className={styles.icon} />
            {expanded && <span className={styles.label}>{label}</span>}
          </NavLink>
        </Tooltip>
      ))}
    </nav>
  );
};

export default SidebarNav;
