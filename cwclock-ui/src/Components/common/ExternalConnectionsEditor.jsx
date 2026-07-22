import React, { useState } from "react";
import { useDispatch } from "react-redux";
import { MdDeleteForever } from "react-icons/md";
import Button from "./Button";
import Tooltip from "./Tooltip";
import ConfirmModal from "./ConfirmModal";
import ConnectionFields from "./ConnectionFields";
import fileToBase64 from "./fileToBase64";
import { addExternalConnectionApi, removeExternalConnectionApi } from "../../Redux/Organizations/Org.actions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
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
  repoUrl: "",
  gitAuthMethod: "password",
  username: "",
  password: "",
  sshPrivateKey: "",
  sshPrivateKeyPassphrase: "",
  path: "",
  flatDirectory: false,
};

const isDraftComplete = (draft) => {
  if (draft.type === "s3") {
    return !!(draft.endpoint && draft.bucketName && draft.region && draft.accessKey && draft.secretKey);
  }
  if (draft.type === "git") {
    return !!(
      draft.repoUrl &&
      (draft.gitAuthMethod === "password" ? draft.username && draft.password : draft.sshPrivateKey)
    );
  }
  return !!(draft.serviceAccountBase64 && draft.folderId);
};

// Manages an organization's `externalConnections` list (ai-instruct-39/40):
// a "+" button reveals a per-type sub-form (S3 or Google Drive fields).
// When orgId/token are given (editing an existing organization), adding or
// removing a connection saves immediately via dedicated PATCH endpoints and
// refreshes the organization in place; otherwise (creating a brand new
// organization, which doesn't exist yet to PATCH) both just mutate the
// local array, submitted along with the rest of the create form.
const ExternalConnectionsEditor = ({ value = [], onChange, orgId, token }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const [adding, setAdding] = useState(false);
  const [draft, setDraft] = useState(emptyDraft);
  const [fileError, setFileError] = useState("");
  const [addError, setAddError] = useState("");
  const [saving, setSaving] = useState(false);
  const [deletingConnection, setDeletingConnection] = useState(null);
  const [removeError, setRemoveError] = useState("");

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

  const handleAdd = async () => {
    if (!isDraftComplete(draft)) return;
    setAddError("");
    const connection = { ...draft };
    delete connection.serviceAccountFileName;
    delete connection.gitAuthMethod;

    if (orgId) {
      setSaving(true);
      try {
        const updatedOrg = await dispatch(addExternalConnectionApi(orgId, connection, token));
        onChange(updatedOrg.externalConnections);
        setDraft(emptyDraft);
        setFileError("");
        setAdding(false);
      } catch (err) {
        setAddError(apiErrorMessage(err, locale));
      } finally {
        setSaving(false);
      }
      return;
    }

    onChange([...value, { id: crypto.randomUUID(), ...connection }]);
    setDraft(emptyDraft);
    setFileError("");
    setAdding(false);
  };

  const handleConfirmRemove = async () => {
    const connection = deletingConnection;
    setRemoveError("");

    if (orgId) {
      try {
        const updatedOrg = await dispatch(removeExternalConnectionApi(orgId, connection.id, token));
        onChange(updatedOrg.externalConnections);
        setDeletingConnection(null);
      } catch (err) {
        setRemoveError(apiErrorMessage(err, locale));
      }
      return;
    }

    onChange(value.filter((c) => c.id !== connection.id));
    setDeletingConnection(null);
  };

  const summaryFor = (conn) => {
    let base;
    if (conn.type === "s3") {
      base = `${conn.bucketName} (${conn.endpoint})`;
    } else if (conn.type === "git") {
      base = conn.repoUrl;
    } else {
      base = `${t("organizations.googleDrive")} - ${conn.folderId}`;
    }
    if (conn.path) base = `${base} - ${conn.path}`;
    return conn.flatDirectory ? `${base} - ${t("organizations.flatDirectoryBadge")}` : base;
  };

  return (
    <div className={styles.container}>
      {value.length > 0 && (
        <ul className={styles.list}>
          {value.map((conn) => (
            <li key={conn.id} className={styles.row}>
              <span className={styles.typeBadge}>
                {conn.type === "s3" && t("organizations.s3")}
                {conn.type === "git" && t("organizations.git")}
                {conn.type !== "s3" && conn.type !== "git" && t("organizations.googleDrive")}
              </span>
              <span className={styles.summary}>{summaryFor(conn)}</span>
              <Tooltip label={t("common.delete")}>
                <button
                  type="button"
                  className={`${styles.iconBtn} ${styles.iconBtnDanger}`}
                  onClick={() => setDeletingConnection(conn)}
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
            <label className="cw-label">{t("organizations.connectionType")}</label>
            <select
              className="cw-select"
              value={draft.type}
              onChange={(e) => setDraft({ ...emptyDraft, type: e.target.value })}
            >
              <option value="s3">{t("organizations.s3")}</option>
              <option value="google_drive">{t("organizations.googleDrive")}</option>
              <option value="git">{t("organizations.git")}</option>
            </select>
          </div>

          <ConnectionFields
            draft={draft}
            setDraftField={setDraftField}
            onServiceAccountFile={handleServiceAccountFile}
            fileError={fileError}
            showFlatDirectory
          />

          <div className={styles.addFormActions}>
            <Button
              type="button"
              size="sm"
              onClick={handleAdd}
              disabled={!isDraftComplete(draft) || saving}
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
                setAddError("");
              }}
              title={t("common.cancel")}
            >
              {t("common.cancel")}
            </Button>
          </div>
          {addError && <p className="cw-error">{addError}</p>}
        </div>
      ) : (
        <button type="button" className={styles.addBtn} onClick={() => setAdding(true)} title={t("organizations.addConnection")}>
          + {t("organizations.addConnection")}
        </button>
      )}

      <ConfirmModal
        show={!!deletingConnection}
        title={t("organizations.deleteConnectionTitle")}
        body={
          deletingConnection && (
            <>
              {t("organizations.deleteConnectionBody", { name: summaryFor(deletingConnection) })}
              {removeError && <p className="cw-error">{removeError}</p>}
            </>
          )
        }
        confirmLabel={t("common.delete")}
        onConfirm={handleConfirmRemove}
        onCancel={() => {
          setDeletingConnection(null);
          setRemoveError("");
        }}
      />
    </div>
  );
};

export default ExternalConnectionsEditor;
