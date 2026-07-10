import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import Modal from "../common/Modal";
import ConfirmModal from "../common/ConfirmModal";
import Button from "../common/Button";
import RequiredMark from "../common/RequiredMark";
import useEmailAutocomplete from "../common/useEmailAutocomplete";
import { transferOwnershipApi } from "../../Redux/Organizations/Org.actions";
import { listAllOrganizationsApi } from "../../Redux/Admin/Admin.actions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";

const TransferOwnershipModal = ({ show, onClose, targetOrg, token }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");
  const [confirmingTransfer, setConfirmingTransfer] = useState(false);
  const suggestions = useEmailAutocomplete(email, show, token);

  useEffect(() => {
    if (show) {
      setEmail("");
      setError("");
      setConfirmingTransfer(false);
    }
  }, [show, targetOrg]);

  if (!targetOrg) return null;

  const handleSubmit = (e) => {
    e.preventDefault();
    setError("");
    if (!email) return;
    setConfirmingTransfer(true);
  };

  const handleTransferConfirm = async () => {
    setConfirmingTransfer(false);
    try {
      await dispatch(transferOwnershipApi(targetOrg.id, email, token));
      await dispatch(listAllOrganizationsApi(token));
      onClose();
    } catch (err) {
      setError(apiErrorMessage(err, locale));
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

      <ConfirmModal
        show={confirmingTransfer}
        title={t("admin.transfer")}
        body={t("organizations.confirmTransfer", { name: targetOrg.name, email })}
        confirmLabel={t("admin.transfer")}
        onConfirm={handleTransferConfirm}
        onCancel={() => setConfirmingTransfer(false)}
      />
    </Modal>
  );
};

export default TransferOwnershipModal;
