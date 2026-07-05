import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import Modal from "./Modal";
import Button from "./Button";
import { updateProfileApi, updatePictureApi } from "../../Redux/Users/User.actions";
import fileToBase64 from "./fileToBase64";
import styles from "./Styles/EditProfileModal.module.css";

const EditProfileModal = ({ show, onClose, user }) => {
  const dispatch = useDispatch();
  const [name, setName] = useState(user.name || "");
  const [surname, setSurname] = useState(user.surname || "");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (show) {
      setName(user.name || "");
      setSurname(user.surname || "");
      setPassword("");
      setConfirmPassword("");
      setError("");
    }
  }, [show, user.name, user.surname]);

  const handleAvatarChange = async (e) => {
    const file = e.target.files[0];
    if (!file) return;
    const picture = await fileToBase64(file);
    dispatch(updatePictureApi(picture, user.token));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!name || !surname) {
      setError("Please add both first and last name.");
      return;
    }
    if (password && password !== confirmPassword) {
      setError("Passwords do not match.");
      return;
    }
    try {
      await dispatch(updateProfileApi(name, surname, password, confirmPassword, user.token));
      onClose();
    } catch (err) {
      setError("Could not update your profile. Please try again.");
    }
  };

  return (
    <Modal show={show} title="Edit profile" onClose={onClose}>
      <div className={styles.avatarRow}>
        {user.picture ? (
          <img src={user.picture} alt="" className={styles.avatar} />
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
            autoFocus
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
        <div className="cw-field">
          <label className="cw-label">Confirm password</label>
          <input
            className="cw-input"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            title="Confirm password"
          />
        </div>
        {error && <p className="cw-error">{error}</p>}
        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 16 }}>
          <Button type="button" variant="secondary" onClick={onClose} title="Discard changes">
            Cancel
          </Button>
          <Button type="submit" title="Save your profile">
            Save
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default EditProfileModal;
