import React, { useEffect, useRef, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FiSend, FiPlus, FiTrash2, FiMessageSquare, FiChevronLeft, FiChevronRight } from "react-icons/fi";
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

// The API sends an exchange's createdAt as an RFC3339 string (not epoch millis);
// render it in the user's locale, and fall back to "" on an absent/unparseable value.
const formatCreditDate = (value, locale) => {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleString(locale, { dateStyle: "medium", timeStyle: "short" });
};

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
  // What a confirmation is pending for: null, { type: "all" } (clear everything),
  // or { type: "one", id } (one conversation's trash icon). One modal serves both,
  // like uprodit.
  const [confirm, setConfirm] = useState(null);
  // How long a generated-file download link stays valid, as the API reports it
  // (null until a chat/conversation response tells us). A message older than this
  // has its download link shown as expired.
  const [downloadTtlSeconds, setDownloadTtlSeconds] = useState(null);
  // A 30s heartbeat so a link flips to "expired" on screen without needing a new
  // message to force a re-render.
  const [, setNowTick] = useState(0);
  // The conversations rail collapses to a thin bar (like uprodit), remembered
  // across reloads. localStorage can throw (private mode) — the toggle still works.
  const [historyCollapsed, setHistoryCollapsed] = useState(() => {
    try {
      return localStorage.getItem("cwclock.assistant.historyCollapsed") === "1";
    } catch {
      return false;
    }
  });
  const scrollRef = useRef(null);

  const toggleHistory = () =>
    setHistoryCollapsed((collapsed) => {
      const next = !collapsed;
      try {
        localStorage.setItem("cwclock.assistant.historyCollapsed", next ? "1" : "0");
      } catch {
        /* storage unavailable; the toggle still works, just isn't remembered */
      }
      return next;
    });

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

  useEffect(() => {
    const id = setInterval(() => setNowTick((n) => n + 1), 30000);
    return () => clearInterval(id);
  }, []);

  const newChat = () => {
    setMessages([]);
    setConversationId(null);
  };

  const openConversation = async (id) => {
    try {
      const conv = await dispatch(getConversationApi(id, user.token));
      if (conv.downloadTtlSeconds) setDownloadTtlSeconds(conv.downloadTtlSeconds);
      const loaded = [];
      (conv.exchanges || []).forEach((ex) => {
        loaded.push({ role: "user", content: ex.question });
        loaded.push({
          role: "assistant",
          content: ex.answer,
          model: ex.model,
          connection: ex.connection,
          createdAt: ex.createdAt,
        });
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
      if (res.downloadTtlSeconds) setDownloadTtlSeconds(res.downloadTtlSeconds);
      setMessages((prev) => [
        ...prev,
        { role: "assistant", content: res.message, model: res.model, createdAt: new Date().toISOString() },
      ]);
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

  // Runs the pending confirmation — clear-all, or a single conversation's trash icon.
  // deleteConversationsApi re-fetches the list itself on success.
  const runConfirm = async () => {
    const target = confirm;
    setConfirm(null);
    if (!target) return;
    try {
      if (target.type === "all") {
        await dispatch(deleteConversationsApi({ all: true }, user.token));
        newChat();
      } else {
        await dispatch(deleteConversationsApi({ ids: [target.id] }, user.token));
        // Deleting the open thread drops you back to a fresh chat.
        if (target.id === conversationId) newChat();
      }
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
      <aside className={`${styles.sidebar} ${historyCollapsed ? styles.sidebarCollapsed : ""}`}>
        <div className={styles.sidebarHeader}>
          <Tooltip label={historyCollapsed ? t("ai.showHistory") : t("ai.hideHistory")} position="right">
            <button
              type="button"
              className={styles.collapseBtn}
              onClick={toggleHistory}
              aria-label={historyCollapsed ? t("ai.showHistory") : t("ai.hideHistory")}
            >
              {historyCollapsed ? <FiChevronRight /> : <FiChevronLeft />}
            </button>
          </Tooltip>
          {!historyCollapsed && <span className={styles.sidebarTitle}>{t("ai.conversations")}</span>}
        </div>

        {/* New chat and Clear stay available when collapsed — icon-only with a
            tooltip, like the main sidebar and uprodit. */}
        <Tooltip label={historyCollapsed ? t("ai.newChat") : null} position="right">
          <button
            type="button"
            className={`${styles.newChatBtn} ${historyCollapsed ? styles.iconOnly : ""}`}
            onClick={newChat}
            aria-label={t("ai.newChat")}
          >
            <FiPlus />
            {!historyCollapsed && <span>{t("ai.newChat")}</span>}
          </button>
        </Tooltip>

        {!historyCollapsed && (
          <div className={styles.conversationList}>
            {conversations.length === 0 ? (
              <p className={styles.sidebarEmpty}>{t("ai.noConversations")}</p>
            ) : (
              conversations.map((c) => (
                <div
                  key={c.id}
                  className={`${styles.conversationRow} ${c.id === conversationId ? styles.conversationActive : ""}`}
                >
                  <button
                    type="button"
                    className={styles.conversationItem}
                    onClick={() => openConversation(c.id)}
                    title={c.title}
                  >
                    <FiMessageSquare className={styles.conversationIcon} />
                    <span className={styles.conversationTitle}>{c.title}</span>
                  </button>
                  <button
                    type="button"
                    className={styles.conversationDelete}
                    onClick={() => setConfirm({ type: "one", id: c.id })}
                    aria-label={t("ai.delete")}
                    title={t("ai.delete")}
                  >
                    <FiTrash2 />
                  </button>
                </div>
              ))
            )}
          </div>
        )}

        {conversations.length > 0 && (
          <Tooltip label={historyCollapsed ? t("ai.clearHistory") : null} position="right">
            <button
              type="button"
              className={`${styles.clearBtn} ${historyCollapsed ? styles.iconOnly : ""}`}
              onClick={() => setConfirm({ type: "all" })}
              aria-label={t("ai.clearHistory")}
            >
              <FiTrash2 />
              {!historyCollapsed && <span>{t("ai.clearHistory")}</span>}
            </button>
          </Tooltip>
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
              {m.role === "assistant" ? (
                <>
                  <Markdown
                    downloadExpiresAt={
                      downloadTtlSeconds && m.createdAt
                        ? new Date(m.createdAt).getTime() + downloadTtlSeconds * 1000
                        : null
                    }
                  >
                    {m.content}
                  </Markdown>
                  {(() => {
                    // Credit under each answer: which model / subscription produced it, and when
                    // (ported from uprodit). `model || connection` — the model, or the subscription
                    // name when the provider reported none.
                    const who = m.model || m.connection;
                    const date = formatCreditDate(m.createdAt, locale);
                    if (!who && !date) return null;
                    return (
                      <p className={styles.credit}>
                        {who ? `${t("ai.generatedBy")} ${who}${date ? " · " : ""}` : ""}
                        {date}
                      </p>
                    );
                  })()}
                </>
              ) : (
                <span>{m.content}</span>
              )}
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
          <textarea
            className={styles.textarea}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={t("ai.askPlaceholder")}
            rows={1}
          />
          {connections.length > 1 && (
            // Sits between the message field and Send. Opens upward (and
            // right-aligned) so its panel isn't clipped off the bottom of the
            // viewport by the composer pinned to the foot of the chat pane.
            <SingleSelect
              label={t("ai.model")}
              allLabel={connections[0] ? connections[0].name || connections[0].provider : ""}
              options={connectionOptions}
              value={connectionId}
              onChange={setConnectionId}
              direction="up"
              align="end"
            />
          )}
          <Tooltip label={t("ai.send")}>
            <button type="submit" className={styles.sendBtn} disabled={isSending || !input.trim()}>
              <FiSend />
            </button>
          </Tooltip>
        </form>
      </section>

      <ConfirmModal
        show={confirm !== null}
        title={confirm?.type === "all" ? t("ai.clearHistoryTitle") : t("ai.deleteConversationTitle")}
        body={confirm?.type === "all" ? t("ai.clearHistoryBody") : t("ai.deleteConversationBody")}
        confirmLabel={confirm?.type === "all" ? t("ai.clearHistory") : t("ai.delete")}
        onConfirm={runConfirm}
        onCancel={() => setConfirm(null)}
      />
    </div>
  );
};

export default AiAssistant;
