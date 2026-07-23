import React, { useState } from "react";
import { MdDeleteForever } from "react-icons/md";
import { FaRegCopy } from "react-icons/fa";
import { toast } from "react-toastify";
import Button from "../common/Button";
import RequiredMark from "../common/RequiredMark";
import Tooltip from "../common/Tooltip";
import ConnectionFields from "../common/ConnectionFields";
import fileToBase64 from "../common/fileToBase64";
import { useI18n } from "../../i18n/I18nContext";
import toastOptions from "../../Redux/toastOptions";
import styles from "./Styles/ExportTargetsEditor.module.css";

const TARGET_TYPES = ["email", "s3", "google_drive", "git"];

const emptyDraft = {
  type: "email",
  toEmails: "",
  ccEmails: "",
  endpoint: "",
  bucketName: "",
  region: "",
  accessKey: "",
  secretKey: "",
  serviceAccountBase64: "",
  serviceAccountFileName: "",
  folderId: "",
  repoUrl: "",
  gitAuthMethod: "password",
  username: "",
  password: "",
  sshPrivateKey: "",
  sshPrivateKeyPassphrase: "",
  path: "",
  flatDirectory: false,
};

const typeLabel = (t, type) => {
  if (type === "email") return t("exportJobs.targetTypeEmail");
  if (type === "s3") return t("organizations.s3");
  if (type === "git") return t("organizations.git");
  return t("organizations.googleDrive");
};

// Mirrors ExternalConnectionsEditor's summaryFor, minus the flatDirectory
// badge.
const connectionSummary = (conn) => {
  if (!conn) return "";
  let base;
  if (conn.type === "s3") base = `${conn.bucketName} (${conn.endpoint})`;
  else if (conn.type === "git") base = conn.repoUrl;
  else base = conn.folderId;
  return conn.path ? `${base} - ${conn.path}` : base;
};

const isDraftComplete = (d) => {
  if (d.type === "email") return !!d.toEmails.trim();
  if (d.type === "s3") return !!(d.endpoint && d.bucketName && d.region && d.accessKey && d.secretKey);
  if (d.type === "git") {
    return !!(d.repoUrl && (d.gitAuthMethod === "password" ? d.username && d.password : d.sshPrivateKey));
  }
  return !!(d.serviceAccountBase64 && d.folderId);
};

// Builds the final target object saved onto the job from an add-form draft:
// an "email" target, or a target carrying its own S3/Google Drive/git
// connection - the same shape as models.ExternalConnection, captured
// through the same fields as ExternalConnectionsEditor, but stored directly
// on the target instead of referencing one of the organization's own
// connections (see ai-instruct-77 - a job can push to a different bucket
// than the one configured for invoices, and isn't affected by that
// connection later being edited or removed).
const buildTarget = (draft) => {
  if (draft.type === "email") {
    return { type: "email", toEmails: draft.toEmails, ccEmails: draft.ccEmails };
  }
  return {
    type: draft.type,
    connection: {
      type: draft.type,
      endpoint: draft.endpoint,
      bucketName: draft.bucketName,
      region: draft.region,
      accessKey: draft.accessKey,
      secretKey: draft.secretKey,
      serviceAccountBase64: draft.serviceAccountBase64,
      folderId: draft.folderId,
      repoUrl: draft.repoUrl,
      username: draft.gitAuthMethod === "password" ? draft.username : "",
      password: draft.gitAuthMethod === "password" ? draft.password : "",
      sshPrivateKey: draft.gitAuthMethod === "sshKey" ? draft.sshPrivateKey : "",
      sshPrivateKeyPassphrase: draft.gitAuthMethod === "sshKey" ? draft.sshPrivateKeyPassphrase : "",
      path: draft.path,
      flatDirectory: draft.flatDirectory,
    },
  };
};

// Manages an export job's `targets` list: as many targets as wanted, each
// either an "email" (a To/CC address list) or its own independent
// S3/Google Drive/git connection (see buildTarget). The connection
// sub-form is the same one ExternalConnectionsEditor uses for an
// organization's external connections - reused here field-for-field, but
// the result is kept on the target itself rather than added to the
// organization.
const ExportTargetsEditor = ({ targets, onChange, error }) => {
  const { t } = useI18n();
  const [adding, setAdding] = useState(false);
  const [draft, setDraft] = useState(emptyDraft);
  const [fileError, setFileError] = useState("");

  const setDraftField = (key, val) => setDraft((d) => ({ ...d, [key]: val }));

  const handleServiceAccountFile = async (e) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setFileError("");
    try {
      const dataUrl = await fileToBase64(file);
      const base64 = dataUrl.split(",")[1] || "";
      setDraft((d) => ({ ...d, serviceAccountBase64: base64, serviceAccountFileName: file.name }));
    } catch {
      setFileError(t("organizations.serviceAccountReadError"));
    }
  };

  const handleAdd = () => {
    if (!isDraftComplete(draft)) return;
    onChange([...targets, buildTarget(draft)]);
    setDraft(emptyDraft);
    setFileError("");
    setAdding(false);
  };

  const handleRemove = (index) => {
    onChange(targets.filter((_, i) => i !== index));
  };

  const handleCopyToEmails = (toEmails) => {
    navigator.clipboard.writeText(toEmails).then(() => {
      toast.success(t("common.copied"), toastOptions);
    });
  };

  const rowSummary = (target) => {
    if (target.type === "email") {
      return target.ccEmails ? `${target.toEmails} (cc: ${target.ccEmails})` : target.toEmails;
    }
    return connectionSummary(target.connection);
  };

  return (
    <div className={styles.container}>
      {targets.length > 0 && (
        <ul className={styles.list}>
          {targets.map((target, index) => (
            <li key={index} className={styles.row}>
              <span className={styles.typeBadge}>{typeLabel(t, target.type)}</span>
              <span className={styles.summary}>{rowSummary(target)}</span>
              {target.type === "email" && (
                <Tooltip label={t("common.copy")}>
                  <button
                    type="button"
                    className={styles.iconBtn}
                    onClick={() => handleCopyToEmails(target.toEmails)}
                  >
                    <FaRegCopy style={{ fontSize: "16px" }} />
                  </button>
                </Tooltip>
              )}
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

          {draft.type === "email" && (
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
                  onChange={(e) => setDraftField("toEmails", e.target.value)}
                  placeholder="email1@example.com, email2@example.com"
                />
              </div>
              <div className="cw-field">
                <label className="cw-label">{t("exportJobs.ccEmails")}</label>
                <input
                  className="cw-input"
                  type="text"
                  value={draft.ccEmails}
                  onChange={(e) => setDraftField("ccEmails", e.target.value)}
                  placeholder="cc@example.com"
                />
              </div>
            </>
          )}

          {draft.type !== "email" && (
            <ConnectionFields
              draft={draft}
              setDraftField={setDraftField}
              onServiceAccountFile={handleServiceAccountFile}
              fileError={fileError}
              showFlatDirectory={false}
            />
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
                setFileError("");
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
