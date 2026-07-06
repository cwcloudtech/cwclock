import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import Modal from "../common/Modal";
import ConfigForm from "../common/ConfigForm";
import { updateOrgApi } from "../../Redux/Organizations/Org.actions";
import { listAllOrganizationsApi } from "../../Redux/Admin/Admin.actions";
import { useI18n } from "../../i18n/I18nContext";

const emptyFields = {
  name: "",
  address: "",
  postalCode: "",
  city: "",
  country: "",
  vatNumber: "",
  siren: "",
  siret: "",
  picture: "",
};

const EditOrgModal = ({ show, onClose, targetOrg, token }) => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const [fields, setFields] = useState(emptyFields);
  const [error, setError] = useState("");

  const orgFormConfig = {
    name: "Organization",
    fields: [
      { name: "name", type: "text", label: t("common.name"), required: true },
      { name: "address", type: "text", label: t("common.address") },
      { name: "postalCode", type: "text", label: t("common.postalCode") },
      { name: "city", type: "text", label: t("common.city") },
      { name: "country", type: "text", label: t("common.country") },
      { name: "vatNumber", type: "text", label: t("common.vatNumber") },
      { name: "siren", type: "text", label: "SIREN" },
      { name: "siret", type: "text", label: "SIRET" },
      { name: "picture", type: "image", label: t("common.picture") },
    ],
  };

  useEffect(() => {
    if (show && targetOrg) {
      setFields({
        name: targetOrg.name || "",
        address: targetOrg.address || "",
        postalCode: targetOrg.postalCode || "",
        city: targetOrg.city || "",
        country: targetOrg.country || "",
        vatNumber: targetOrg.vatNumber || "",
        siren: targetOrg.siren || "",
        siret: targetOrg.siret || "",
        picture: targetOrg.picture || "",
      });
      setError("");
    }
  }, [show, targetOrg]);

  if (!targetOrg) return null;

  const setField = (key, value) => setFields({ ...fields, [key]: value });

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    try {
      await dispatch(updateOrgApi(targetOrg.id, fields, token));
      await dispatch(listAllOrganizationsApi(token));
      onClose();
    } catch (err) {
      setError(t("organizations.nameRequired"));
    }
  };

  return (
    <Modal show={show} title={t("admin.editOrgTitle", { name: targetOrg.name })} onClose={onClose}>
      <ConfigForm
        config={orgFormConfig}
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

export default EditOrgModal;
