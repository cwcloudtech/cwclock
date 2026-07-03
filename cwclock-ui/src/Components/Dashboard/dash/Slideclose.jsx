import React from "react";

import { NavLink } from "react-router-dom";
import OverlayTrigger from "react-bootstrap/OverlayTrigger";
import Tooltip from "react-bootstrap/Tooltip";

import styles from "./STYLE/Slideclose.module.css";
const navbaractive = {
  backgroundColor: "var(--cw-text)",
  color: "var(--cw-primary)",
  fontWeight: "bold",
  borderLeft: "4px solid var(--cw-primary)",
};
const navbarnotactive = {
  backgroundColor: "var(--cw-bg-alt)",
  color: "var(--cw-text)",
  borderLeft: "4px solid transparent",
};

const items = [
  { to: "/dashboard/timetracker", label: "TIME TRACKER" },
  { to: "/dashboard/organizations", label: "ORGANIZATIONS" },
  { to: "/dashboard/clients", label: "CLIENTS" },
  { to: "/dashboard/projects", label: "PROJECTS" },
];

const Slideclose = () => {
  return (
    <div className={styles.main}>
      <div className={styles.iconsname}>
        {items.map(({ to, label }) => (
          <OverlayTrigger key={to} placement="right" overlay={<Tooltip>{label}</Tooltip>}>
            <NavLink
              to={to}
              style={({ isActive }) => (isActive ? navbaractive : navbarnotactive)}
            >
              {" "}
              <div>
                <p>{label}</p>
              </div>
            </NavLink>
          </OverlayTrigger>
        ))}
      </div>
    </div>
  );
};

export default Slideclose;
