export const formatHMS = (secs) => {
  const h = Math.floor(secs / 3600);
  const m = Math.floor((secs % 3600) / 60);
  const s = secs % 60;
  return [h, m, s].map((v) => String(v).padStart(2, "0")).join(":");
};

// Compact "1h05" style used by chart hover tooltips (daily bar chart and
// project donut chart), shorter than formatHMS since it never needs seconds.
export const formatHours = (secs) => {
  const h = Math.floor(secs / 3600);
  const m = Math.round((secs % 3600) / 60);
  return `${h}h${String(m).padStart(2, "0")}`;
};
