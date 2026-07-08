import React, { useState } from "react";
import styles from "./Styles/DropZone.module.css";

// Wraps a file-upload control (a file input, a label, ...) so a file can
// also be dropped onto it instead of picked through the OS file dialog.
// Shows a highlight while a file is dragged over so an otherwise plain
// target (a label, a small input) doesn't look inert.
const DropZone = ({ onFile, className, children }) => {
  const [active, setActive] = useState(false);

  const handleDragOver = (e) => {
    e.preventDefault();
    setActive(true);
  };

  const handleDragLeave = () => setActive(false);

  const handleDrop = (e) => {
    e.preventDefault();
    setActive(false);
    const file = e.dataTransfer.files && e.dataTransfer.files[0];
    if (file) onFile(file);
  };

  return (
    <div
      className={[styles.dropZone, active ? styles.active : "", className].filter(Boolean).join(" ")}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      {children}
    </div>
  );
};

export default DropZone;
