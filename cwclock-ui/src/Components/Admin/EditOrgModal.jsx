import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import Modal from "../common/Modal";
import ConfigForm from "../common/ConfigForm";
import { updateOrgApi } from "../../Redux/Organizations/Org.actions";
import { listAllOrganizationsApi } from "../../Redux/Admin/Admin.actions";

const orgFormConfig = {
  name: "Organization",
  fields: [
    { name: "name", type: "text", label: "Name", required: true },
    { name: "address", type: "text", label: "Address" },
    { name: "postalCode", type: "text", label: "Postal code" },
    { name: "city", type: "text", label: "City" },
    { name: "country", type: "text", label: "Country" },
    { name: "vatNumber", type: "text", label: "VAT number" },
    { name: "siren", type: "text", label: "SIREN" },
    { name: "siret", type: "text", label: "SIRET" },
    { name: "picture", type: "image", label: "Picture" },
  ],
};

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
  const dispatch = useDispatch();
  const [fields, setFields] = useState(emptyFields);
  const [error, setError] = useState("");

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
      setError("Name is required.");
    }
  };

  return (
    <Modal show={show} title={`Edit ${targetOrg.name}`} onClose={onClose}>
      <ConfigForm
        config={orgFormConfig}
        values={fields}
        onChange={setField}
        onSubmit={handleSubmit}
        submitLabel="Save"
        onCancel={onClose}
        error={error}
      />
    </Modal>
  );
};

export default EditOrgModal;
