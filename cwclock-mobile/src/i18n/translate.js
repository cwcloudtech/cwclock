import AsyncStorage from "@react-native-async-storage/async-storage";
import en from "./en";
import fr from "./fr";

export const STORAGE_KEY = "cwclock.locale";
export const dictionaries = { en, fr };

export const LANGUAGES = [
  { code: "en", label: "English" },
  { code: "fr", label: "Français" },
];

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

// AsyncStorage is async (unlike the web app's synchronous localStorage), so
// the stored locale is loaded once by I18nContext on mount rather than at
// module load time - it starts as "en" for the very first render.
export const loadStoredLocale = async () => {
  const stored = await AsyncStorage.getItem(STORAGE_KEY);
  return stored && dictionaries[stored] ? stored : "en";
};

export const storeLocale = (locale) => AsyncStorage.setItem(STORAGE_KEY, locale);

// apiErrorMessage translates a failed request for display, whichever of the
// two request paths it came from: an axios error (error.response.data -
// src/api/client.js calls, same i18n_code/message shape cwclock-api sends
// everywhere) or the Error thrown by src/api/blobRequest.js's PDF fetches
// (error.i18nCode/error.message, same underlying shape, just not nested
// under .response.data since it isn't an axios error).
export const apiErrorMessage = (error, locale) => {
  const i18nCode = error?.response?.data?.i18n_code || error?.i18nCode;
  if (i18nCode) {
    const translated = resolve(dictionaries[locale] ?? dictionaries.en, i18nCode);
    if (translated) return translated;
  }
  const message = error?.response?.data?.message || error?.message;
  return message || translate(locale, "errors.network");
};
