import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import Modal from "../common/Modal";
import Button from "../common/Button";
import RequiredMark from "../common/RequiredMark";
import useEmailAutocomplete from "../common/useEmailAutocomplete";
import { transferOwnershipApi } from "../../Redux/Organizations/Org.actions";
import { listAllOrganizationsApi } from "../../Redux/Admin/Admin.actions";

const TransferOwnershipModal = ({ show, onClose, targetOrg, token }) => {
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
    if (!window.confirm(`Transfer ownership of "${targetOrg.name}" to ${email}?`)) {
      return;
    }
    setError("");
    try {
      await dispatch(transferOwnershipApi(targetOrg.id, email, token));
      await dispatch(listAllOrganizationsApi(token));
      onClose();
    } catch (err) {
      setError("Could not transfer ownership. Check the email and try again.");
    }
  };

  return (
    <Modal show={show} title={`Transfer ownership of ${targetOrg.name}`} onClose={onClose}>
      <form onSubmit={handleSubmit}>
        <div className="cw-field">
          <label className="cw-label">
            New owner's email
            <RequiredMark />
          </label>
          <input
            className="cw-input"
            list="transfer-ownership-suggestions"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            title="New owner's email"
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
          <Button type="button" variant="secondary" onClick={onClose} title="Discard">
            Cancel
          </Button>
          <Button type="submit" variant="danger" title="Transfer ownership to this member">
            Transfer
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default TransferOwnershipModal;
