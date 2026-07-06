import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import Modal from "../common/Modal";
import Button from "../common/Button";
import RequiredMark from "../common/RequiredMark";
import useEmailAutocomplete from "../common/useEmailAutocomplete";
import { transferOwnershipApi } from "../../Redux/Organizations/Org.actions";
import { listAllOrganizationsApi } from "../../Redux/Admin/Admin.actions";
import { useI18n } from "../../i18n/I18nContext";

const TransferOwnershipModal = ({ show, onClose, targetOrg, token }) => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");
  const suggestions = useEmailAutocomplete(email, show, token);

  useEffect(() => {
    if (show) {
      setEmail("");
      setError("");
    }
  }, [show, targetOrg]);

  if (!targetOrg) return null;

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!email) return;
    if (!window.confirm(t("organizations.confirmTransfer", { name: targetOrg.name, email }))) {
      return;
    }
    setError("");
    try {
      await dispatch(transferOwnershipApi(targetOrg.id, email, token));
      await dispatch(listAllOrganizationsApi(token));
      onClose();
    } catch (err) {
      setError(t("organizations.couldNotTransfer"));
    }
  };

  return (
    <Modal show={show} title={t("admin.transferModalTitle", { name: targetOrg.name })} onClose={onClose}>
      <form onSubmit={handleSubmit}>
        <div className="cw-field">
          <label className="cw-label">
            {t("organizations.newOwnerEmail")}
            <RequiredMark />
          </label>
          <input
            className="cw-input"
            list="transfer-ownership-suggestions"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            title={t("organizations.newOwnerEmail")}
            required
            autoFocus
          />
          <datalist id="transfer-ownership-suggestions">
            {suggestions.map((u) => (
              <option key={u.id} value={u.email} />
            ))}
          </datalist>
        </div>
        {error && <p className="cw-error">{error}</p>}
        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 16 }}>
          <Button type="button" variant="secondary" onClick={onClose} title={t("common.discardChanges")}>
            {t("common.cancel")}
          </Button>
          <Button type="submit" variant="danger" title={t("organizations.transferOwnershipTooltip")}>
            {t("admin.transfer")}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default TransferOwnershipModal;
