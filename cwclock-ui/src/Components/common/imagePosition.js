// Converts a react-easy-crop percentage crop rect (x/y/width/height, all in
// % of the source image, with zoom locked at 1 so the crop box always
// exactly covers the image on its shorter axis) into CSS object-position
// percentages. Storing just this (x, y) - rather than a resized/redrawn
// copy of the image - lets the frontend reproduce the same framing later
// with a plain object-fit: cover + object-position, on the untouched
// original image.
const cropAreaToPosition = ({ x, y, width, height }) => ({
  x: width >= 100 ? 0 : (x / (100 - width)) * 100,
  y: height >= 100 ? 0 : (y / (100 - height)) * 100,
});

export default cropAreaToPosition;
