import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import Modal from "../common/Modal";
import Button from "../common/Button";
import fileToBase64 from "../common/fileToBase64";
import { updateUserApi } from "../../Redux/Admin/Admin.actions";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/Admin.module.css";

const ROLES = ["superuser", "confirmed", "disabled"];

const EditUserModal = ({ show, onClose, targetUser, token }) => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [surname, setSurname] = useState("");
  const [role, setRole] = useState("confirmed");
  const [password, setPassword] = useState("");
  const [picture, setPicture] = useState(undefined);
  const [error, setError] = useState("");

  useEffect(() => {
    if (show && targetUser) {
      setEmail(targetUser.email || "");
      setName(targetUser.name || "");
      setSurname(targetUser.surname || "");
      setRole(targetUser.role || "confirmed");
      setPassword("");
      setPicture(undefined);
      setError("");
    }
  }, [show, targetUser]);

  if (!targetUser) return null;

  const roleLabel = (r) => t(`common.globalRole${r.charAt(0).toUpperCase()}${r.slice(1)}`);

  const handleAvatarChange = async (e) => {
    const file = e.target.files[0];
    if (!file) return;
    setPicture(await fileToBase64(file));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!email || !name || !surname) {
      setError(t("admin.pleaseValidUserFields"));
      return;
    }
    const fields = { email, name, surname, role };
    if (password) fields.password = password;
    if (picture !== undefined) fields.picture = picture;
    try {
      await dispatch(updateUserApi(targetUser.id, fields, token));
      onClose();
    } catch (err) {
      setError(
        err.response?.status === 400 && /email/i.test(err.response?.data?.message || "")
          ? t("admin.emailInUse")
          : t("admin.couldNotUpdateUser")
      );
    }
  };

  return (
    <Modal show={show} title={t("admin.editUserTitle", { email: targetUser.email })} onClose={onClose}>
      <div className={styles.avatarRow}>
        {(picture ?? targetUser.picture) ? (
          <img src={picture ?? targetUser.picture} alt="" className={styles.avatar} />
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
          <label className="cw-label">{t("common.email")}</label>
          <input
            className="cw-input"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            title={t("common.email")}
          />
        </div>
        <div className="cw-field">
          <label className="cw-label">{t("common.firstName")}</label>
          <input
            className="cw-input"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            title={t("common.firstName")}
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
          <label className="cw-label">{t("admin.globalRole")}</label>
          <select className="cw-select" value={role} onChange={(e) => setRole(e.target.value)} title={t("admin.globalRole")}>
            {ROLES.map((r) => (
              <option key={r} value={r}>
                {roleLabel(r)}
              </option>
            ))}
          </select>
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
        {error && <p className="cw-error">{error}</p>}
        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 16 }}>
          <Button type="button" variant="secondary" onClick={onClose} title={t("common.discardChanges")}>
            {t("common.cancel")}
          </Button>
          <Button type="submit" title={t("admin.saveUser")}>
            {t("common.save")}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default EditUserModal;
