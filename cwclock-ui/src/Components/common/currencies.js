// Fixed, ordered list of ISO 4217 currency codes an organization may bill
// in. Mirrors models.AllowedCurrencies on the backend (cwclock-api), which
// is the source of truth enforced on create/update.
const CURRENCIES = [
  "EUR", "USD", "GBP", "CAD", "CHF", "TND", "DZD", "MAD", "TRY", "EGP",
  "SAR", "AED", "QAR", "CNY", "HKD", "SGD", "JPY", "AUD", "NZD",
];

export const DEFAULT_CURRENCY = "EUR";

export default CURRENCIES;
