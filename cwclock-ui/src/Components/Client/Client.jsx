import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaRegEdit } from "react-icons/fa";
import { MdDeleteForever } from "react-icons/md";
import styles from "./Styles/Client.module.css";
import { listClientsApi, createClientApi, deleteClientApi } from "../../Redux/Clients/Client.actions";
import { listMembersApi } from "../../Redux/Organizations/Org.actions";
import ConfigForm from "../common/ConfigForm";
import CollapsiblePanel from "../common/CollapsiblePanel";
import Tooltip from "../common/Tooltip";
import ConfirmModal from "../common/ConfirmModal";
import EmptyState from "../common/EmptyState";
import CopyIdButton from "../common/CopyIdButton";
import EditClientModal from "./EditClientModal";
import { isAdminOrOwner as computeIsAdminOrOwner } from "../common/permissions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";

const initialFields = {
  name: "",
  email: "",
  address: "",
  postalCode: "",
  city: "",
  country: "",
  vatNumber: "",
  vatRate: "",
  vatDischargeMotive: "",
  purchaseOrder: "",
  hoursPerDay: "",
  dailyRate: "",
};

const Clients = () => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId } = useSelector((state) => state.organizations);
  const { clients } = useSelector((state) => state.clients);
  const { members } = useSelector((state) => state.organizations);
  const [fields, setFields] = useState(initialFields);
  const [error, setError] = useState("");
  const [editingClient, setEditingClient] = useState(null);
  const [deletingClient, setDeletingClient] = useState(null);

  const isAdminOrOwner = computeIsAdminOrOwner(user, members);

  const clientFormConfig = {
    name: "Client",
    fields: [
      { name: "name", type: "text", label: t("common.name"), required: true },
      { name: "email", type: "email", label: t("common.email") },
      { name: "address", type: "text", label: t("common.address") },
      { name: "postalCode", type: "text", label: t("common.postalCode") },
      { name: "city", type: "text", label: t("common.city") },
      { name: "country", type: "text", label: t("common.country") },
      { name: "vatNumber", type: "text", label: t("common.vatNumber") },
      { name: "vatRate", type: "number", label: t("clients.vatRateLabel"), step: "0.01" },
      { name: "vatDischargeMotive", type: "text", label: t("clients.vatDischargeMotive") },
      { name: "purchaseOrder", type: "text", label: t("clients.purchaseOrder") },
      { name: "hoursPerDay", type: "number", label: t("clients.hoursPerDay"), step: "0.01" },
      { name: "dailyRate", type: "number", label: t("clients.dailyRate"), step: "0.01", min: "0" },
    ],
  };

  useEffect(() => {
    if (currentOrgId) {
      dispatch(listClientsApi(currentOrgId, user.token));
      dispatch(listMembersApi(currentOrgId, user.token));
    }
  }, [dispatch, currentOrgId, user.token]);

  const setField = (key, value) => setFields({ ...fields, [key]: value });

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    if (!currentOrgId) return;
    const payload = {
      ...fields,
      vatRate: fields.vatRate === "" ? undefined : Number(fields.vatRate),
      hoursPerDay: fields.hoursPerDay === "" ? undefined : Number(fields.hoursPerDay),
      dailyRate: fields.dailyRate === "" ? undefined : Number(fields.dailyRate),
    };
    try {
      await dispatch(createClientApi(currentOrgId, payload, user.token));
      setFields(initialFields);
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    }
  };

  const handleDelete = async () => {
    const clientId = deletingClient.id;
    setDeletingClient(null);
    await dispatch(deleteClientApi(currentOrgId, clientId, user.token));
  };

  if (!currentOrgId) {
    return <h1 className="cw-title">{t("organizations.selectOrCreateFirst")}</h1>;
  }

  return (
    <div className={styles.main}>
      <h1 className="cw-title">{t("clients.title")}</h1>
      <CollapsiblePanel title={t("clients.createClient")}>
        <ConfigForm
          config={clientFormConfig}
          values={fields}
          onChange={setField}
          onSubmit={handleSubmit}
          submitLabel={t("common.add")}
          error={error}
        />
      </CollapsiblePanel>

      {clients.length === 0 && (
        <EmptyState title={t("clients.emptyTitle")} body={t("clients.emptyBody")} />
      )}

      <ul className="cw-list">
        {clients.map((client) => (
          <li className={`cw-list-item ${styles.clientRow}`} key={client.id}>
            <span>
              <strong>{client.name}</strong>
              {client.address && `, ${client.address}`}
              {client.postalCode && ` ${client.postalCode}`}
              {client.city && ` ${client.city}`}
              {client.country && ` ${client.country}`}
              {t("clients.vatAndHours", { rate: client.vatRate, hours: client.hoursPerDay })}
              {isAdminOrOwner && client.dailyRate
                ? t("clients.dailyRateSet", { rate: client.dailyRate })
                : ""}
            </span>
            <div className={styles.rowActions}>
              <CopyIdButton id={client.id} className={styles.iconBtn} />
              {isAdminOrOwner && (
                <Tooltip label={t("common.edit")}>
                  <button
                    type="button"
                    className={styles.iconBtn}
                    onClick={() => setEditingClient(client)}
                  >
                    <FaRegEdit style={{ fontSize: "18px" }} />
                  </button>
                </Tooltip>
              )}
              {isAdminOrOwner && (
                <Tooltip label={t("common.delete")}>
                  <button
                    type="button"
                    className={`${styles.iconBtn} ${styles.iconBtnDanger}`}
                    onClick={() => setDeletingClient(client)}
                  >
                    <MdDeleteForever style={{ fontSize: "20px" }} />
                  </button>
                </Tooltip>
              )}
            </div>
          </li>
        ))}
      </ul>

      <EditClientModal
        show={!!editingClient}
        onClose={() => setEditingClient(null)}
        targetClient={editingClient}
        orgId={currentOrgId}
        token={user.token}
      />

      <ConfirmModal
        show={!!deletingClient}
        title={t("clients.deleteClientTitle")}
        body={deletingClient ? t("clients.deleteClientBody", { name: deletingClient.name }) : ""}
        confirmLabel={t("common.delete")}
        onConfirm={handleDelete}
        onCancel={() => setDeletingClient(null)}
      />
    </div>
  );
};

export default Clients;
