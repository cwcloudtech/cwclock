import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import Modal from "../common/Modal";
import ConfigForm from "../common/ConfigForm";
import applyImageField from "../common/applyImageField";
import { updateOrgApi } from "../../Redux/Organizations/Org.actions";
import { listAllOrganizationsApi } from "../../Redux/Admin/Admin.actions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import useCurrencies from "../common/useCurrencies";

const emptyFields = {
  name: "",
  address: "",
  postalCode: "",
  city: "",
  country: "",
  vatNumber: "",
  siren: "",
  siret: "",
  naf: "",
  picture: "",
  pictureX: 50,
  pictureY: 50,
  stamp: "",
  stampX: 50,
  stampY: 50,
  currency: "",
};

const EditOrgModal = ({ show, onClose, targetOrg, token }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const [fields, setFields] = useState(emptyFields);
  const [error, setError] = useState("");
  const currencies = useCurrencies();

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
      { name: "naf", type: "text", label: t("organizations.naf") },
      {
        name: "currency",
        type: "select",
        label: t("common.currency"),
        required: true,
        options: currencies.map((code) => ({ value: code, label: code })),
      },
      { name: "picture", type: "image", label: t("common.picture") },
      { name: "stamp", type: "image", label: t("organizations.stamp") },
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
        naf: targetOrg.naf || "",
        currency: targetOrg.currency || currencies[0] || "",
        picture: targetOrg.picture || "",
        pictureX: targetOrg.pictureX ?? 50,
        pictureY: targetOrg.pictureY ?? 50,
        stamp: targetOrg.stamp || "",
        stampX: targetOrg.stampX ?? 50,
        stampY: targetOrg.stampY ?? 50,
      });
      setError("");
    }
  }, [show, targetOrg, currencies]);

  if (!targetOrg) return null;

  const setField = (key, value) => setFields((f) => applyImageField(f, key, value));

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    try {
      await dispatch(updateOrgApi(targetOrg.id, fields, token));
      await dispatch(listAllOrganizationsApi(token));
      onClose();
    } catch (err) {
      setError(apiErrorMessage(err, locale));
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
