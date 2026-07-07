// Fixed output size (px) for every cropped/resized upload (avatar,
// organization picture, stamp). Baked in at build time via REACT_APP_IMAGE_SIZE
// (see .env.react), falling back to 256 for local dev where it isn't set.
const IMAGE_SIZE = Number(process.env.REACT_APP_IMAGE_SIZE) || 256;

export default IMAGE_SIZE;
