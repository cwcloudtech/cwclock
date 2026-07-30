// A small shared palette, not a full design system - matches cwclock-ui's
// --cw-primary-500 (src/index.css) for basic brand consistency.
const theme = {
  color: {
    primary: "#1cb9f7",
    primaryDark: "#0f8fc2",
    danger: "#e5484d",
    text: "#1a1a1a",
    textMuted: "#6b7280",
    border: "#e2e8f0",
    background: "#ffffff",
    backgroundMuted: "#f5f7fa",
    white: "#ffffff",
  },
  spacing: (n) => n * 8,
  radius: 8,
};

export default theme;
