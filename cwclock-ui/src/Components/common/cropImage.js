const loadImage = (src) =>
  new Promise((resolve, reject) => {
    const image = new Image();
    image.addEventListener("load", () => resolve(image));
    image.addEventListener("error", reject);
    image.src = src;
  });

// Draws the cropped region of imageSrc onto an outputSize x outputSize
// canvas, so every uploaded avatar/logo/stamp ends up the same fixed
// dimensions regardless of the source image's size or aspect ratio.
const cropImage = async (imageSrc, cropPixels, outputSize) => {
  const image = await loadImage(imageSrc);
  const canvas = document.createElement("canvas");
  canvas.width = outputSize;
  canvas.height = outputSize;

  const ctx = canvas.getContext("2d");
  ctx.drawImage(
    image,
    cropPixels.x,
    cropPixels.y,
    cropPixels.width,
    cropPixels.height,
    0,
    0,
    outputSize,
    outputSize
  );

  return canvas.toDataURL("image/png");
};

export default cropImage;
