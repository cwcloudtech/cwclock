// Zero-pads each ":"-separated segment of a time string (e.g. "9:5:3" ->
// "09:05:03"). The native <input type="time"> element requires a strict
// "HH:MM" or "HH:MM:SS" format per the HTML spec and silently blanks out
// anything looser, so an entry saved with an un-padded hour/minute/second
// (e.g. from Date#getHours() interpolated straight into a string) could no
// longer be displayed or edited (ai-instruct-40).
const padTimeString = (value) => {
  if (!value) return value;
  return value
    .split(":")
    .map((part) => part.padStart(2, "0"))
    .join(":");
};

export default padTimeString;
