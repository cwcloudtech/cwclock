// ConfigForm's "image" field type reports a single {image, x, y} value via
// its onChange(field.name, value) call. This expands it into the 3 flat
// keys (e.g. picture/pictureX/pictureY) the create/update API payloads
// expect, keyed off the field's own name, while leaving every other field
// type's plain onChange(key, value) untouched.
const applyImageField = (fields, key, value) => {
  if (value && typeof value === "object" && "image" in value) {
    return { ...fields, [key]: value.image, [`${key}X`]: value.x, [`${key}Y`]: value.y };
  }
  return { ...fields, [key]: value };
};

export default applyImageField;
