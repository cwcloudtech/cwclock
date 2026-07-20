// Date-range shortcuts for the reports time selector, Grafana-style. Weeks
// are Sunday-first, matching this app's own day-of-week indexing (see
// i18n "days" dictionary, where 0 = Sunday).
const startOfDay = (d) => {
  const x = new Date(d);
  x.setHours(0, 0, 0, 0);
  return x;
};
const addDays = (d, n) => {
  const x = new Date(d);
  x.setDate(x.getDate() + n);
  return x;
};
const startOfWeek = (d) => addDays(startOfDay(d), -d.getDay());

export const toISODate = (d) => {
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
};

// fromISODate parses a "YYYY-MM-DD" string into a local Date at midnight -
// toISODate's inverse. Deliberately not `new Date(isoString)`: that parses
// as UTC midnight, which shifts to the previous/next day once read back
// through local getters (getFullYear/getMonth/getDate, as toISODate and
// this module's other helpers all do) for any timezone ahead of UTC.
export const fromISODate = (iso) => {
  const [yyyy, mm, dd] = iso.split("-").map(Number);
  return new Date(yyyy, mm - 1, dd);
};

// referenceDate defaults to the real current date/time - pass a specific
// one (e.g. the user's most recently logged time entry) to compute these
// shortcuts as if that were "today" instead, as Reports.jsx's initial
// default range does (ai-instruct-63).
export const dateRangeShortcuts = (t, referenceDate = new Date()) => {
  const today = startOfDay(referenceDate);

  return [
    { key: "today", label: t("reports.shortcutToday"), range: () => [today, today] },
    {
      key: "yesterday",
      label: t("reports.shortcutYesterday"),
      range: () => {
        const y = addDays(today, -1);
        return [y, y];
      },
    },
    { key: "thisWeek", label: t("reports.shortcutThisWeek"), range: () => [startOfWeek(today), today] },
    {
      key: "lastWeek",
      label: t("reports.shortcutLastWeek"),
      range: () => {
        const start = addDays(startOfWeek(today), -7);
        return [start, addDays(start, 6)];
      },
    },
    { key: "pastTwoWeeks", label: t("reports.shortcutPastTwoWeeks"), range: () => [addDays(today, -13), today] },
    {
      key: "thisMonth",
      label: t("reports.shortcutThisMonth"),
      range: () => [new Date(today.getFullYear(), today.getMonth(), 1), today],
    },
    {
      key: "lastMonth",
      label: t("reports.shortcutLastMonth"),
      range: () => [
        new Date(today.getFullYear(), today.getMonth() - 1, 1),
        new Date(today.getFullYear(), today.getMonth(), 0),
      ],
    },
    {
      key: "pastTwoMonths",
      label: t("reports.shortcutPastTwoMonths"),
      range: () => [new Date(today.getFullYear(), today.getMonth() - 1, 1), today],
    },
    {
      key: "thisYear",
      label: t("reports.shortcutThisYear"),
      range: () => [new Date(today.getFullYear(), 0, 1), today],
    },
    {
      key: "lastYear",
      label: t("reports.shortcutLastYear"),
      range: () => [new Date(today.getFullYear() - 1, 0, 1), new Date(today.getFullYear() - 1, 11, 31)],
    },
  ];
};
