import React from "react";
import RequiredMark from "./RequiredMark";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/ConnectionFields.module.css";

// Renders the S3 / git / Google Drive connection sub-form fields shared by
// ExternalConnectionsEditor (an organization's own external connections)
// and ExportTargetsEditor (an export job's independent per-target
// connection, see ai-instruct-77) - same models.ExternalConnection shape,
// same fields, captured through the same draft object either way.
// showFlatDirectory controls whether the "Flat directory" checkbox is
// offered at all: an organization's own connection can choose either
// layout, but an export job's target connection is always forced flat
// server-side regardless of input (see exportJobPayload.targetsValid), so
// ExportTargetsEditor never shows the checkbox in the first place
// (ai-instruct-78).
const ConnectionFields = ({ draft, setDraftField, onServiceAccountFile, fileError, showFlatDirectory }) => {
  const { t } = useI18n();

  return (
    <>
      {draft.type === "s3" && (
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
      )}

      {draft.type === "git" && (
        <>
          <div className="cw-field">
            <label className="cw-label">
              {t("organizations.repoUrl")}
              <RequiredMark />
            </label>
            <input
              className="cw-input"
              type="text"
              value={draft.repoUrl}
              onChange={(e) => setDraftField("repoUrl", e.target.value)}
            />
          </div>

          <div className="cw-field">
            <label className="cw-label">{t("organizations.gitAuthMethod")}</label>
            <div className={styles.authMethodChoice}>
              <label>
                <input
                  type="radio"
                  name="gitAuthMethod"
                  checked={draft.gitAuthMethod === "password"}
                  onChange={() => setDraftField("gitAuthMethod", "password")}
                />
                {t("organizations.gitAuthMethodPassword")}
              </label>
              <label>
                <input
                  type="radio"
                  name="gitAuthMethod"
                  checked={draft.gitAuthMethod === "sshKey"}
                  onChange={() => setDraftField("gitAuthMethod", "sshKey")}
                />
                {t("organizations.gitAuthMethodSshKey")}
              </label>
            </div>
          </div>

          {draft.gitAuthMethod === "password" ? (
            <>
              <div className="cw-field">
                <label className="cw-label">
                  {t("organizations.gitUsername")}
                  <RequiredMark />
                </label>
                <input
                  className="cw-input"
                  type="text"
                  value={draft.username}
                  onChange={(e) => setDraftField("username", e.target.value)}
                />
              </div>
              <div className="cw-field">
                <label className="cw-label">
                  {t("organizations.gitPassword")}
                  <RequiredMark />
                </label>
                <input
                  className="cw-input"
                  type="password"
                  value={draft.password}
                  onChange={(e) => setDraftField("password", e.target.value)}
                />
              </div>
            </>
          ) : (
            <>
              <div className="cw-field">
                <label className="cw-label">
                  {t("organizations.sshPrivateKey")}
                  <RequiredMark />
                </label>
                <textarea
                  className="cw-input"
                  rows={5}
                  value={draft.sshPrivateKey}
                  onChange={(e) => setDraftField("sshPrivateKey", e.target.value)}
                />
              </div>
              <div className="cw-field">
                <label className="cw-label">{t("organizations.sshPrivateKeyPassphrase")}</label>
                <input
                  className="cw-input"
                  type="password"
                  value={draft.sshPrivateKeyPassphrase}
                  onChange={(e) => setDraftField("sshPrivateKeyPassphrase", e.target.value)}
                />
              </div>
            </>
          )}
        </>
      )}

      {draft.type === "google_drive" && (
        <>
          <div className="cw-field">
            <label className="cw-label">
              {t("organizations.serviceAccount")}
              <RequiredMark />
            </label>
            <input className="cw-input" type="file" accept=".json" onChange={onServiceAccountFile} />
            {draft.serviceAccountFileName && <span className={styles.fileName}>{draft.serviceAccountFileName}</span>}
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

      <div className="cw-field">
        <label className="cw-label">{t("organizations.subfolderPath")}</label>
        <input
          className="cw-input"
          type="text"
          value={draft.path}
          placeholder={t("organizations.subfolderPathPlaceholder")}
          onChange={(e) => setDraftField("path", e.target.value)}
        />
        <p className={styles.hint}>{t("organizations.subfolderPathHint")}</p>
      </div>

      {showFlatDirectory && (
        <>
          <label className={`cw-checkbox ${styles.flatDirectoryField}`}>
            <input
              type="checkbox"
              checked={draft.flatDirectory}
              onChange={(e) => setDraftField("flatDirectory", e.target.checked)}
            />
            {t("organizations.flatDirectory")}
          </label>
          <p className={styles.hint}>{t("organizations.flatDirectoryHint")}</p>
        </>
      )}
    </>
  );
};

export default ConnectionFields;
