import React from "react";
import Button from "./Button";
import RequiredMark from "./RequiredMark";
import fileToBase64 from "./fileToBase64";

// Renders a form from a field-config object instead of hand-rolled JSX, e.g.:
// { name: "Organization", fields: [{ name: "name", type: "text", label: "Name", required: true }] }
const ConfigForm = ({ config, values, onChange, onSubmit, submitLabel = "Save", error }) => {
  const setImageField = (field) => async (e) => {
    const file = e.target.files[0];
    if (!file) return;
    onChange(field.name, await fileToBase64(file));
  };

  const renderControl = (field) => {
    const value = values[field.name] ?? "";

    switch (field.type) {
      case "checkbox":
        return (
          <label className="cw-checkbox">
            <input
              type="checkbox"
              checked={!!value}
              onChange={(e) => onChange(field.name, e.target.checked)}
            />
            {field.label}
          </label>
        );
      case "select":
        return (
          <select
            className="cw-select"
            value={value}
            onChange={(e) => onChange(field.name, e.target.value)}
            required={field.required}
          >
            <option value="">{field.placeholder || `Select ${field.label}`}</option>
            {(field.options || []).map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        );
      case "image":
        return (
          <>
            <input className="cw-input" type="file" accept="image/*" onChange={setImageField(field)} />
            {value && <img src={value} alt="" className="cw-image-preview" />}
          </>
        );
      case "number":
        return (
          <input
            className="cw-input"
            type="number"
            step={field.step}
            min={field.min}
            placeholder={field.placeholder}
            value={value}
            onChange={(e) => onChange(field.name, e.target.value)}
            required={field.required}
          />
        );
      case "color":
        return (
          <input
            className="cw-input"
            type="color"
            value={value}
            onChange={(e) => onChange(field.name, e.target.value)}
          />
        );
      case "email":
        return (
          <input
            className="cw-input"
            type="email"
            placeholder={field.placeholder}
            value={value}
            onChange={(e) => onChange(field.name, e.target.value)}
            required={field.required}
          />
        );
      default:
        return (
          <input
            className="cw-input"
            type="text"
            placeholder={field.placeholder}
            value={value}
            onChange={(e) => onChange(field.name, e.target.value)}
            required={field.required}
          />
        );
    }
  };

  return (
    <form onSubmit={onSubmit}>
      {config.fields.map((field) => (
        <div className="cw-field" key={field.name}>
          {field.type !== "checkbox" && (
            <label className="cw-label">
              {field.label}
              {field.required && <RequiredMark />}
            </label>
          )}
          {renderControl(field)}
        </div>
      ))}
      <Button type="submit">{submitLabel}</Button>
      {error && <p className="cw-error">{error}</p>}
    </form>
  );
};

export default ConfigForm;
