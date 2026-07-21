import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaRegEdit } from "react-icons/fa";
import { MdDeleteForever, MdSecurity } from "react-icons/md";
import Tooltip from "../common/Tooltip";
import ConfirmModal from "../common/ConfirmModal";
import CopyIdButton from "../common/CopyIdButton";
import memberLabel from "../common/memberLabel";
import EditUserModal from "./EditUserModal";
import { listAllUsersApi, deleteUserApi, disableMfaApi } from "../../Redux/Admin/Admin.actions";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/Admin.module.css";

const roleBadgeClass = {
  superuser: styles.roleSuperuser,
  confirmed: styles.roleConfirmed,
  disabled: styles.roleDisabled,
  ban: styles.roleBan,
};

const FILTERS = ["all", "superuser", "confirmed", "disabled", "ban"];

const Admin = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { users } = useSelector((state) => state.admin);
  const [editingUser, setEditingUser] = useState(null);
  const [deletingUser, setDeletingUser] = useState(null);
  const [disablingMfaUser, setDisablingMfaUser] = useState(null);
  const [filter, setFilter] = useState("all");

  const filterLabel = (f) => (f === "all" ? t("admin.filterAll") : t(`common.globalRole${f.charAt(0).toUpperCase()}${f.slice(1)}`));
  const roleLabel = (role) => t(`common.globalRole${(role || "confirmed").charAt(0).toUpperCase()}${(role || "confirmed").slice(1)}`);

  useEffect(() => {
    dispatch(listAllUsersApi(user.token));
  }, [dispatch, user.token]);

  const handleDelete = () => {
    dispatch(deleteUserApi(deletingUser.id, user.token));
    setDeletingUser(null);
  };

  const handleDisableMfa = () => {
    dispatch(disableMfaApi(disablingMfaUser.id, user.token));
    setDisablingMfaUser(null);
  };

  const visibleUsers = users.filter((u) => filter === "all" || (u.role || "confirmed") === filter);

  return (
    <div className={styles.main}>
      <h1 className="cw-title">{t("admin.usersTitle")}</h1>

      <div className={styles.filters}>
        {FILTERS.map((f) => (
          <button
            key={f}
            type="button"
            className={`${styles.filterBtn} ${filter === f ? styles.filterBtnActive : ""}`}
            onClick={() => setFilter(f)}
            title={t("admin.showFilterUsers", { filter: filterLabel(f) })}
          >
            {filterLabel(f)}
          </button>
        ))}
      </div>

      <ul className={`cw-list ${styles.userList}`}>
        {visibleUsers.map((u) => (
          <li className={`cw-list-item ${styles.userRow}`} key={u.id}>
            <span className={styles.denomination}>{memberLabel(u)}</span>
            <span className={styles.email}>{u.email}</span>
            <span className={`${styles.roleBadge} ${roleBadgeClass[u.role] || ""}`}>{roleLabel(u.role)}</span>
            <div className={styles.rowActions}>
              <CopyIdButton id={u.id} className={styles.iconBtn} />
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

      <EditUserModal
        show={!!editingUser}
        onClose={() => setEditingUser(null)}
        targetUser={editingUser}
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
