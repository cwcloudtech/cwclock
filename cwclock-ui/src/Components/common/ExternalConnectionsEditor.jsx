import React, { useState } from "react";
import Button from "./Button";
import RequiredMark from "./RequiredMark";
import fileToBase64 from "./fileToBase64";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/ExternalConnectionsEditor.module.css";

const emptyDraft = {
  type: "s3",
  endpoint: "",
  bucketName: "",
  region: "",
  accessKey: "",
  secretKey: "",
  serviceAccountBase64: "",
  serviceAccountFileName: "",
  folderId: "",
};

const isDraftComplete = (draft) => {
  if (draft.type === "s3") {
    return !!(draft.endpoint && draft.bucketName && draft.region && draft.accessKey && draft.secretKey);
  }
  return !!(draft.serviceAccountBase64 && draft.folderId);
};

// Manages an organization's `externalConnections` list (ai-instruct-39): a
// "+" button reveals a per-type sub-form (S3 or Google Drive fields), and
// existing connections are listed with a remove button. The whole array
// rides along in the organization create/edit form's own fields object, so
// no dedicated Redux actions or API calls are needed here.
const ExternalConnectionsEditor = ({ value = [], onChange }) => {
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
    const connection = { id: crypto.randomUUID(), ...draft };
    delete connection.serviceAccountFileName;
    onChange([...value, connection]);
    setDraft(emptyDraft);
    setFileError("");
    setAdding(false);
  };

  const handleRemove = (id) => onChange(value.filter((c) => c.id !== id));

  const summaryFor = (conn) =>
    conn.type === "s3"
      ? `${conn.bucketName} (${conn.endpoint})`
      : `${t("organizations.googleDrive")} - ${conn.folderId}`;

  return (
    <div className={styles.container}>
      {value.length > 0 && (
        <ul className={styles.list}>
          {value.map((conn) => (
            <li key={conn.id} className={styles.row}>
              <span className={styles.typeBadge}>
                {conn.type === "s3" ? t("organizations.s3") : t("organizations.googleDrive")}
              </span>
              <span className={styles.summary}>{summaryFor(conn)}</span>
              <button
                type="button"
                className={styles.remove}
                onClick={() => handleRemove(conn.id)}
                title={t("common.delete")}
              >
                ×
              </button>
            </li>
          ))}
        </ul>
      )}

      {adding ? (
        <div className={styles.addForm}>
          <div className="cw-field">
            <label className="cw-label">{t("organizations.connectionType")}</label>
            <select
              className="cw-select"
              value={draft.type}
              onChange={(e) => setDraft({ ...emptyDraft, type: e.target.value })}
            >
              <option value="s3">{t("organizations.s3")}</option>
              <option value="google_drive">{t("organizations.googleDrive")}</option>
            </select>
          </div>

          {draft.type === "s3" ? (
            <>
              <div className="cw-field">
                <label className="cw-label">
                  {t("organizations.endpoint")}
                  <RequiredMark />
                </label>
                <input
                  className="cw-input"
                  type="text"
                  value={draft.endpoint}
                  onChange={(e) => setDraftField("endpoint", e.target.value)}
                />
              </div>
              <div className="cw-field">
                <label className="cw-label">
                  {t("organizations.bucketName")}
                  <RequiredMark />
                </label>
                <input
                  className="cw-input"
                  type="text"
                  value={draft.bucketName}
                  onChange={(e) => setDraftField("bucketName", e.target.value)}
                />
              </div>
              <div className="cw-field">
                <label className="cw-label">
                  {t("organizations.region")}
                  <RequiredMark />
                </label>
                <input
                  className="cw-input"
                  type="text"
                  value={draft.region}
                  onChange={(e) => setDraftField("region", e.target.value)}
                />
              </div>
              <div className="cw-field">
                <label className="cw-label">
                  {t("organizations.accessKey")}
                  <RequiredMark />
                </label>
                <input
                  className="cw-input"
                  type="text"
                  value={draft.accessKey}
                  onChange={(e) => setDraftField("accessKey", e.target.value)}
                />
              </div>
              <div className="cw-field">
                <label className="cw-label">
                  {t("organizations.secretKey")}
                  <RequiredMark />
                </label>
                <input
                  className="cw-input"
                  type="password"
                  value={draft.secretKey}
                  onChange={(e) => setDraftField("secretKey", e.target.value)}
                />
              </div>
            </>
          ) : (
            <>
              <div className="cw-field">
                <label className="cw-label">
                  {t("organizations.serviceAccount")}
                  <RequiredMark />
                </label>
                <input className="cw-input" type="file" accept=".json" onChange={handleServiceAccountFile} />
                {draft.serviceAccountFileName && (
                  <span className={styles.fileName}>{draft.serviceAccountFileName}</span>
                )}
                {fileError && <p className="cw-error">{fileError}</p>}
              </div>
              <div className="cw-field">
                <label className="cw-label">
                  {t("organizations.folderId")}
                  <RequiredMark />
                </label>
                <input
                  className="cw-input"
                  type="text"
                  value={draft.folderId}
                  onChange={(e) => setDraftField("folderId", e.target.value)}
                />
              </div>
            </>
          )}

          <div className={styles.addFormActions}>
            <Button
              type="button"
              size="sm"
              onClick={handleAdd}
              disabled={!isDraftComplete(draft)}
              title={t("organizations.addConnection")}
            >
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
              title={t("common.cancel")}
            >
              {t("common.cancel")}
            </Button>
          </div>
        </div>
      ) : (
        <button type="button" className={styles.addBtn} onClick={() => setAdding(true)} title={t("organizations.addConnection")}>
          + {t("organizations.addConnection")}
        </button>
      )}
    </div>
  );
};

export default ExternalConnectionsEditor;
