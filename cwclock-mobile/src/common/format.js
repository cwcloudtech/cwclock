// Ported from cwclock-ui/src/Components/common/padTimeString.js /
// Reports/reportFormat.js / TimerApp/TasksApp.jsx - same pure logic, no RN
// dependency, just copied over since the two apps don't share a workspace.

// Zero-pads each ":"-separated segment of a time string (e.g. "9:5:3" ->
// "09:05:03"), matching what the API's "HH:MM:SS" fields expect.
export const padTimeString = (value) => {
  if (!value) return value;
  return value
    .split(":")
    .map((part) => part.padStart(2, "0"))
    .join(":");
};

// toDayString/toHMS convert a Date (as produced by DateField/DateTimePicker)
// into the "YYYY-MM-DD"/"HH:MM:SS" strings the API expects for day/start/end
// - local calendar date and wall-clock time, no timezone conversion, same
// assumption cwclock-api's report/compute.go documents throughout.
export const toDayString = (date) => {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
};

export const toHMS = (date) =>
  [date.getHours(), date.getMinutes(), date.getSeconds()].map((v) => String(v).padStart(2, "0")).join(":");

export const formatHMS = (secs) => {
  const h = Math.floor(secs / 3600);
  const m = Math.floor((secs % 3600) / 60);
  const s = secs % 60;
  return [h, m, s].map((v) => String(v).padStart(2, "0")).join(":");
};

const parseSecondsOfDay = (hms) => {
  if (!hms) return 0;
  const [h, m, s] = hms.split(":").map(Number);
  return (h || 0) * 3600 + (m || 0) * 60 + (s || 0);
};

// entryDurationSecs mirrors TasksApp.jsx: an all-day or half-day entry
// contributes nothing to this personal duration total - that billing math
// needs the client's HoursPerDay, which this screen doesn't load (see
// cwclock-api's report/compute.go DurationSecs for the real duration those
// entries represent, used by Reports/Invoices instead).
export const entryDurationSecs = (item) => {
  if (item.allDay || item.half || !item.start || !item.end) return 0;
  const secs = parseSecondsOfDay(item.end) - parseSecondsOfDay(item.start);
  return secs > 0 ? secs : 0;
};

// groupByDay mirrors TasksApp.jsx's dayGroups: entries already arrive
// sorted day DESC from the API (and stay sorted client-side, see
// timeEntries.reducer.js's compareEntries), so a single pass grouping
// consecutive same-day items is enough.
export const groupByDay = (items) => {
  const groups = [];
  for (const item of items) {
    const last = groups[groups.length - 1];
    if (last && last.day === item.day) {
      last.items.push(item);
      last.totalSecs += entryDurationSecs(item);
    } else {
      groups.push({ day: item.day, items: [item], totalSecs: entryDurationSecs(item) });
    }
  }
  return groups;
};
