import React, { createContext, useCallback, useContext, useMemo, useState } from "react";
import { STORAGE_KEY, LANGUAGES, dictionaries, getStoredLocale, translate } from "./translate";

const I18nContext = createContext(null);

export { LANGUAGES };

// Minimal i18n: a flat-ish nested dictionary per locale plus a t(key, vars)
// lookup with {{var}} interpolation and an English fallback for missing
// keys, which is enough for this app without pulling in a full i18n library.
export const I18nProvider = ({ children }) => {
  const [locale, setLocaleState] = useState(getStoredLocale);

  const setLocale = useCallback((next) => {
    if (!dictionaries[next]) return;
    localStorage.setItem(STORAGE_KEY, next);
    setLocaleState(next);
  }, []);

  const t = useCallback((key, vars) => translate(locale, key, vars), [locale]);

  const value = useMemo(() => ({ locale, setLocale, t }), [locale, setLocale, t]);

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
};

export const useI18n = () => useContext(I18nContext);
