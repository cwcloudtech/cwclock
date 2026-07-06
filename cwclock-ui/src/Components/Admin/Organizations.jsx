import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaRegEdit, FaExchangeAlt } from "react-icons/fa";
import { MdDeleteForever } from "react-icons/md";
import Tooltip from "../common/Tooltip";
import ConfirmModal from "../common/ConfirmModal";
import EditOrgModal from "./EditOrgModal";
import TransferOwnershipModal from "./TransferOwnershipModal";
import { listAllOrganizationsApi } from "../../Redux/Admin/Admin.actions";
import { deleteOrgApi } from "../../Redux/Organizations/Org.actions";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/Admin.module.css";

const Organizations = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { organizations } = useSelector((state) => state.admin);
  const [editingOrg, setEditingOrg] = useState(null);
  const [transferringOrg, setTransferringOrg] = useState(null);
  const [deletingOrg, setDeletingOrg] = useState(null);

  useEffect(() => {
    dispatch(listAllOrganizationsApi(user.token));
  }, [dispatch, user.token]);

  const handleDelete = async () => {
    const orgId = deletingOrg.id;
    setDeletingOrg(null);
    await dispatch(deleteOrgApi(orgId, user.token));
    dispatch(listAllOrganizationsApi(user.token));
  };

  return (
    <div className={styles.main}>
      <h1 className="cw-title">{t("admin.organizationsTitle")}</h1>

      <ul className={`cw-list ${styles.userList}`}>
        {organizations.map((org) => (
          <li className={`cw-list-item ${styles.orgRow}`} key={org.id}>
            <span className={styles.denomination}>{org.name}</span>
            <span className={styles.email}>{org.ownerEmail}</span>
            <span />
            <div className={styles.rowActions}>
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

      <EditOrgModal
        show={!!editingOrg}
        onClose={() => setEditingOrg(null)}
        targetOrg={editingOrg}
        token={user.token}
      />

      <TransferOwnershipModal
        show={!!transferringOrg}
        onClose={() => setTransferringOrg(null)}
        targetOrg={transferringOrg}
        token={user.token}
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
