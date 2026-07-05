import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import Modal from "../common/Modal";
import Button from "../common/Button";
import fileToBase64 from "../common/fileToBase64";
import { updateUserApi } from "../../Redux/Admin/Admin.actions";
import styles from "./Styles/Admin.module.css";

const ROLES = ["superuser", "confirmed", "disabled"];

const EditUserModal = ({ show, onClose, targetUser, token }) => {
  const dispatch = useDispatch();
  const [name, setName] = useState("");
  const [surname, setSurname] = useState("");
  const [role, setRole] = useState("confirmed");
  const [password, setPassword] = useState("");
  const [picture, setPicture] = useState(undefined);
  const [error, setError] = useState("");

  useEffect(() => {
    if (show && targetUser) {
      setName(targetUser.name || "");
      setSurname(targetUser.surname || "");
      setRole(targetUser.role || "confirmed");
      setPassword("");
      setPicture(undefined);
      setError("");
    }
  }, [show, targetUser]);

  if (!targetUser) return null;

  const handleAvatarChange = async (e) => {
    const file = e.target.files[0];
    if (!file) return;
    setPicture(await fileToBase64(file));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!name || !surname) {
      setError("Please add both first and last name.");
      return;
    }
    const fields = { name, surname, role };
    if (password) fields.password = password;
    if (picture !== undefined) fields.picture = picture;
    try {
      await dispatch(updateUserApi(targetUser.id, fields, token));
      onClose();
    } catch (err) {
      setError("Could not update this user. Please try again.");
    }
  };

  return (
    <Modal show={show} title={`Edit ${targetUser.email}`} onClose={onClose}>
      <div className={styles.avatarRow}>
        {(picture ?? targetUser.picture) ? (
          <img src={picture ?? targetUser.picture} alt="" className={styles.avatar} />
        ) : (
          <div className={styles.avatarPlaceholder} />
        )}
        <label className={styles.avatarLabel} title="Upload a new profile picture">
          Update avatar
          <input type="file" accept="image/*" hidden onChange={handleAvatarChange} />
        </label>
      </div>

      <form onSubmit={handleSubmit}>
        <div className="cw-field">
          <label className="cw-label">First name</label>
          <input
            className="cw-input"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            title="First name"
          />
        </div>
        <div className="cw-field">
          <label className="cw-label">Last name</label>
          <input
            className="cw-input"
            type="text"
            value={surname}
            onChange={(e) => setSurname(e.target.value)}
            title="Last name"
          />
        </div>
        <div className="cw-field">
          <label className="cw-label">Global role</label>
          <select className="cw-select" value={role} onChange={(e) => setRole(e.target.value)} title="Global role">
            {ROLES.map((r) => (
              <option key={r} value={r}>
                {r}
              </option>
            ))}
          </select>
        </div>
        <div className="cw-field">
          <label className="cw-label">New password</label>
          <input
            className="cw-input"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Leave blank to keep the current password"
            title="Leave blank to keep the current password"
          />
        </div>
        {error && <p className="cw-error">{error}</p>}
        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 16 }}>
          <Button type="button" variant="secondary" onClick={onClose} title="Discard changes">
            Cancel
          </Button>
          <Button type="submit" title="Save this user">
            Save
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default EditUserModal;
