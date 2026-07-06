import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import Modal from "../common/Modal";
import ConfigForm from "../common/ConfigForm";
import { updateProjectApi } from "../../Redux/Projects/Project.actions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";

const emptyFields = { name: "", color: "#1cb9f7" };

const EditProjectModal = ({ show, onClose, targetProject, orgId, token }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const [fields, setFields] = useState(emptyFields);
  const [error, setError] = useState("");

  const projectFormConfig = {
    name: "Project",
    fields: [
      { name: "name", type: "text", label: t("common.name"), required: true },
      { name: "color", type: "color", label: t("common.color") },
    ],
  };

  useEffect(() => {
    if (show && targetProject) {
      setFields({
        name: targetProject.name || "",
        color: targetProject.color || "#1cb9f7",
      });
      setError("");
    }
  }, [show, targetProject]);

  if (!targetProject) return null;

  const setField = (key, value) => setFields({ ...fields, [key]: value });

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    try {
      await dispatch(updateProjectApi(orgId, targetProject.id, fields.name, fields.color, token));
      onClose();
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    }
  };

  return (
    <Modal show={show} title={t("projects.editProjectTitle", { name: targetProject.name })} onClose={onClose}>
      <ConfigForm
        config={projectFormConfig}
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

export default EditProjectModal;
