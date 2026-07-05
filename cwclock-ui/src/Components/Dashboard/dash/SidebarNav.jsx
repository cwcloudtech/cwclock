import React from "react";
import { NavLink } from "react-router-dom";
import { FiClock } from "react-icons/fi";
import { FaFileAlt, FaRegUserCircle, FaBuilding, FaUserShield, FaShieldAlt } from "react-icons/fa";
import Tooltip from "../../common/Tooltip";
import styles from "./STYLE/SidebarNav.module.css";

const items = [
  { to: "/dashboard/timetracker", label: "Time Tracker", Icon: FiClock },
  { to: "/dashboard/organizations", label: "Organizations", Icon: FaBuilding },
  { to: "/dashboard/clients", label: "Clients", Icon: FaRegUserCircle },
  { to: "/dashboard/projects", label: "Projects", Icon: FaFileAlt },
];

const adminItems = [
  { to: "/dashboard/admin", label: "Users", Icon: FaUserShield },
  { to: "/dashboard/admin/organizations", label: "Organizations admin", Icon: FaShieldAlt },
];

// Single sidebar nav that renders either as an icon rail or an expanded
// rail with labels, replacing the previously separate Slideopen/Slideclose
// components that duplicated the same link list.
const SidebarNav = ({ expanded, isSuperuser }) => {
  const links = isSuperuser ? [...items, ...adminItems] : items;
  return (
    <nav className={styles.nav}>
      {links.map(({ to, label, Icon }) => (
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
