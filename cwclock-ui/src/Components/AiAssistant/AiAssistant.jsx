import React, { useEffect, useRef, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FiSend, FiPlus, FiTrash2, FiMessageSquare } from "react-icons/fi";
import { toast } from "react-toastify";
import Markdown from "../common/Markdown";
import SingleSelect from "../common/SingleSelect";
import Tooltip from "../common/Tooltip";
import ConfirmModal from "../common/ConfirmModal";
import EmptyState from "../common/EmptyState";
import {
  listConnectionsApi,
  listConversationsApi,
  getConversationApi,
  sendChatApi,
  deleteConversationsApi,
} from "../../Redux/AI/AI.actions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import toastOptions from "../../Redux/toastOptions";
import styles from "./Styles/AiAssistant.module.css";

const AiAssistant = () => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId } = useSelector((state) => state.organizations);
  const { connections, conversations, isSending } = useSelector((state) => state.ai);

  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState("");
  const [conversationId, setConversationId] = useState(null);
  const [connectionId, setConnectionId] = useState("");
  const [confirmClear, setConfirmClear] = useState(false);
  const scrollRef = useRef(null);

  useEffect(() => {
    dispatch(listConnectionsApi(user.token));
    dispatch(listConversationsApi(user.token));
  }, [dispatch, user.token]);

  // Default to the first (default) subscription until the user picks another.
  useEffect(() => {
    if (!connectionId && connections.length > 0) {
      setConnectionId(connections[0].id);
    }
  }, [connections, connectionId]);

  useEffect(() => {
    const el = scrollRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [messages, isSending]);

  const newChat = () => {
    setMessages([]);
    setConversationId(null);
  };

  const openConversation = async (id) => {
    try {
      const conv = await dispatch(getConversationApi(id, user.token));
      const loaded = [];
      (conv.exchanges || []).forEach((ex) => {
        loaded.push({ role: "user", content: ex.question });
        loaded.push({ role: "assistant", content: ex.answer });
      });
      setMessages(loaded);
      setConversationId(conv.id);
    } catch (err) {
      toast.error(apiErrorMessage(err, locale), toastOptions);
    }
  };

  const handleSend = async (e) => {
    e.preventDefault();
    const text = input.trim();
    if (!text || isSending) return;
    setInput("");
    setMessages((prev) => [...prev, { role: "user", content: text }]);
    try {
      const res = await dispatch(
        sendChatApi({ message: text, conversationId, connectionId, orgId: currentOrgId }, user.token)
      );
      setMessages((prev) => [...prev, { role: "assistant", content: res.message }]);
      if (res.conversationId) setConversationId(res.conversationId);
      dispatch(listConversationsApi(user.token));
    } catch (err) {
      const message = apiErrorMessage(err, locale);
      toast.error(message, toastOptions);
      setMessages((prev) => [...prev, { role: "error", content: message }]);
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend(e);
    }
  };

  const handleClearAll = async () => {
    setConfirmClear(false);
    try {
      await dispatch(deleteConversationsApi({ all: true }, user.token));
      newChat();
    } catch (err) {
      /* toast already shown */
    }
  };

  const connectionOptions = connections.map((c) => ({
    value: c.id,
    label: c.name || c.provider,
  }));

  if (connections.length === 0) {
    return (
      <div className={styles.main}>
        <EmptyState title={t("ai.chatNoSubscriptionTitle")} body={t("ai.chatNoSubscriptionBody")} />
      </div>
    );
  }

  return (
    <div className={styles.main}>
      <aside className={styles.sidebar}>
        <button type="button" className={styles.newChatBtn} onClick={newChat}>
          <FiPlus /> {t("ai.newChat")}
        </button>
        <div className={styles.conversationList}>
          {conversations.length === 0 ? (
            <p className={styles.sidebarEmpty}>{t("ai.noConversations")}</p>
          ) : (
            conversations.map((c) => (
              <button
                key={c.id}
                type="button"
                className={`${styles.conversationItem} ${c.id === conversationId ? styles.conversationActive : ""}`}
                onClick={() => openConversation(c.id)}
                title={c.title}
              >
                <FiMessageSquare className={styles.conversationIcon} />
                <span className={styles.conversationTitle}>{c.title}</span>
              </button>
            ))
          )}
        </div>
        {conversations.length > 0 && (
          <button type="button" className={styles.clearBtn} onClick={() => setConfirmClear(true)}>
            <FiTrash2 /> {t("ai.clearHistory")}
          </button>
        )}
      </aside>

      <section className={styles.chat}>
        <div className={styles.messages} ref={scrollRef}>
          {messages.length === 0 && (
            <div className={styles.welcome}>
              <h2 className={styles.welcomeTitle}>{t("ai.welcomeTitle")}</h2>
              <p className={styles.welcomeBody}>{t("ai.welcomeBody")}</p>
            </div>
          )}
          {messages.map((m, i) => (
            <div
              key={i}
              className={`${styles.message} ${
                m.role === "user"
                  ? styles.messageUser
                  : m.role === "error"
                  ? styles.messageError
                  : styles.messageAssistant
              }`}
            >
              {m.role === "assistant" ? <Markdown>{m.content}</Markdown> : <span>{m.content}</span>}
            </div>
          ))}
          {isSending && (
            <div className={`${styles.message} ${styles.messageAssistant}`}>
              <span className={styles.typing}>
                <span />
                <span />
                <span />
              </span>
            </div>
          )}
        </div>

        <form className={styles.composer} onSubmit={handleSend}>
          {connections.length > 1 && (
            <SingleSelect
              label={t("ai.model")}
              allLabel={connections[0] ? connections[0].name || connections[0].provider : ""}
              options={connectionOptions}
              value={connectionId}
              onChange={setConnectionId}
            />
          )}
          <textarea
            className={styles.textarea}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={t("ai.askPlaceholder")}
            rows={1}
          />
          <Tooltip label={t("ai.send")}>
            <button type="submit" className={styles.sendBtn} disabled={isSending || !input.trim()}>
              <FiSend />
            </button>
          </Tooltip>
        </form>
      </section>

      <ConfirmModal
        show={confirmClear}
        title={t("ai.clearHistoryTitle")}
        body={t("ai.clearHistoryBody")}
        confirmLabel={t("ai.clearHistory")}
        onConfirm={handleClearAll}
        onCancel={() => setConfirmClear(false)}
      />
    </div>
  );
};

export default AiAssistant;
