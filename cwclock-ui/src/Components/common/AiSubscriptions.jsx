import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { MdDeleteForever, MdEdit, MdStar, MdStarBorder } from "react-icons/md";
import { FiShare2 } from "react-icons/fi";
import Button from "./Button";
import Modal from "./Modal";
import SingleSelect from "./SingleSelect";
import ConfirmModal from "./ConfirmModal";
import Tooltip from "./Tooltip";
import {
  listProvidersApi,
  listConnectionsApi,
  createConnectionApi,
  updateConnectionApi,
  deleteConnectionApi,
  reorderConnectionsApi,
  listSharesApi,
  shareConnectionApi,
  unshareConnectionApi,
} from "../../Redux/AI/AI.actions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import memberLabel from "./memberLabel";
import styles from "./Styles/AiSubscriptions.module.css";

const emptyForm = { provider: "", name: "", token: "", model: "", baseUrl: "", maxTokens: "" };

// AiSubscriptions is the self-service "AI subscriptions" section shown in
// EditProfileModal (ai-instruct-132): add a connection to an AI provider with your
// own key, share it with other users, and pick which one is the default. Once at
// least one exists, the assistant chat appears in the sidebar.
const AiSubscriptions = ({ token }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { providers, connections } = useSelector((state) => state.ai);

  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState(null); // connection being edited, or null
  const [form, setForm] = useState(emptyForm);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [deleting, setDeleting] = useState(null);

  // sharing modal state
  const [sharing, setSharing] = useState(null); // connection whose shares are open
  const [shares, setShares] = useState([]);
  const [shareEmail, setShareEmail] = useState("");
  const [shareBusy, setShareBusy] = useState(false);

  useEffect(() => {
    dispatch(listProvidersApi(token));
    dispatch(listConnectionsApi(token));
  }, [dispatch, token]);

  const selectedProvider = providers.find((p) => p.name === form.provider);
  const providerOptions = providers.map((p) => ({ value: p.name, label: p.label }));

  const openAdd = () => {
    setEditing(null);
    setForm(emptyForm);
    setError("");
    setShowForm(true);
  };

  const openEdit = (c) => {
    setEditing(c);
    setForm({
      provider: c.provider || "",
      name: c.name || "",
      token: "",
      model: c.model || "",
      baseUrl: c.baseUrl || "",
      maxTokens: c.maxTokens ? String(c.maxTokens) : "",
    });
    setError("");
    setShowForm(true);
  };

  const setField = (key, value) => setForm((f) => ({ ...f, [key]: value }));

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    if (!form.provider) {
      setError(t("ai.pickProviderError"));
      return;
    }
    if (!editing && !form.token.trim()) {
      setError(t("ai.keyRequired"));
      return;
    }
    const payload = {
      provider: form.provider,
      name: form.name.trim(),
      token: form.token.trim(),
      model: form.model.trim(),
      baseUrl: form.baseUrl.trim(),
      maxTokens: form.maxTokens ? Number(form.maxTokens) : 0,
    };
    setBusy(true);
    try {
      if (editing) {
        await dispatch(updateConnectionApi(editing.id, payload, token));
      } else {
        await dispatch(createConnectionApi(payload, token));
      }
      setShowForm(false);
      setForm(emptyForm);
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    } finally {
      setBusy(false);
    }
  };

  const handleDelete = async () => {
    const c = deleting;
    setDeleting(null);
    try {
      await dispatch(deleteConnectionApi(c.id, token));
    } catch (err) {
      /* toast already shown */
    }
  };

  const handleSetDefault = async (c) => {
    try {
      await dispatch(reorderConnectionsApi([c.id], token));
    } catch (err) {
      /* toast already shown */
    }
  };

  const openShares = async (c) => {
    setSharing(c);
    setShareEmail("");
    try {
      setShares(await dispatch(listSharesApi(c.id, token)));
    } catch (err) {
      setShares([]);
    }
  };

  const handleAddShare = async (e) => {
    e.preventDefault();
    if (!shareEmail.trim()) return;
    setShareBusy(true);
    try {
      const updated = await dispatch(shareConnectionApi(sharing.id, shareEmail.trim(), token));
      setShares(updated);
      setShareEmail("");
    } catch (err) {
      /* toast already shown */
    } finally {
      setShareBusy(false);
    }
  };

  const handleRemoveShare = async (userId) => {
    try {
      await dispatch(unshareConnectionApi(sharing.id, userId, token));
      setShares((prev) => prev.filter((s) => s.userId !== userId));
    } catch (err) {
      /* toast already shown */
    }
  };

  const owned = (c) => c.ownerId === user.id || c.role === "owner";

  return (
    <div className={styles.container}>
      <div className={styles.headingRow}>
        <h3 className={styles.heading}>{t("ai.title")}</h3>
        <Button type="button" size="sm" onClick={openAdd} title={t("ai.addSubscription")}>
          {t("ai.addSubscription")}
        </Button>
      </div>
      <p className={styles.intro}>{t("ai.intro")}</p>

      {connections.length === 0 ? (
        <p className={styles.empty}>{t("ai.noSubscriptions")}</p>
      ) : (
        <ul className={styles.list}>
          {connections.map((c, index) => (
            <li key={c.id} className={styles.row}>
              <div className={styles.rowInfo}>
                <span className={styles.rowName}>
                  {c.name || c.provider}
                  {index === 0 && <span className={styles.defaultBadge}>{t("ai.default")}</span>}
                  {!owned(c) && <span className={styles.sharedBadge}>{t("ai.shared")}</span>}
                </span>
                <span className={styles.rowMeta}>
                  {c.provider}
                  {c.model ? ` · ${c.model}` : ""}
                </span>
              </div>
              <div className={styles.rowActions}>
                {index !== 0 && (
                  <Tooltip label={t("ai.setDefault")}>
                    <button type="button" className={styles.iconBtn} onClick={() => handleSetDefault(c)}>
                      <MdStarBorder style={{ fontSize: "18px" }} />
                    </button>
                  </Tooltip>
                )}
                {index === 0 && <MdStar className={styles.defaultStar} style={{ fontSize: "18px" }} />}
                {owned(c) && (
                  <>
                    <Tooltip label={t("ai.share")}>
                      <button type="button" className={styles.iconBtn} onClick={() => openShares(c)}>
                        <FiShare2 style={{ fontSize: "16px" }} />
                      </button>
                    </Tooltip>
                    <Tooltip label={t("common.edit")}>
                      <button type="button" className={styles.iconBtn} onClick={() => openEdit(c)}>
                        <MdEdit style={{ fontSize: "18px" }} />
                      </button>
                    </Tooltip>
                    <Tooltip label={t("common.delete")}>
                      <button type="button" className={styles.iconBtnDanger} onClick={() => setDeleting(c)}>
                        <MdDeleteForever style={{ fontSize: "18px" }} />
                      </button>
                    </Tooltip>
                  </>
                )}
              </div>
            </li>
          ))}
        </ul>
      )}

      {/* add / edit form */}
      <Modal
        show={showForm}
        title={editing ? t("ai.editSubscription") : t("ai.addSubscription")}
        onClose={() => setShowForm(false)}
      >
        <form onSubmit={handleSubmit}>
          <div className="cw-field">
            <label className="cw-label">{t("ai.provider")}</label>
            <SingleSelect
              label={t("ai.provider")}
              allLabel={t("ai.pickProvider")}
              options={providerOptions}
              value={form.provider}
              onChange={(v) => setField("provider", v)}
            />
          </div>
          <div className="cw-field">
            <label className="cw-label">{t("ai.name")}</label>
            <input
              className="cw-input"
              type="text"
              value={form.name}
              onChange={(e) => setField("name", e.target.value)}
              placeholder={t("ai.namePlaceholder")}
            />
          </div>
          <div className="cw-field">
            <label className="cw-label">{t("ai.apiKey")}</label>
            <input
              className="cw-input"
              type="password"
              autoComplete="off"
              value={form.token}
              onChange={(e) => setField("token", e.target.value)}
              placeholder={
                editing
                  ? t("ai.keyLeaveBlank")
                  : selectedProvider?.keyHint || t("ai.apiKey")
              }
            />
          </div>
          <div className="cw-field">
            <label className="cw-label">{t("ai.model")}</label>
            <input
              className="cw-input"
              type="text"
              value={form.model}
              onChange={(e) => setField("model", e.target.value)}
              placeholder={selectedProvider?.defaultModel || t("ai.modelPlaceholder")}
            />
          </div>
          <div className="cw-field">
            <label className="cw-label">{t("ai.baseUrl")}</label>
            <input
              className="cw-input"
              type="text"
              value={form.baseUrl}
              onChange={(e) => setField("baseUrl", e.target.value)}
              placeholder={selectedProvider?.defaultUrl || "https://..."}
            />
          </div>
          <div className="cw-field">
            <label className="cw-label">{t("ai.maxTokens")}</label>
            <input
              className="cw-input"
              type="number"
              min="1"
              value={form.maxTokens}
              onChange={(e) => setField("maxTokens", e.target.value)}
              placeholder={t("ai.maxTokensPlaceholder")}
            />
          </div>
          {error && <p className="cw-error">{error}</p>}
          <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 16 }}>
            <Button type="button" variant="secondary" onClick={() => setShowForm(false)}>
              {t("common.cancel")}
            </Button>
            <Button type="submit" disabled={busy}>
              {editing ? t("common.save") : t("ai.addSubscription")}
            </Button>
          </div>
        </form>
      </Modal>

      {/* sharing */}
      <Modal show={!!sharing} title={t("ai.shareTitle")} onClose={() => setSharing(null)}>
        <p className={styles.intro}>{t("ai.shareIntro")}</p>
        <form onSubmit={handleAddShare} className={styles.shareForm}>
          <input
            className="cw-input"
            type="email"
            value={shareEmail}
            onChange={(e) => setShareEmail(e.target.value)}
            placeholder={t("ai.shareEmailPlaceholder")}
          />
          <Button type="submit" size="sm" disabled={shareBusy || !shareEmail.trim()}>
            {t("ai.shareAdd")}
          </Button>
        </form>
        {shares.length === 0 ? (
          <p className={styles.empty}>{t("ai.shareEmpty")}</p>
        ) : (
          <ul className={styles.shareList}>
            {shares.map((s) => (
              <li key={s.userId} className={styles.shareRow}>
                <span className={styles.shareWho}>{memberLabel(s) || s.email}</span>
                <Tooltip label={t("ai.shareRemove")}>
                  <button type="button" className={styles.iconBtnDanger} onClick={() => handleRemoveShare(s.userId)}>
                    <MdDeleteForever style={{ fontSize: "18px" }} />
                  </button>
                </Tooltip>
              </li>
            ))}
          </ul>
        )}
      </Modal>

      <ConfirmModal
        show={!!deleting}
        title={t("ai.deleteTitle")}
        body={deleting ? t("ai.deleteBody", { name: deleting.name || deleting.provider }) : ""}
        confirmLabel={t("common.delete")}
        onConfirm={handleDelete}
        onCancel={() => setDeleting(null)}
      />
    </div>
  );
};

export default AiSubscriptions;
