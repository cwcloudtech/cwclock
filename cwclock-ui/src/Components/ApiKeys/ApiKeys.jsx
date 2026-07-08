import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { MdDeleteForever } from "react-icons/md";
import { FaRegCopy } from "react-icons/fa";
import { toast } from "react-toastify";
import ConfigForm from "../common/ConfigForm";
import CollapsiblePanel from "../common/CollapsiblePanel";
import ConfirmModal from "../common/ConfirmModal";
import Modal from "../common/Modal";
import Tooltip from "../common/Tooltip";
import EmptyState from "../common/EmptyState";
import { listApiKeysApi, createApiKeyApi, deleteApiKeyApi } from "../../Redux/ApiKeys/ApiKey.actions";
import { useI18n } from "../../i18n/I18nContext";
import toastOptions from "../../Redux/toastOptions";
import styles from "./Styles/ApiKeys.module.css";

const initialFields = { description: "", expiresAt: "" };

const ApiKeys = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { keys } = useSelector((state) => state.apiKeys);
  const [fields, setFields] = useState(initialFields);
  const [deletingKey, setDeletingKey] = useState(null);
  const [createdToken, setCreatedToken] = useState(null);

  useEffect(() => {
    dispatch(listApiKeysApi(user.token));
  }, [dispatch, user.token]);

  const setField = (key, value) => setFields({ ...fields, [key]: value });

  const formConfig = {
    name: "ApiKey",
    fields: [
      { name: "description", type: "text", label: t("apiKeys.description"), required: true },
      { name: "expiresAt", type: "date", label: t("apiKeys.expiresAt") },
    ],
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!fields.description) return;
    const payload = {
      description: fields.description,
      expiresAt: fields.expiresAt ? `${fields.expiresAt}T23:59:59.000Z` : null,
    };
    try {
      const created = await dispatch(createApiKeyApi(payload, user.token));
      setCreatedToken(created.token);
      setFields(initialFields);
    } catch (e) {
      // error toast already shown by createApiKeyApi
    }
  };

  const handleDelete = () => {
    dispatch(deleteApiKeyApi(deletingKey.id, user.token));
    setDeletingKey(null);
  };

  const handleCopy = (text) => {
    navigator.clipboard.writeText(text).then(() => {
      toast.success(t("common.copied"), toastOptions);
    });
  };

  return (
    <div className={styles.main}>
      <h1 className="cw-title">{t("apiKeys.title")}</h1>
      <p className={styles.intro}>{t("apiKeys.intro")}</p>

      <CollapsiblePanel title={t("apiKeys.createKey")}>
        <ConfigForm config={formConfig} values={fields} onChange={setField} onSubmit={handleSubmit} submitLabel={t("common.create")} />
      </CollapsiblePanel>

      {keys.length === 0 && <EmptyState title={t("apiKeys.emptyTitle")} body={t("apiKeys.emptyBody")} />}

      <ul className="cw-list">
        {keys.map((key) => (
          <li className={`cw-list-item ${styles.keyRow}`} key={key.id}>
            <strong className={styles.description}>{key.description}</strong>
            <span className={styles.expiresAt}>
              {key.expiresAt ? t("apiKeys.expiresOn", { date: new Date(key.expiresAt).toLocaleDateString() }) : t("apiKeys.neverExpires")}
            </span>
            <div className={styles.rowActions}>
              <Tooltip label={t("common.delete")}>
                <button type="button" className={styles.iconBtn} onClick={() => setDeletingKey(key)}>
                  <MdDeleteForever style={{ fontSize: "20px" }} />
                </button>
              </Tooltip>
            </div>
          </li>
        ))}
      </ul>

      <Modal show={!!createdToken} title={t("apiKeys.createdTitle")} onClose={() => setCreatedToken(null)}>
        <p className={styles.warning}>{t("apiKeys.createdWarning")}</p>
        <div className={styles.tokenBox}>
          <code className={styles.token}>{createdToken}</code>
          <Tooltip label={t("common.copy")}>
            <button type="button" className={styles.iconBtn} onClick={() => handleCopy(createdToken)}>
              <FaRegCopy style={{ fontSize: "16px" }} />
            </button>
          </Tooltip>
        </div>
      </Modal>

      <ConfirmModal
        show={!!deletingKey}
        title={t("apiKeys.deleteKeyTitle")}
        body={deletingKey ? t("apiKeys.deleteKeyBody", { description: deletingKey.description }) : ""}
        confirmLabel={t("common.delete")}
        onConfirm={handleDelete}
        onCancel={() => setDeletingKey(null)}
      />
    </div>
  );
};

export default ApiKeys;
