import React, { useEffect, useState } from "react";
import Cropper from "react-easy-crop";
import Modal from "./Modal";
import Button from "./Button";
import cropImage from "./cropImage";
import IMAGE_SIZE from "./imageSize";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/ImageCropModal.module.css";

// Lets the user pick the crop/zoom of a just-selected image before it's
// resized down to a fixed IMAGE_SIZE x IMAGE_SIZE square, so every uploaded
// avatar/logo/stamp ends up the same dimensions no matter the source photo.
const ImageCropModal = ({ show, imageSrc, onCancel, onConfirm }) => {
  const { t } = useI18n();
  const [crop, setCrop] = useState({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [croppedAreaPixels, setCroppedAreaPixels] = useState(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (show) {
      setCrop({ x: 0, y: 0 });
      setZoom(1);
      setCroppedAreaPixels(null);
      setSaving(false);
    }
  }, [show, imageSrc]);

  if (!show) return null;

  const handleConfirm = async () => {
    if (!croppedAreaPixels) return;
    setSaving(true);
    const base64 = await cropImage(imageSrc, croppedAreaPixels, IMAGE_SIZE);
    setSaving(false);
    onConfirm(base64);
  };

  return (
    <Modal show={show} title={t("common.cropImageTitle")} onClose={onCancel}>
      <div className={styles.cropArea}>
        <Cropper
          image={imageSrc}
          crop={crop}
          zoom={zoom}
          aspect={1}
          cropShape="round"
          showGrid={false}
          onCropChange={setCrop}
          onZoomChange={setZoom}
          onCropComplete={(_, pixels) => setCroppedAreaPixels(pixels)}
        />
      </div>
      <div className="cw-field">
        <label className="cw-label">{t("common.zoom")}</label>
        <input
          className={styles.zoomSlider}
          type="range"
          min={1}
          max={3}
          step={0.05}
          value={zoom}
          onChange={(e) => setZoom(Number(e.target.value))}
          title={t("common.zoom")}
        />
      </div>
      <div className={styles.actions}>
        <Button variant="secondary" onClick={onCancel} title={t("common.cancel")}>
          {t("common.cancel")}
        </Button>
        <Button onClick={handleConfirm} disabled={saving || !croppedAreaPixels} title={t("common.save")}>
          {t("common.save")}
        </Button>
      </div>
    </Modal>
  );
};

export default ImageCropModal;
