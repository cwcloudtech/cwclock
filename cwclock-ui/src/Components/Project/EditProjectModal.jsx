import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import Modal from "../common/Modal";
import ConfigForm from "../common/ConfigForm";
import AutocompleteSelect from "../common/AutocompleteSelect";
import Button from "../common/Button";
import { updateProjectApi } from "../../Redux/Projects/Project.actions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";

const emptyFields = { name: "", color: "#1cb9f7", dailyRate: "", subdivisions: [] };

const EditProjectModal = ({ show, onClose, targetProject, orgId, token }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { clients } = useSelector((state) => state.clients);
  const [fields, setFields] = useState(emptyFields);
  const [error, setError] = useState("");
  const [targetClientId, setTargetClientId] = useState("");
  const [transferError, setTransferError] = useState("");

  const otherClientOptions = clients
    .filter((c) => c.id !== targetProject?.clientId)
    .map((c) => ({ value: c.id, label: c.name }));
  const targetClient = clients.find((c) => c.id === targetClientId);

  const projectFormConfig = {
    name: "Project",
    fields: [
      { name: "name", type: "text", label: t("common.name"), required: true },
      { name: "color", type: "color", label: t("common.color") },
      { name: "dailyRate", type: "number", label: t("projects.dailyRate"), step: "0.01", min: "0" },
      {
        name: "subdivisions",
        type: "tags",
        label: t("projects.subdivisions"),
        placeholder: t("projects.subdivisionsPlaceholder"),
      },
    ],
  };

  useEffect(() => {
    if (show && targetProject) {
      setFields({
        name: targetProject.name || "",
        color: targetProject.color || "#1cb9f7",
        dailyRate: targetProject.dailyRate ?? "",
        subdivisions: targetProject.subdivisions || [],
      });
      setError("");
      setTargetClientId("");
      setTransferError("");
    }
  }, [show, targetProject]);

  if (!targetProject) return null;

  const setField = (key, value) => setFields({ ...fields, [key]: value });

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    const dailyRate = fields.dailyRate === "" ? undefined : Number(fields.dailyRate);
    try {
      await dispatch(
        updateProjectApi(orgId, targetProject.id, "", fields.name, fields.color, dailyRate, fields.subdivisions, token)
      );
      onClose();
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    }
  };

  const handleTransfer = async () => {
    setTransferError("");
    if (!targetClientId) return;
    if (
      !window.confirm(
        t("projects.transferConfirmBody", { name: targetProject.name, clientName: targetClient?.name || "" })
      )
    ) {
      return;
    }
    try {
      await dispatch(
        updateProjectApi(
          orgId,
          targetProject.id,
          targetClientId,
          targetProject.name,
          targetProject.color,
          targetProject.dailyRate,
          targetProject.subdivisions,
          token
        )
      );
      onClose();
    } catch (err) {
      setTransferError(apiErrorMessage(err, locale));
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

      {otherClientOptions.length > 0 && (
        <>
          <hr className="cw-hr" />
          <div className="cw-field">
            <AutocompleteSelect
              label={t("projects.transferTargetClient")}
              placeholder={t("projects.transferTargetClient")}
              options={otherClientOptions}
              value={targetClientId}
              onChange={setTargetClientId}
            />
            <Button variant="danger" disabled={!targetClientId} onClick={handleTransfer} title={t("projects.transferHint")}>
              {t("projects.transferButton")}
            </Button>
            {transferError && <p className="cw-error">{transferError}</p>}
          </div>
        </>
      )}
    </Modal>
  );
};

export default EditProjectModal;
