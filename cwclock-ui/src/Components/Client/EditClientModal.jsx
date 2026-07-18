import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import Modal from "../common/Modal";
import ConfirmModal from "../common/ConfirmModal";
import ConfigForm from "../common/ConfigForm";
import AutocompleteSelect from "../common/AutocompleteSelect";
import Button from "../common/Button";
import { updateClientApi, transferClientApi } from "../../Redux/Clients/Client.actions";
import { isOrgOwner } from "../common/permissions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import useCountries from "../common/useCountries";
import useCountryFields from "../common/useCountryFields";
import { identificationFieldConfig } from "../common/identificationFields";

const emptyFields = {
  name: "",
  email: "",
  invoiceEmails: "",
  contactName: "",
  address: "",
  postalCode: "",
  city: "",
  country: "",
  vatNumber: "",
  vatRate: "",
  vatDischargeMotive: "",
  siren: "",
  siret: "",
  naf: "",
  mf: "",
  identificationNumber: "",
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
  const [confirmingTransfer, setConfirmingTransfer] = useState(false);

  const currentOrg = organizations.find((o) => o.id === orgId);
  const canTransfer = isOrgOwner(user, currentOrg);
  const ownedOrgOptions = organizations
    .filter((o) => o.ownerId === user.id && o.id !== orgId)
    .map((o) => ({ value: o.id, label: o.name }));
  const targetOrg = organizations.find((o) => o.id === targetOrgId);
  const countries = useCountries();
  const identificationFields = useCountryFields(fields.country);

  const clientFormConfig = {
    name: "Client",
    fields: [
      { name: "name", type: "text", label: t("common.name"), required: true },
      { name: "email", type: "email", label: t("common.email") },
      { name: "invoiceEmails", type: "text", label: t("clients.invoiceEmails"), placeholder: t("clients.invoiceEmailsHint") },
      { name: "contactName", type: "text", label: t("clients.contactName") },
      { name: "address", type: "text", label: t("common.address") },
      { name: "postalCode", type: "text", label: t("common.postalCode") },
      { name: "city", type: "text", label: t("common.city") },
      {
        name: "country",
        type: "autocomplete",
        label: t("common.country"),
        placeholder: t("common.country"),
        required: true,
        options: countries.map((c) => ({ value: c.iso, label: c.name })),
      },
      { name: "vatRate", type: "number", label: t("clients.vatRateLabel"), step: "0.01" },
      { name: "vatDischargeMotive", type: "text", label: t("clients.vatDischargeMotive") },
      ...identificationFields.map((name) => identificationFieldConfig(name, t)),
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
        invoiceEmails: targetClient.invoiceEmails || "",
        contactName: targetClient.contactName || "",
        address: targetClient.address || "",
        postalCode: targetClient.postalCode || "",
        city: targetClient.city || "",
        country: targetClient.country || "",
        vatNumber: targetClient.vatNumber || "",
        vatRate: targetClient.vatRate ?? "",
        vatDischargeMotive: targetClient.vatDischargeMotive || "",
        siren: targetClient.siren || "",
        siret: targetClient.siret || "",
        naf: targetClient.naf || "",
        mf: targetClient.mf || "",
        identificationNumber: targetClient.identificationNumber || "",
        purchaseOrder: targetClient.purchaseOrder || "",
        hoursPerDay: targetClient.hoursPerDay ?? "",
        dailyRate: targetClient.dailyRate ?? "",
      });
      setError("");
      setTargetOrgId("");
      setTransferError("");
      setConfirmingTransfer(false);
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

  const handleTransferClick = () => {
    setTransferError("");
    if (!targetOrgId) return;
    setConfirmingTransfer(true);
  };

  const handleTransferConfirm = async () => {
    setConfirmingTransfer(false);
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
        <>
          <hr className="cw-hr" />
          <div className="cw-field">
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
              onClick={handleTransferClick}
              title={t("clients.transferHint")}
            >
              {t("clients.transferButton")}
            </Button>
            {transferError && <p className="cw-error">{transferError}</p>}
          </div>
        </>
      )}

      <ConfirmModal
        show={confirmingTransfer}
        title={t("clients.transferButton")}
        body={t("clients.transferConfirmBody", { name: targetClient.name, orgName: targetOrg?.name || "" })}
        confirmLabel={t("clients.transferButton")}
        onConfirm={handleTransferConfirm}
        onCancel={() => setConfirmingTransfer(false)}
      />
    </Modal>
  );
};

export default EditClientModal;
