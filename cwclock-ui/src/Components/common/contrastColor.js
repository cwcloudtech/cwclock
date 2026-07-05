// Picks black or white text so it stays readable over an arbitrary hex
// background color, using the standard relative-luminance threshold.
const contrastColor = (hex) => {
  const clean = (hex || "").replace("#", "");
  if (clean.length !== 6) return "#000000";
  const r = parseInt(clean.slice(0, 2), 16);
  const g = parseInt(clean.slice(2, 4), 16);
  const b = parseInt(clean.slice(4, 6), 16);
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
  return luminance > 0.6 ? "#000000" : "#ffffff";
};

export default contrastColor;
