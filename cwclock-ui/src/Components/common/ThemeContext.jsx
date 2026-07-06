import React, { createContext, useCallback, useContext, useEffect, useState } from "react";

const STORAGE_KEY = "theme";
const ThemeContext = createContext(null);

const getSystemTheme = () =>
  window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";

// Applies the OS-level color scheme by default; once the user toggles the
// switch in the profile dropdown, that explicit choice is persisted and
// wins over the OS setting in both directions (light OS + dark choice, or
// the reverse), via the [data-theme] attribute on <html> (see index.css).
export const ThemeProvider = ({ children }) => {
  const [explicitTheme, setExplicitTheme] = useState(() => localStorage.getItem(STORAGE_KEY));
  const [systemTheme, setSystemTheme] = useState(getSystemTheme);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = (e) => setSystemTheme(e.matches ? "dark" : "light");
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);

  useEffect(() => {
    if (explicitTheme) {
      document.documentElement.setAttribute("data-theme", explicitTheme);
    } else {
      document.documentElement.removeAttribute("data-theme");
    }
  }, [explicitTheme]);

  const theme = explicitTheme || systemTheme;

  const toggleTheme = useCallback(() => {
    setExplicitTheme((current) => {
      const next = (current || getSystemTheme()) === "dark" ? "light" : "dark";
      localStorage.setItem(STORAGE_KEY, next);
      return next;
    });
  }, []);

  return <ThemeContext.Provider value={{ theme, toggleTheme }}>{children}</ThemeContext.Provider>;
};

export const useTheme = () => useContext(ThemeContext);
