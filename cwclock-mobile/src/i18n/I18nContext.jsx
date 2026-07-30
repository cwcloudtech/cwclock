import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { LANGUAGES, dictionaries, loadStoredLocale, storeLocale, translate } from "./translate";

const I18nContext = createContext(null);

export { LANGUAGES };

// Minimal i18n: a flat-ish nested dictionary per locale plus a t(key, vars)
// lookup with {{var}} interpolation and an English fallback for missing
// keys - ported from cwclock-ui/src/i18n/I18nContext.jsx, swapped from
// synchronous localStorage to async AsyncStorage (see translate.js).
export const I18nProvider = ({ children }) => {
  const [locale, setLocaleState] = useState("en");

  useEffect(() => {
    loadStoredLocale().then(setLocaleState);
  }, []);

  const setLocale = useCallback((next) => {
    if (!dictionaries[next]) return;
    storeLocale(next);
    setLocaleState(next);
  }, []);

  const t = useCallback((key, vars) => translate(locale, key, vars), [locale]);

  const value = useMemo(() => ({ locale, setLocale, t }), [locale, setLocale, t]);

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
};

export const useI18n = () => useContext(I18nContext);
