import React from "react";
import styles from "./STYLE/Slideopen.module.css";
import { FiClock } from "react-icons/fi";
import { FaFileAlt, FaRegUserCircle, FaBuilding } from "react-icons/fa";
import { NavLink } from "react-router-dom";
import OverlayTrigger from "react-bootstrap/OverlayTrigger";
import Tooltip from "react-bootstrap/Tooltip";
const navbaractive = {
  backgroundColor: "var(--cw-text)",
  color: "var(--cw-primary)",
  borderLeft: "4px solid var(--cw-primary)",
};
const navbarnotactive = {
  backgroundColor: "var(--cw-bg-alt)",
  color: "var(--cw-text)",
  borderLeft: "4px solid transparent",
};

const items = [
  { to: "/dashboard/timetracker", label: "Time Tracker", Icon: FiClock },
  { to: "/dashboard/organizations", label: "Organizations", Icon: FaBuilding },
  { to: "/dashboard/clients", label: "Clients", Icon: FaRegUserCircle },
  { to: "/dashboard/projects", label: "Projects", Icon: FaFileAlt },
];

const Slideopen = () => {
  return (
    <div className={styles.main}>
      <div className={styles.iconsbox}>
        {items.map(({ to, label, Icon }) => (
          <OverlayTrigger key={to} placement="right" overlay={<Tooltip>{label}</Tooltip>}>
            <NavLink
              to={to}
              style={({ isActive }) => (isActive ? navbaractive : navbarnotactive)}
            >
              {" "}
              <div>
                <Icon className={styles.icons1} />
              </div>
            </NavLink>
          </OverlayTrigger>
        ))}
      </div>
    </div>
  );
};

export default Slideopen;
