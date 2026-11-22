import en from "./translations/en";
import fr from "./translations/fr";

export const STORAGE_KEY = "locale";
export const dictionaries = { en, fr };

// LANGUAGES drives both the account-menu language picker and browser-locale
// detection, so adding a third language only means adding it here plus a
// translations/xx.js dictionary.
export const LANGUAGES = [
  { code: "en", label: "English", flag: "🇬🇧" },
  { code: "fr", label: "Français", flag: "🇫🇷" },
];

export const getStoredLocale = () => {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored && dictionaries[stored]) {
    return stored;
  }
  const browserLocale = (navigator.language || "en").slice(0, 2).toLowerCase();
  return dictionaries[browserLocale] ? browserLocale : "en";
};

export const resolve = (dict, key) =>
  key.split(".").reduce((acc, part) => (acc && acc[part] !== undefined ? acc[part] : undefined), dict);

export const translate = (locale, key, vars) => {
  let str = resolve(dictionaries[locale], key) ?? resolve(dictionaries.en, key) ?? key;
  if (vars) {
    Object.keys(vars).forEach((k) => {
      str = str.replaceAll(`{{${k}}}`, vars[k]);
    });
  }
  return str;
};

// apiErrorMessage translates an API error for display in a toast: it prefers
// the server's i18n_code (translated locally), falling back to the raw
// message the API sent, then to a generic network-error string.
export const apiErrorMessage = (error, locale) => {
  const data = error?.response?.data;
  let message = translate(locale, "errors.network");
  if (data?.i18n_code) {
    const translated = resolve(dictionaries[locale] ?? dictionaries.en, data.i18n_code);
    if (translated) message = translated;
    else if (data?.message) message = data.message;
  } else if (data?.message) {
    message = data.message;
  }
  // A 5xx is a server fault the user can't act on (the message is generic):
  // append the correlation id (X-Cid response header) so it can be quoted to
  // support, tying the toast to the [ERROR] alert and the logs. A 4xx message
  // stands on its own and gets no id.
  const cid = error?.response?.headers?.["x-cid"];
  if (Number(error?.response?.status) >= 500 && cid) {
    message = `${message} (ref: ${cid})`;
  }
  return message;
};
