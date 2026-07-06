import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import Modal from "./Modal";
import Button from "./Button";
import { updateProfileApi, updatePictureApi } from "../../Redux/Users/User.actions";
import fileToBase64 from "./fileToBase64";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/EditProfileModal.module.css";

const EditProfileModal = ({ show, onClose, user }) => {
  const { t } = useI18n();
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
      setError(t("profile.pleaseAddNames"));
      return;
    }
    if (password && password !== confirmPassword) {
      setError(t("profile.passwordsDoNotMatch"));
      return;
    }
    try {
      await dispatch(updateProfileApi(name, surname, password, confirmPassword, user.token));
      onClose();
    } catch (err) {
      setError(t("profile.couldNotUpdate"));
    }
  };

  return (
    <Modal show={show} title={t("profile.editProfileTitle")} onClose={onClose}>
      <div className={styles.avatarRow}>
        {user.picture ? (
          <img src={user.picture} alt="" className={styles.avatar} />
        ) : (
          <div className={styles.avatarPlaceholder} />
        )}
        <label className={styles.avatarLabel} title={t("common.uploadNewPicture")}>
          {t("common.updateAvatar")}
          <input type="file" accept="image/*" hidden onChange={handleAvatarChange} />
        </label>
      </div>

      <form onSubmit={handleSubmit}>
        <div className="cw-field">
          <label className="cw-label">{t("common.firstName")}</label>
          <input
            className="cw-input"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            title={t("common.firstName")}
            autoFocus
          />
        </div>
        <div className="cw-field">
          <label className="cw-label">{t("common.lastName")}</label>
          <input
            className="cw-input"
            type="text"
            value={surname}
            onChange={(e) => setSurname(e.target.value)}
            title={t("common.lastName")}
          />
        </div>
        <div className="cw-field">
          <label className="cw-label">{t("common.newPassword")}</label>
          <input
            className="cw-input"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder={t("common.leaveBlankPassword")}
            title={t("common.leaveBlankPassword")}
          />
        </div>
        <div className="cw-field">
          <label className="cw-label">{t("common.confirmPassword")}</label>
          <input
            className="cw-input"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            title={t("common.confirmPassword")}
          />
        </div>
        {error && <p className="cw-error">{error}</p>}
        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 16 }}>
          <Button type="button" variant="secondary" onClick={onClose} title={t("common.discardChanges")}>
            {t("common.cancel")}
          </Button>
          <Button type="submit" title={t("profile.saveProfile")}>
            {t("common.save")}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default EditProfileModal;
