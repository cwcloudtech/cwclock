import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import Button from "../common/Button";
import memberLabel from "../common/memberLabel";
import EditUserModal from "./EditUserModal";
import { listAllUsersApi } from "../../Redux/Admin/Admin.actions";
import styles from "./Styles/Admin.module.css";

const roleBadgeClass = {
  superuser: styles.roleSuperuser,
  confirmed: styles.roleConfirmed,
  disabled: styles.roleDisabled,
};

const Admin = () => {
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { users } = useSelector((state) => state.admin);
  const [editingUser, setEditingUser] = useState(null);

  useEffect(() => {
    dispatch(listAllUsersApi(user.token));
  }, [dispatch, user.token]);

  return (
    <div className={styles.main}>
      <h1 className="cw-title">Users</h1>

      <ul className="cw-list">
        {users.map((u) => (
          <li className={`cw-list-item ${styles.userRow}`} key={u.id}>
            <div>
              <strong>{memberLabel(u)}</strong>
              {u.name && <span className={styles.email}> ({u.email})</span>}
              {!u.name && <span className={styles.email}>{u.email}</span>}
            </div>
            <span className={`${styles.roleBadge} ${roleBadgeClass[u.role] || ""}`}>{u.role || "confirmed"}</span>
            <Button size="sm" variant="secondary" onClick={() => setEditingUser(u)} title={`Edit ${u.email}`}>
              Edit
            </Button>
          </li>
        ))}
      </ul>

      <EditUserModal
        show={!!editingUser}
        onClose={() => setEditingUser(null)}
        targetUser={editingUser}
        token={user.token}
      />
    </div>
  );
};

export default Admin;
