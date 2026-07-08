import React from "react";
import projectLabel from "./projectLabel";
import contrastColor from "./contrastColor";
import styles from "./Styles/ProjectBadge.module.css";

// Renders "{client} - {project}" underlined and chipped in the project's own
// hex color, so records are visually grouped by project at a glance.
const ProjectBadge = ({ project, clients }) => {
  if (!project) return null;
  const color = project.color || "#1cb9f7";

  return (
    <span
      className={styles.badge}
      style={{ backgroundColor: color, color: contrastColor(color), textDecorationColor: color }}
    >
      {projectLabel(project, clients)}
    </span>
  );
};

export default ProjectBadge;
