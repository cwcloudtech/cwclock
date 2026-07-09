// Max upload size (bytes) for avatar/logo/stamp images, enforced client-side
// before ever calling the API. Baked in at build time via
// REACT_APP_MAX_IMAGE_SIZE (see .env.react), mirroring the backend's own
// CWCLOCK_MAX_IMAGE_SIZE limit - falls back to 2MB when unset or unparsable.
const parsed = Number(process.env.REACT_APP_MAX_IMAGE_SIZE);
const MAX_IMAGE_SIZE = Number.isFinite(parsed) && parsed > 0 ? parsed : 2 * 1024 * 1024;

export default MAX_IMAGE_SIZE;
