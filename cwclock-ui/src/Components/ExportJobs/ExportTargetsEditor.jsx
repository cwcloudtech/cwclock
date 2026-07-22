import React, { useState } from "react";
import { MdDeleteForever } from "react-icons/md";
import Button from "../common/Button";
import RequiredMark from "../common/RequiredMark";
import AutocompleteSelect from "../common/AutocompleteSelect";
import Tooltip from "../common/Tooltip";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/ExportTargetsEditor.module.css";

const TARGET_TYPES = ["email", "s3", "google_drive", "git"];

const emptyDraft = { type: "email", toEmails: "", ccEmails: "", connection: "" };

const typeLabel = (t, type) => {
  if (type === "email") return t("exportJobs.targetTypeEmail");
  if (type === "s3") return t("organizations.s3");
  if (type === "git") return t("organizations.git");
  return t("organizations.googleDrive");
};

// Mirrors ExternalConnectionsEditor's summaryFor, minus the flatDirectory
// badge (not relevant to picking an existing connection).
const connectionSummary = (conn) => {
  if (!conn) return null;
  if (conn.type === "s3") return `${conn.bucketName} (${conn.endpoint})`;
  if (conn.type === "git") return conn.path ? `${conn.repoUrl} (${conn.path})` : conn.repoUrl;
  return conn.folderId;
};

// Manages an export job's `targets` list: as many targets as wanted, each
// either an "email" (a To/CC address list, entered directly) or a reference
// to one of the organization's existing S3/Google Drive/git external
// connections (see common/ExternalConnectionsEditor, ai-instruct-39/40) -
// targets never hold their own credentials, only the connection's id.
const ExportTargetsEditor = ({ targets, onChange, connections, error }) => {
  const { t } = useI18n();
  const [adding, setAdding] = useState(false);
  const [draft, setDraft] = useState(emptyDraft);

  const connectionsByType = (type) => connections.filter((c) => c.type === type);

  const isDraftComplete = (d) => (d.type === "email" ? !!d.toEmails.trim() : !!d.connection);

  const handleAdd = () => {
    if (!isDraftComplete(draft)) return;
    onChange([...targets, draft]);
    setDraft(emptyDraft);
    setAdding(false);
  };

  const handleRemove = (index) => {
    onChange(targets.filter((_, i) => i !== index));
  };

  const rowSummary = (target) => {
    if (target.type === "email") {
      return target.ccEmails ? `${target.toEmails} (cc: ${target.ccEmails})` : target.toEmails;
    }
    return connectionSummary(connections.find((c) => c.id === target.connection)) || t("exportJobs.connectionNotFound");
  };

  return (
    <div className={styles.container}>
      {targets.length > 0 && (
        <ul className={styles.list}>
          {targets.map((target, index) => (
            <li key={index} className={styles.row}>
              <span className={styles.typeBadge}>{typeLabel(t, target.type)}</span>
              <span className={styles.summary}>{rowSummary(target)}</span>
              <Tooltip label={t("exportJobs.removeTarget")}>
                <button
                  type="button"
                  className={`${styles.iconBtn} ${styles.iconBtnDanger}`}
                  onClick={() => handleRemove(index)}
                >
                  <MdDeleteForever style={{ fontSize: "18px" }} />
                </button>
              </Tooltip>
            </li>
          ))}
        </ul>
      )}

      {adding ? (
        <div className={styles.addForm}>
          <div className="cw-field">
            <label className="cw-label">{t("exportJobs.targetType")}</label>
            <select
              className="cw-select"
              value={draft.type}
              onChange={(e) => setDraft({ ...emptyDraft, type: e.target.value })}
            >
              {TARGET_TYPES.map((type) => (
                <option key={type} value={type}>
                  {typeLabel(t, type)}
                </option>
              ))}
            </select>
          </div>

          {draft.type === "email" ? (
            <>
              <div className="cw-field">
                <label className="cw-label">
                  {t("exportJobs.toEmails")}
                  <RequiredMark />
                </label>
                <input
                  className="cw-input"
                  type="text"
                  value={draft.toEmails}
                  onChange={(e) => setDraft({ ...draft, toEmails: e.target.value })}
                  placeholder="email1@example.com, email2@example.com"
                />
              </div>
              <div className="cw-field">
                <label className="cw-label">{t("exportJobs.ccEmails")}</label>
                <input
                  className="cw-input"
                  type="text"
                  value={draft.ccEmails}
                  onChange={(e) => setDraft({ ...draft, ccEmails: e.target.value })}
                  placeholder="cc@example.com"
                />
              </div>
            </>
          ) : connectionsByType(draft.type).length === 0 ? (
            <p className={styles.hint}>{t("exportJobs.noConnectionsForType", { type: typeLabel(t, draft.type) })}</p>
          ) : (
            <div className="cw-field">
              <label className="cw-label">
                {t("exportJobs.selectConnection")}
                <RequiredMark />
              </label>
              <AutocompleteSelect
                options={connectionsByType(draft.type).map((c) => ({ value: c.id, label: connectionSummary(c) }))}
                value={draft.connection}
                onChange={(value) => setDraft({ ...draft, connection: value })}
                placeholder={t("exportJobs.selectConnection")}
              />
            </div>
          )}

          <div className={styles.addFormActions}>
            <Button type="button" size="sm" onClick={handleAdd} disabled={!isDraftComplete(draft)}>
              {t("common.add")}
            </Button>
            <Button
              type="button"
              size="sm"
              variant="secondary"
              onClick={() => {
                setAdding(false);
                setDraft(emptyDraft);
              }}
            >
              {t("common.cancel")}
            </Button>
          </div>
        </div>
      ) : (
        <button type="button" className={styles.addBtn} onClick={() => setAdding(true)}>
          + {t("exportJobs.addTarget")}
        </button>
      )}

      {error && <p className="cw-error">{error}</p>}
    </div>
  );
};

export default ExportTargetsEditor;
