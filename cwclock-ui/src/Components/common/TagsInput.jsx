import React, { useState } from "react";
import styles from "./Styles/TagsInput.module.css";

const TagsInput = ({ value = [], onChange, placeholder }) => {
  const [input, setInput] = useState("");

  const addTag = () => {
    const trimmed = input.trim();
    if (trimmed && !value.includes(trimmed)) {
      onChange([...value, trimmed]);
    }
    setInput("");
  };

  const removeTag = (tag) => {
    onChange(value.filter((t) => t !== tag));
  };

  const handleKeyDown = (e) => {
    if (e.key === "Enter") {
      e.preventDefault();
      addTag();
    }
  };

  return (
    <div className={styles.container}>
      {value.length > 0 && (
        <div className={styles.tags}>
          {value.map((tag) => (
            <span key={tag} className={styles.tag}>
              {tag}
              <button type="button" className={styles.remove} onClick={() => removeTag(tag)}>
                ×
              </button>
            </span>
          ))}
        </div>
      )}
      <div className={styles.inputRow}>
        <input
          className="cw-input"
          type="text"
          value={input}
          placeholder={placeholder}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
        />
        <button type="button" className={styles.addBtn} onClick={addTag}>
          +
        </button>
      </div>
    </div>
  );
};

export default TagsInput;
