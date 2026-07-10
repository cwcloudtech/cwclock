// Maps a business identification field name returned by GET /v1/fields
// (ai-instruct-35's per-country decision table, e.g. "SIRET", "VAT Code")
// to the form field it fills in and a translated label - so the UI stays
// localized even though the backend's field names are plain English.
const FIELD_DEFS = {
  SIRET: { key: "siret", fallbackLabel: "SIRET" },
  SIREN: { key: "siren", fallbackLabel: "SIREN" },
  NAF: { key: "naf", labelKey: "organizations.naf" },
  MF: { key: "mf", fallbackLabel: "MF" },
  "VAT Code": { key: "vatNumber", labelKey: "common.vatNumber" },
  "Identification number": { key: "identificationNumber", labelKey: "common.identificationNumber" },
};

// Builds a ConfigForm field definition for a single decision-table entry.
export const identificationFieldConfig = (name, t) => {
  const def = FIELD_DEFS[name] || { key: name, fallbackLabel: name };
  return {
    name: def.key,
    type: "text",
    label: def.labelKey ? t(def.labelKey) : def.fallbackLabel,
  };
};

// The full set of form field keys any identification field can map to, so
// callers can reset them all when the country (and therefore the set of
// fields to show) changes.
export const IDENTIFICATION_FIELD_KEYS = Object.values(FIELD_DEFS).map((d) => d.key);
