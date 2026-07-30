import React from "react";
import { QRCodeSVG } from "qrcode.react";
import Modal from "./Modal";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/AndroidQrModal.module.css";

// Shown instead of following the Android download link directly on
// non-Android devices, where opening the link is useless (ai-instruct-114).
const AndroidQrModal = ({ show, onClose, url }) => {
  const { t } = useI18n();

  return (
    <Modal show={show} title={t("nav.downloadAndroidApp")} onClose={onClose}>
      <p className={styles.instructions}>{t("nav.scanToDownloadAndroidApp")}</p>
      {url && (
        <div className={styles.qrWrapper}>
          <QRCodeSVG value={url} size={200} />
        </div>
      )}
    </Modal>
  );
};

export default AndroidQrModal;
