import React, { useEffect, useRef, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaRegEdit, FaExchangeAlt } from "react-icons/fa";
import { FiSearch } from "react-icons/fi";
import { MdDeleteForever } from "react-icons/md";
import Tooltip from "../common/Tooltip";
import ConfirmModal from "../common/ConfirmModal";
import CopyIdButton from "../common/CopyIdButton";
import EditOrgModal from "./EditOrgModal";
import TransferOwnershipModal from "./TransferOwnershipModal";
import { listAllOrganizationsApi, ADMIN_PAGE_SIZE } from "../../Redux/Admin/Admin.actions";
import { deleteOrgApi } from "../../Redux/Organizations/Org.actions";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/Admin.module.css";

const SEARCH_DEBOUNCE_MS = 400;

const Organizations = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { organizations, organizationsTotal, isLoading } = useSelector((state) => state.admin);
  const [editingOrg, setEditingOrg] = useState(null);
  const [transferringOrg, setTransferringOrg] = useState(null);
  const [deletingOrg, setDeletingOrg] = useState(null);

  // termInput is debounced into query.term (resetting to the first page); the
  // term + page live in one query object so a filter change resets the page
  // atomically. See Admin.jsx (users) for the same pattern.
  const [termInput, setTermInput] = useState("");
  const [query, setQuery] = useState({ term: "", startIndex: 0 });
  const sentinelRef = useRef(null);

  const hasMore = organizations.length < organizationsTotal;

  useEffect(() => {
    const id = setTimeout(() => {
      setQuery((q) => (q.term === termInput ? q : { ...q, term: termInput, startIndex: 0 }));
    }, SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(id);
  }, [termInput]);

  useEffect(() => {
    dispatch(
      listAllOrganizationsApi(user.token, {
        term: query.term,
        startIndex: query.startIndex,
        maxResults: ADMIN_PAGE_SIZE,
      })
    );
  }, [dispatch, user.token, query]);

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

  // Reload the paged list from the top (preserving the current term), keeping the
  // infinite-scroll page state in sync with the slice after a mutation.
  const reloadFromTop = () => setQuery((q) => ({ ...q, startIndex: 0 }));

  const handleDelete = async () => {
    const orgId = deletingOrg.id;
    setDeletingOrg(null);
    await dispatch(deleteOrgApi(orgId, user.token));
    reloadFromTop();
  };

  return (
    <div className={styles.main}>
      <h1 className="cw-title">{t("admin.organizationsTitle")}</h1>

      <div className={styles.toolbar}>
        <div className={styles.search}>
          <FiSearch className={styles.searchIcon} aria-hidden="true" />
          <input
            type="search"
            className={styles.searchInput}
            value={termInput}
            onChange={(e) => setTermInput(e.target.value)}
            placeholder={t("admin.searchOrgsPlaceholder")}
            aria-label={t("admin.searchOrgsPlaceholder")}
          />
        </div>
      </div>

      {!isLoading && organizations.length === 0 && (
        <p className={styles.empty}>{t("admin.noOrgsMatch")}</p>
      )}

      <ul className={`cw-list ${styles.userList}`}>
        {organizations.map((org) => (
          <li className={`cw-list-item ${styles.orgRow}`} key={org.id}>
            <span className={styles.denomination}>{org.name}</span>
            <span className={styles.email}>{org.ownerEmail}</span>
            <span />
            <div className={styles.rowActions}>
              <CopyIdButton id={org.id} className={styles.iconBtn} />
              <Tooltip label={t("organizations.transferOwnership")}>
                <button
                  type="button"
                  className={styles.iconBtn}
                  onClick={() => setTransferringOrg(org)}
                >
                  <FaExchangeAlt style={{ fontSize: "16px" }} />
                </button>
              </Tooltip>
              <Tooltip label={t("common.edit")}>
                <button type="button" className={styles.iconBtn} onClick={() => setEditingOrg(org)}>
                  <FaRegEdit style={{ fontSize: "18px" }} />
                </button>
              </Tooltip>
              <Tooltip label={t("common.delete")}>
                <button
                  type="button"
                  className={`${styles.iconBtn} ${styles.iconBtnDanger}`}
                  onClick={() => setDeletingOrg(org)}
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

      <EditOrgModal
        show={!!editingOrg}
        onClose={() => setEditingOrg(null)}
        targetOrg={editingOrg}
        token={user.token}
        onSaved={reloadFromTop}
      />

      <TransferOwnershipModal
        show={!!transferringOrg}
        onClose={() => setTransferringOrg(null)}
        targetOrg={transferringOrg}
        token={user.token}
        onSaved={reloadFromTop}
      />

      <ConfirmModal
        show={!!deletingOrg}
        title={t("admin.deleteOrgTitle")}
        body={deletingOrg ? t("admin.deleteOrgBody", { name: deletingOrg.name }) : ""}
        confirmLabel={t("common.delete")}
        onConfirm={handleDelete}
        onCancel={() => setDeletingOrg(null)}
      />
    </div>
  );
};

export default Organizations;
