import React, { useEffect, useState } from "react";
import Cropper from "react-easy-crop";
import Modal from "./Modal";
import Button from "./Button";
import cropAreaToPosition from "./imagePosition";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/ImageCropModal.module.css";

// Lets the user pick which part of a just-selected image stays visible in
// its fixed-size frame (avatar/organization picture/stamp). The image
// itself is never resized or redrawn - only the chosen x/y position is kept,
// and the frontend later displays the untouched original with
// object-fit: cover + object-position to reproduce the same framing. Zoom is
// locked at 1 (the image always exactly covers the crop frame) since only a
// position, not a scale, is stored.
const ImageCropModal = ({ show, imageSrc, onCancel, onConfirm }) => {
  const { t } = useI18n();
  const [crop, setCrop] = useState({ x: 0, y: 0 });
  const [position, setPosition] = useState({ x: 50, y: 50 });

  useEffect(() => {
    if (show) {
      setCrop({ x: 0, y: 0 });
      setPosition({ x: 50, y: 50 });
    }
  }, [show, imageSrc]);

  if (!show) return null;

  const handleConfirm = () => {
    onConfirm({ image: imageSrc, x: position.x, y: position.y });
  };

  return (
    <Modal show={show} title={t("common.cropImageTitle")} onClose={onCancel}>
      <div className={styles.cropArea}>
        <Cropper
          image={imageSrc}
          crop={crop}
          zoom={1}
          minZoom={1}
          maxZoom={1}
          aspect={1}
          cropShape="round"
          showGrid={false}
          onCropChange={setCrop}
          onCropComplete={(croppedAreaPercent) => setPosition(cropAreaToPosition(croppedAreaPercent))}
        />
      </div>
      <div className={styles.actions}>
        <Button variant="secondary" onClick={onCancel} title={t("common.cancel")}>
          {t("common.cancel")}
        </Button>
        <Button onClick={handleConfirm} title={t("common.save")}>
          {t("common.save")}
        </Button>
      </div>
    </Modal>
  );
};

export default ImageCropModal;
