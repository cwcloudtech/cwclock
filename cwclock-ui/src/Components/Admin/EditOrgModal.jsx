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
import useCountries from "../common/useCountries";
import useCountryFields from "../common/useCountryFields";
import { identificationFieldConfig } from "../common/identificationFields";

const emptyFields = {
  name: "",
  accountingEmail: "",
  address: "",
  postalCode: "",
  city: "",
  country: "",
  vatNumber: "",
  siren: "",
  siret: "",
  naf: "",
  mf: "",
  identificationNumber: "",
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
  const countries = useCountries();
  const identificationFields = useCountryFields(fields.country);

  const orgFormConfig = {
    name: "Organization",
    fields: [
      { name: "name", type: "text", label: t("common.name"), required: true },
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
      ...identificationFields.map((name) => identificationFieldConfig(name, t)),
      {
        name: "accountingEmail",
        type: "email",
        label: t("organizations.accountingEmail"),
        placeholder: t("organizations.accountingEmailPlaceholder"),
      },
      {
        name: "currency",
        type: "select",
        label: t("common.currency"),
        required: true,
        options: currencies.map((c) => ({ value: c.iso, label: c.iso })),
      },
      { name: "picture", type: "image", label: t("common.picture") },
      { name: "stamp", type: "image", label: t("organizations.stamp") },
    ],
  };

  useEffect(() => {
    if (show && targetOrg) {
      setFields({
        name: targetOrg.name || "",
        accountingEmail: targetOrg.accountingEmail || "",
        address: targetOrg.address || "",
        postalCode: targetOrg.postalCode || "",
        city: targetOrg.city || "",
        country: targetOrg.country || "",
        vatNumber: targetOrg.vatNumber || "",
        siren: targetOrg.siren || "",
        siret: targetOrg.siret || "",
        naf: targetOrg.naf || "",
        mf: targetOrg.mf || "",
        identificationNumber: targetOrg.identificationNumber || "",
        currency: targetOrg.currency || "",
        picture: targetOrg.picture || "",
        pictureX: targetOrg.pictureX ?? 50,
        pictureY: targetOrg.pictureY ?? 50,
        stamp: targetOrg.stamp || "",
        stampX: targetOrg.stampX ?? 50,
        stampY: targetOrg.stampY ?? 50,
      });
      setError("");
    }
    // Only react to the modal opening/target changing, not to the currency
    // list becoming available (that no longer affects the initial value).
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [show, targetOrg]);

  useEffect(() => {
    if (!fields.country || fields.currency) return;
    const country = countries.find((c) => c.iso === fields.country);
    if (country) setFields((f) => ({ ...f, currency: country.currency }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fields.country, countries]);

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
