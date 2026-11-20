import React from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { FiDownload } from "react-icons/fi";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/Markdown.module.css";

// A link the assistant posts to a file it generated (a report/invoice PDF) points
// at the API's /ai-downloads/ endpoint. Those files are short-lived, so such a link
// is rendered specially — a download chip while it is valid, and a muted "expired"
// label once the message that carried it is older than the download TTL.
const isDownloadHref = (href) => typeof href === "string" && href.includes("/ai-downloads/");

// Markdown renders an assistant answer as GitHub-flavoured Markdown — headings,
// lists, code, links and, crucially, tables (remark-gfm) — the way the uprodit /
// strong-fish assistants do. Links open in a new tab; nothing dangerous is allowed
// through (react-markdown does not render raw HTML unless asked, and we never ask).
//
// downloadExpiresAt (ms epoch, or null) is when any download link in THIS message
// stops working — message createdAt + the TTL the API reported. null means unknown,
// in which case a download link is left as an ordinary, always-clickable link.
const Markdown = ({ children, downloadExpiresAt = null }) => {
  const { t } = useI18n();

  const renderAnchor = ({ node, href, children: linkChildren, ...props }) => {
    if (isDownloadHref(href)) {
      const expired = downloadExpiresAt != null && Date.now() >= downloadExpiresAt;
      if (expired) {
        return (
          <span className={styles.downloadExpired} title={t("ai.downloadExpiredHint")}>
            {linkChildren} <span className={styles.downloadExpiredTag}>({t("ai.expired")})</span>
          </span>
        );
      }
      return (
        <a
          href={href}
          target="_blank"
          rel="noopener noreferrer"
          className={styles.downloadLink}
          title={downloadExpiresAt != null ? t("ai.downloadTemporaryHint") : undefined}
        >
          <FiDownload className={styles.downloadIcon} aria-hidden="true" />
          {linkChildren}
        </a>
      );
    }
    return (
      <a href={href} target="_blank" rel="noopener noreferrer" {...props}>
        {linkChildren}
      </a>
    );
  };

  return (
    <div className={styles.markdown}>
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={{ a: renderAnchor }}>
        {children || ""}
      </ReactMarkdown>
    </div>
  );
};

export default Markdown;
