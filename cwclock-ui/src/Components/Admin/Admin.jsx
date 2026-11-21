import React, { useEffect, useRef, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaRegEdit, FaEye } from "react-icons/fa";
import { FiSearch } from "react-icons/fi";
import { MdDeleteForever, MdSecurity } from "react-icons/md";
import Tooltip from "../common/Tooltip";
import ConfirmModal from "../common/ConfirmModal";
import CopyIdButton from "../common/CopyIdButton";
import memberLabel from "../common/memberLabel";
import EditUserModal from "./EditUserModal";
import UserIpsModal from "./UserIpsModal";
import {
  listAllUsersApi,
  deleteUserApi,
  disableMfaApi,
  ADMIN_PAGE_SIZE,
} from "../../Redux/Admin/Admin.actions";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/Admin.module.css";

const roleBadgeClass = {
  superuser: styles.roleSuperuser,
  confirmed: styles.roleConfirmed,
  disabled: styles.roleDisabled,
  ban: styles.roleBan,
};

const FILTERS = ["all", "superuser", "confirmed", "disabled", "ban"];
const SEARCH_DEBOUNCE_MS = 400;

const Admin = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { users, usersTotal, isLoading } = useSelector((state) => state.admin);
  const [editingUser, setEditingUser] = useState(null);
  const [deletingUser, setDeletingUser] = useState(null);
  const [disablingMfaUser, setDisablingMfaUser] = useState(null);
  const [viewingIpsUser, setViewingIpsUser] = useState(null);

  // termInput is the raw, immediately-typed value; it's debounced into
  // query.term so the roster auto-reloads without a request per keystroke. The
  // whole filter (term + role) and the page live in one query object so a
  // filter change can reset the page atomically (startIndex back to 0).
  const [termInput, setTermInput] = useState("");
  const [query, setQuery] = useState({ term: "", role: "", startIndex: 0 });
  const sentinelRef = useRef(null);

  const hasMore = users.length < usersTotal;

  const filterLabel = (f) => (f === "all" ? t("admin.filterAll") : t(`common.globalRole${f.charAt(0).toUpperCase()}${f.slice(1)}`));
  const roleLabel = (role) => t(`common.globalRole${(role || "confirmed").charAt(0).toUpperCase()}${(role || "confirmed").slice(1)}`);

  // Debounce the search field into the query (resetting to the first page).
  useEffect(() => {
    const id = setTimeout(() => {
      setQuery((q) => (q.term === termInput ? q : { ...q, term: termInput, startIndex: 0 }));
    }, SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(id);
  }, [termInput]);

  // Load the page whenever the filter or the page changes.
  useEffect(() => {
    dispatch(
      listAllUsersApi(user.token, {
        term: query.term,
        role: query.role,
        startIndex: query.startIndex,
        maxResults: ADMIN_PAGE_SIZE,
      })
    );
  }, [dispatch, user.token, query]);

  // Infinite scroll: advance a page when the sentinel scrolls into view.
  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel || isLoading || !hasMore) return undefined;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          setQuery((q) => ({ ...q, startIndex: q.startIndex + 1 }));
        }
      },
      { rootMargin: "200px" }
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [isLoading, hasMore]);

  const activeRole = query.role || "all";
  const selectFilter = (f) => setQuery((q) => ({ ...q, role: f === "all" ? "" : f, startIndex: 0 }));

  const handleDelete = () => {
    dispatch(deleteUserApi(deletingUser.id, user.token));
    setDeletingUser(null);
  };

  const handleDisableMfa = () => {
    dispatch(disableMfaApi(disablingMfaUser.id, user.token));
    setDisablingMfaUser(null);
  };

  return (
    <div className={styles.main}>
      <h1 className="cw-title">{t("admin.usersTitle")}</h1>

      <div className={styles.toolbar}>
        <div className={styles.filters}>
          {FILTERS.map((f) => (
            <button
              key={f}
              type="button"
              className={`${styles.filterBtn} ${activeRole === f ? styles.filterBtnActive : ""}`}
              onClick={() => selectFilter(f)}
              title={t("admin.showFilterUsers", { filter: filterLabel(f) })}
            >
              {filterLabel(f)}
            </button>
          ))}
        </div>
        <div className={styles.search}>
          <FiSearch className={styles.searchIcon} aria-hidden="true" />
          <input
            type="search"
            className={styles.searchInput}
            value={termInput}
            onChange={(e) => setTermInput(e.target.value)}
            placeholder={t("admin.searchPlaceholder")}
            aria-label={t("admin.searchPlaceholder")}
          />
        </div>
      </div>

      {!isLoading && users.length === 0 && <p className={styles.empty}>{t("admin.noUsersMatch")}</p>}

      <ul className={`cw-list ${styles.userList}`}>
        {users.map((u) => (
          <li className={`cw-list-item ${styles.userRow}`} key={u.id}>
            <span className={styles.denomination}>{memberLabel(u)}</span>
            <span className={styles.email}>{u.email}</span>
            <span className={`${styles.roleBadge} ${roleBadgeClass[u.role] || ""}`}>{roleLabel(u.role)}</span>
            <div className={styles.rowActions}>
              <CopyIdButton id={u.id} className={styles.iconBtn} />
              <Tooltip label={t("admin.viewIps")}>
                <button type="button" className={styles.iconBtn} onClick={() => setViewingIpsUser(u)}>
                  <FaEye style={{ fontSize: "17px" }} />
                </button>
              </Tooltip>
              {u.mfaEnabled && (
                <Tooltip label={t("admin.disableMfa")}>
                  <button type="button" className={styles.iconBtn} onClick={() => setDisablingMfaUser(u)}>
                    <MdSecurity style={{ fontSize: "18px" }} />
                  </button>
                </Tooltip>
              )}
              <Tooltip label={t("common.edit")}>
                <button type="button" className={styles.iconBtn} onClick={() => setEditingUser(u)}>
                  <FaRegEdit style={{ fontSize: "18px" }} />
                </button>
              </Tooltip>
              <Tooltip label={t("common.delete")}>
                <button
                  type="button"
                  className={`${styles.iconBtn} ${styles.iconBtnDanger}`}
                  onClick={() => setDeletingUser(u)}
                  disabled={u.id === user.id}
                >
                  <MdDeleteForever style={{ fontSize: "20px" }} />
                </button>
              </Tooltip>
            </div>
          </li>
        ))}
      </ul>

      {isLoading && <div className={styles.loadingMore}><span className={styles.spinner} /></div>}
      {hasMore && <div ref={sentinelRef} className={styles.sentinel} />}

      <EditUserModal
        show={!!editingUser}
        onClose={() => setEditingUser(null)}
        targetUser={editingUser}
        token={user.token}
      />

      <UserIpsModal
        show={!!viewingIpsUser}
        onClose={() => setViewingIpsUser(null)}
        targetUser={viewingIpsUser}
        token={user.token}
      />

      <ConfirmModal
        show={!!deletingUser}
        title={t("admin.deleteUserTitle")}
        body={deletingUser ? t("admin.deleteUserBody", { email: deletingUser.email }) : ""}
        confirmLabel={t("common.delete")}
        onConfirm={handleDelete}
        onCancel={() => setDeletingUser(null)}
      />

      <ConfirmModal
        show={!!disablingMfaUser}
        title={t("admin.disableMfaTitle")}
        body={disablingMfaUser ? t("admin.disableMfaBody", { email: disablingMfaUser.email }) : ""}
        confirmLabel={t("admin.disableMfa")}
        onConfirm={handleDisableMfa}
        onCancel={() => setDisablingMfaUser(null)}
      />
    </div>
  );
};

export default Admin;
