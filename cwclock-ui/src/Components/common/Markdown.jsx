import React from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import styles from "./Styles/Markdown.module.css";

// Markdown renders an assistant answer as GitHub-flavoured Markdown — headings,
// lists, code, links and, crucially, tables (remark-gfm) — the way the uprodit /
// strong-fish assistants do. Links open in a new tab; nothing dangerous is allowed
// through (react-markdown does not render raw HTML unless asked, and we never ask).
const Markdown = ({ children }) => (
  <div className={styles.markdown}>
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        a: ({ node, children, ...props }) => (
          <a target="_blank" rel="noopener noreferrer" {...props}>
            {children}
          </a>
        ),
      }}
    >
      {children || ""}
    </ReactMarkdown>
  </div>
);

export default Markdown;
