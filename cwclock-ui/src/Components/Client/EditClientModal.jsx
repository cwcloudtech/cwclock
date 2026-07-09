import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import Modal from "../common/Modal";
import ConfigForm from "../common/ConfigForm";
import AutocompleteSelect from "../common/AutocompleteSelect";
import Button from "../common/Button";
import { updateClientApi, transferClientApi } from "../../Redux/Clients/Client.actions";
import { isOrgOwner } from "../common/permissions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";

const emptyFields = {
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

const EditClientModal = ({ show, onClose, targetClient, orgId, token }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { organizations } = useSelector((state) => state.organizations);
  const [fields, setFields] = useState(emptyFields);
  const [error, setError] = useState("");
  const [targetOrgId, setTargetOrgId] = useState("");
  const [transferError, setTransferError] = useState("");

  const currentOrg = organizations.find((o) => o.id === orgId);
  const canTransfer = isOrgOwner(user, currentOrg);
  const ownedOrgOptions = organizations
    .filter((o) => o.ownerId === user.id && o.id !== orgId)
    .map((o) => ({ value: o.id, label: o.name }));
  const targetOrg = organizations.find((o) => o.id === targetOrgId);

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
    if (show && targetClient) {
      setFields({
        name: targetClient.name || "",
        email: targetClient.email || "",
        address: targetClient.address || "",
        postalCode: targetClient.postalCode || "",
        city: targetClient.city || "",
        country: targetClient.country || "",
        vatNumber: targetClient.vatNumber || "",
        vatRate: targetClient.vatRate ?? "",
        vatDischargeMotive: targetClient.vatDischargeMotive || "",
        purchaseOrder: targetClient.purchaseOrder || "",
        hoursPerDay: targetClient.hoursPerDay ?? "",
        dailyRate: targetClient.dailyRate ?? "",
      });
      setError("");
      setTargetOrgId("");
      setTransferError("");
    }
  }, [show, targetClient]);

  if (!targetClient) return null;

  const setField = (key, value) => setFields({ ...fields, [key]: value });

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    const payload = {
      ...fields,
      vatRate: fields.vatRate === "" ? undefined : Number(fields.vatRate),
      hoursPerDay: fields.hoursPerDay === "" ? undefined : Number(fields.hoursPerDay),
      dailyRate: fields.dailyRate === "" ? undefined : Number(fields.dailyRate),
    };
    try {
      await dispatch(updateClientApi(orgId, targetClient.id, payload, token));
      onClose();
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    }
  };

  const handleTransfer = async () => {
    setTransferError("");
    if (!targetOrgId) return;
    if (!window.confirm(t("clients.transferConfirmBody", { name: targetClient.name, orgName: targetOrg?.name || "" }))) {
      return;
    }
    try {
      await dispatch(transferClientApi(orgId, targetClient.id, targetOrgId, token));
      onClose();
    } catch (err) {
      setTransferError(apiErrorMessage(err, locale));
    }
  };

  return (
    <Modal show={show} title={t("clients.editClientTitle", { name: targetClient.name })} onClose={onClose}>
      <ConfigForm
        config={clientFormConfig}
        values={fields}
        onChange={setField}
        onSubmit={handleSubmit}
        submitLabel={t("common.save")}
        onCancel={onClose}
        error={error}
      />

      {canTransfer && ownedOrgOptions.length > 0 && (
        <div className="cw-field" style={{ marginTop: "2rem" }}>
          <AutocompleteSelect
            label={t("clients.transferTargetOrg")}
            placeholder={t("clients.transferTargetOrg")}
            options={ownedOrgOptions}
            value={targetOrgId}
            onChange={setTargetOrgId}
          />
          <Button
            variant="danger"
            disabled={!targetOrgId}
            onClick={handleTransfer}
            title={t("clients.transferHint")}
          >
            {t("clients.transferButton")}
          </Button>
          {transferError && <p className="cw-error">{transferError}</p>}
        </div>
      )}
    </Modal>
  );
};

export default EditClientModal;
