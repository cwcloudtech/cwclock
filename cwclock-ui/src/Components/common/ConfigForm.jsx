import React from "react";
import Form from "react-bootstrap/Form";
import Button from "react-bootstrap/Button";
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
          <Form.Check
            type="checkbox"
            label={field.label}
            checked={!!value}
            onChange={(e) => onChange(field.name, e.target.checked)}
          />
        );
      case "select":
        return (
          <Form.Select
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
          </Form.Select>
        );
      case "image":
        return (
          <>
            <Form.Control type="file" accept="image/*" onChange={setImageField(field)} />
            {value && <img src={value} alt="" className="cw-image-preview" />}
          </>
        );
      case "number":
        return (
          <Form.Control
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
          <Form.Control
            type="color"
            value={value}
            onChange={(e) => onChange(field.name, e.target.value)}
          />
        );
      case "email":
        return (
          <Form.Control
            type="email"
            placeholder={field.placeholder}
            value={value}
            onChange={(e) => onChange(field.name, e.target.value)}
            required={field.required}
          />
        );
      default:
        return (
          <Form.Control
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
    <Form onSubmit={onSubmit}>
      {config.fields.map((field) => (
        <Form.Group className="mb-2" key={field.name}>
          {field.type !== "checkbox" && (
            <Form.Label>
              {field.label}
              {field.required && <RequiredMark />}
            </Form.Label>
          )}
          {renderControl(field)}
        </Form.Group>
      ))}
      <Button type="submit">{submitLabel}</Button>
      {error && <p className="cw-error">{error}</p>}
    </Form>
  );
};

export default ConfigForm;
