import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import Modal from "../common/Modal";
import ConfigForm from "../common/ConfigForm";
import { updateClientApi } from "../../Redux/Clients/Client.actions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";

const emptyFields = {
  name: "",
  address: "",
  postalCode: "",
  city: "",
  country: "",
  vatNumber: "",
  vatRate: "",
  vatDischargeMotive: "",
  purchaseOrder: "",
  hoursPerDay: "",
};

const EditClientModal = ({ show, onClose, targetClient, orgId, token }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const [fields, setFields] = useState(emptyFields);
  const [error, setError] = useState("");

  const clientFormConfig = {
    name: "Client",
    fields: [
      { name: "name", type: "text", label: t("common.name"), required: true },
      { name: "address", type: "text", label: t("common.address") },
      { name: "postalCode", type: "text", label: t("common.postalCode") },
      { name: "city", type: "text", label: t("common.city") },
      { name: "country", type: "text", label: t("common.country") },
      { name: "vatNumber", type: "text", label: t("common.vatNumber") },
      { name: "vatRate", type: "number", label: t("clients.vatRateLabel"), step: "0.01" },
      { name: "vatDischargeMotive", type: "text", label: t("clients.vatDischargeMotive") },
      { name: "purchaseOrder", type: "text", label: t("clients.purchaseOrder") },
      { name: "hoursPerDay", type: "number", label: t("clients.hoursPerDay"), step: "0.01" },
    ],
  };

  useEffect(() => {
    if (show && targetClient) {
      setFields({
        name: targetClient.name || "",
        address: targetClient.address || "",
        postalCode: targetClient.postalCode || "",
        city: targetClient.city || "",
        country: targetClient.country || "",
        vatNumber: targetClient.vatNumber || "",
        vatRate: targetClient.vatRate ?? "",
        vatDischargeMotive: targetClient.vatDischargeMotive || "",
        purchaseOrder: targetClient.purchaseOrder || "",
        hoursPerDay: targetClient.hoursPerDay ?? "",
      });
      setError("");
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
    };
    try {
      await dispatch(updateClientApi(orgId, targetClient.id, payload, token));
      onClose();
    } catch (err) {
      setError(apiErrorMessage(err, locale));
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
    </Modal>
  );
};

export default EditClientModal;
