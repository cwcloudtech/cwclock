import React from "react";

const variantClass = {
  primary: "",
  secondary: "cw-button--secondary",
  danger: "cw-button--danger",
  ghost: "cw-button--ghost",
};

const Button = ({ variant = "primary", size, className = "", children, ...props }) => {
  const classes = ["cw-button", variantClass[variant], size === "sm" ? "cw-button--sm" : "", className]
    .filter(Boolean)
    .join(" ");

  return (
    <button type="button" className={classes} {...props}>
      {children}
    </button>
  );
};

export default Button;
