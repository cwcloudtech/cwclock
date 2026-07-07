import React, { useState } from "react";
import ImageCropModal from "./ImageCropModal";
import fileToBase64 from "./fileToBase64";

// Wraps the "pick a file -> crop it -> resize to a fixed square" pipeline so
// every image upload spot (avatar, organization picture, stamp) shares the
// same behavior instead of storing the raw picked file as-is. The trigger
// UI itself (a file input, a label, ...) is left to the caller via a render
// prop, since each spot's existing layout differs.
const ImagePicker = ({ onChange, children }) => {
  const [rawImage, setRawImage] = useState(null);

  const handlePick = async (e) => {
    const file = e.target.files[0];
    e.target.value = "";
    if (!file) return;
    setRawImage(await fileToBase64(file));
  };

  const handleConfirm = (base64) => {
    setRawImage(null);
    onChange(base64);
  };

  return (
    <>
      {children({ onPick: handlePick })}
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
