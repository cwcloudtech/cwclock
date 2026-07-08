import React from "react";
import Button from "./Button";
import RequiredMark from "./RequiredMark";
import ImagePicker from "./ImagePicker";
import DropZone from "./DropZone";
import contrastColor from "./contrastColor";
import { useI18n } from "../../i18n/I18nContext";
import styles from "./Styles/ConfigForm.module.css";

// Renders a form from a field-config object instead of hand-rolled JSX, e.g.:
// { name: "Organization", fields: [{ name: "name", type: "text", label: "Name", required: true }] }
const ConfigForm = ({
  config,
  values,
  onChange,
  onSubmit,
  submitLabel,
  onCancel,
  cancelLabel,
  error,
}) => {
  const { t } = useI18n();
  const resolvedSubmitLabel = submitLabel ?? t("common.save");
  const resolvedCancelLabel = cancelLabel ?? t("common.cancel");

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
            <option value="">{field.placeholder || `${t("common.select")} ${field.label}`}</option>
            {(field.options || []).map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        );
      case "image":
        return (
          <ImagePicker onChange={(base64) => onChange(field.name, base64)}>
            {({ onPick, onFile }) => (
              <DropZone onFile={onFile}>
                <input className="cw-input" type="file" accept="image/*" onChange={onPick} />
                {value && <img src={value} alt="" className="cw-image-preview" />}
              </DropZone>
            )}
          </ImagePicker>
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
      case "color": {
        const isValidHex = /^#[0-9a-f]{6}$/i.test(value);
        return (
          <div className={styles.colorField}>
            <input
              className={styles.colorInput}
              type="color"
              value={isValidHex ? value : "#000000"}
              onChange={(e) => onChange(field.name, e.target.value)}
            />
            <input
              className={styles.colorHex}
              type="text"
              value={value}
              onChange={(e) => onChange(field.name, e.target.value)}
              style={{ backgroundColor: isValidHex ? value : "transparent", color: contrastColor(value) }}
              maxLength={7}
            />
          </div>
        );
      }
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
      case "date":
        return (
          <input
            className="cw-input"
            type="date"
            min={field.min}
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
      <div className={styles.actions}>
        <Button type="submit" title={resolvedSubmitLabel}>{resolvedSubmitLabel}</Button>
        {onCancel && (
          <Button type="button" variant="secondary" onClick={onCancel} title={resolvedCancelLabel}>
            {resolvedCancelLabel}
          </Button>
        )}
      </div>
      {error && <p className="cw-error">{error}</p>}
    </form>
  );
};

export default ConfigForm;
