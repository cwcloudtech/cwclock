import React, { useState } from "react";
import { toast } from "react-toastify";
import ImageCropModal from "./ImageCropModal";
import fileToBase64 from "./fileToBase64";
import MAX_IMAGE_SIZE from "./maxImageSize";
import toastOptions from "../../Redux/toastOptions";
import { translate, getStoredLocale } from "../../i18n/translate";

// Wraps the "pick a file -> position it -> hand back {image, x, y}" pipeline
// so every image upload spot (avatar, organization picture, stamp) shares
// the same behavior instead of storing the raw picked file as-is. The image
// itself is passed through untouched (never resized/redrawn); only its
// display position is chosen. The trigger UI itself (a file input, a label,
// ...) is left to the caller via a render prop, since each spot's existing
// layout differs.
const ImagePicker = ({ onChange, children }) => {
  const [rawImage, setRawImage] = useState(null);

  const loadFile = async (file) => {
    if (!file) return;
    if (file.size > MAX_IMAGE_SIZE) {
      toast.error(translate(getStoredLocale(), "errors.imageTooLarge"), toastOptions);
      return;
    }
    setRawImage(await fileToBase64(file));
  };

  const handlePick = async (e) => {
    const file = e.target.files[0];
    e.target.value = "";
    await loadFile(file);
  };

  const handleConfirm = (result) => {
    setRawImage(null);
    onChange(result);
  };

  return (
    <>
      {children({ onPick: handlePick, onFile: loadFile })}
      <ImageCropModal
        show={!!rawImage}
        imageSrc={rawImage}
        onCancel={() => setRawImage(null)}
        onConfirm={handleConfirm}
      />
    </>
  );
};

export default ImagePicker;
