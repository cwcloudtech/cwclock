import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaRegEdit } from "react-icons/fa";
import { MdDeleteForever } from "react-icons/md";
import Tooltip from "../common/Tooltip";
import ConfirmModal from "../common/ConfirmModal";
import memberLabel from "../common/memberLabel";
import EditUserModal from "./EditUserModal";
import { listAllUsersApi, deleteUserApi } from "../../Redux/Admin/Admin.actions";
import styles from "./Styles/Admin.module.css";

const roleBadgeClass = {
  superuser: styles.roleSuperuser,
  confirmed: styles.roleConfirmed,
  disabled: styles.roleDisabled,
};

const FILTERS = ["all", "superuser", "confirmed", "disabled"];

const Admin = () => {
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { users } = useSelector((state) => state.admin);
  const [editingUser, setEditingUser] = useState(null);
  const [deletingUser, setDeletingUser] = useState(null);
  const [filter, setFilter] = useState("all");

  useEffect(() => {
    dispatch(listAllUsersApi(user.token));
  }, [dispatch, user.token]);

  const handleDelete = () => {
    dispatch(deleteUserApi(deletingUser.id, user.token));
    setDeletingUser(null);
  };

  const visibleUsers = users.filter((u) => filter === "all" || (u.role || "confirmed") === filter);

  return (
    <div className={styles.main}>
      <h1 className="cw-title">Users</h1>

      <div className={styles.filters}>
        {FILTERS.map((f) => (
          <button
            key={f}
            type="button"
            className={`${styles.filterBtn} ${filter === f ? styles.filterBtnActive : ""}`}
            onClick={() => setFilter(f)}
            title={`Show ${f} users`}
          >
            {f}
          </button>
        ))}
      </div>

      <ul className={`cw-list ${styles.userList}`}>
        {visibleUsers.map((u) => (
          <li className={`cw-list-item ${styles.userRow}`} key={u.id}>
            <span className={styles.denomination}>{memberLabel(u)}</span>
            <span className={styles.email}>{u.email}</span>
            <span className={`${styles.roleBadge} ${roleBadgeClass[u.role] || ""}`}>{u.role || "confirmed"}</span>
            <div className={styles.rowActions}>
              <Tooltip label="Edit">
                <button type="button" className={styles.iconBtn} onClick={() => setEditingUser(u)}>
                  <FaRegEdit style={{ fontSize: "18px" }} />
                </button>
              </Tooltip>
              <Tooltip label="Delete">
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
        title="Delete user"
        body={deletingUser ? `Delete "${deletingUser.email}"? This cannot be undone.` : ""}
        confirmLabel="Delete"
        onConfirm={handleDelete}
        onCancel={() => setDeletingUser(null)}
      />
    </div>
  );
};

export default Admin;
