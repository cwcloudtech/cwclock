import React, { createContext, useCallback, useContext, useMemo, useState } from "react";
import en from "./translations/en";
import fr from "./translations/fr";

const STORAGE_KEY = "locale";
const dictionaries = { en, fr };

const I18nContext = createContext(null);

const detectLocale = () => {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored && dictionaries[stored]) return stored;
  const browserLocale = (navigator.language || "en").slice(0, 2).toLowerCase();
  return dictionaries[browserLocale] ? browserLocale : "en";
};

const resolve = (dict, key) =>
  key.split(".").reduce((acc, part) => (acc && acc[part] !== undefined ? acc[part] : undefined), dict);

// Minimal i18n: a flat-ish nested dictionary per locale plus a t(key, vars)
// lookup with {{var}} interpolation and an English fallback for missing
// keys, which is enough for this app without pulling in a full i18n library.
export const I18nProvider = ({ children }) => {
  const [locale, setLocaleState] = useState(detectLocale);

  const setLocale = useCallback((next) => {
    if (!dictionaries[next]) return;
    localStorage.setItem(STORAGE_KEY, next);
    setLocaleState(next);
  }, []);

  const t = useCallback(
    (key, vars) => {
      let str = resolve(dictionaries[locale], key) ?? resolve(dictionaries.en, key) ?? key;
      if (vars) {
        Object.keys(vars).forEach((k) => {
          str = str.replaceAll(`{{${k}}}`, vars[k]);
        });
      }
      return str;
    },
    [locale]
  );

  const value = useMemo(() => ({ locale, setLocale, t }), [locale, setLocale, t]);

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
};

export const useI18n = () => useContext(I18nContext);
