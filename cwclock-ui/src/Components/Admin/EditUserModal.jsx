import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import Modal from "../common/Modal";
import Button from "../common/Button";
import ImagePicker from "../common/ImagePicker";
import DropZone from "../common/DropZone";
import { updateUserApi } from "../../Redux/Admin/Admin.actions";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import styles from "./Styles/Admin.module.css";

const ROLES = ["superuser", "confirmed", "disabled", "ban"];

const EditUserModal = ({ show, onClose, targetUser, token }) => {
  const { t, locale } = useI18n();
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

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!email || !name || !surname) {
      setError(t("admin.pleaseValidUserFields"));
      return;
    }
    const fields = { email, name, surname, role };
    if (password) fields.password = password;
    if (picture !== undefined) {
      fields.picture = picture.image;
      fields.pictureX = picture.x;
      fields.pictureY = picture.y;
    }
    try {
      await dispatch(updateUserApi(targetUser.id, fields, token));
      onClose();
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    }
  };

  return (
    <Modal show={show} title={t("admin.editUserTitle", { email: targetUser.email })} onClose={onClose}>
      <div className={styles.avatarRow}>
        {(() => {
          const shown = picture ?? {
            image: targetUser.picture,
            x: targetUser.pictureX ?? 50,
            y: targetUser.pictureY ?? 50,
          };
          return shown.image ? (
            <img
              src={shown.image}
              alt=""
              className={styles.avatar}
              style={{ objectPosition: `${shown.x}% ${shown.y}%` }}
            />
          ) : (
            <div className={styles.avatarPlaceholder} />
          );
        })()}
        <ImagePicker onChange={setPicture}>
          {({ onPick, onFile }) => (
            <DropZone onFile={onFile}>
              <label className={styles.avatarLabel} title={t("common.uploadNewPicture")}>
                {t("common.updateAvatar")}
                <input type="file" accept="image/*" hidden onChange={onPick} />
              </label>
            </DropZone>
          )}
        </ImagePicker>
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
