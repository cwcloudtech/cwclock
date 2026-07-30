import React, { useState } from "react";
import { Platform, StyleSheet, Text, TouchableOpacity } from "react-native";
import DateTimePicker from "@react-native-community/datetimepicker";
import theme from "../theme";

const formatValue = (value, mode) => {
  if (mode === "time") {
    return `${String(value.getHours()).padStart(2, "0")}:${String(value.getMinutes()).padStart(2, "0")}`;
  }
  return value.toISOString().slice(0, 10);
};

// DateField shows a formatted date/time value; tapping it opens the native
// picker dialog (Android shows this as a dialog itself, so the component
// only needs to render <DateTimePicker> while `show` is true). `value`/the
// value passed to onChange are both plain Date objects - screens convert to
// the API's "YYYY-MM-DD"/"HH:MM:SS" strings at the edge (see
// src/common/format.js and each screen's submit handler).
const DateField = ({ label, value, onChange, mode = "date" }) => {
  const [show, setShow] = useState(false);

  return (
    <>
      <TouchableOpacity style={styles.field} onPress={() => setShow(true)}>
        {label ? <Text style={styles.label}>{label}</Text> : null}
        <Text style={styles.value}>{formatValue(value, mode)}</Text>
      </TouchableOpacity>
      {show && (
        <DateTimePicker
          value={value}
          mode={mode}
          is24Hour
          display="default"
          onChange={(event, selected) => {
            setShow(Platform.OS === "ios");
            if (event.type === "set" && selected) {
              onChange(selected);
            }
          }}
        />
      )}
    </>
  );
};

const styles = StyleSheet.create({
  field: {
    borderWidth: 1,
    borderColor: theme.color.border,
    borderRadius: theme.radius,
    paddingHorizontal: theme.spacing(1.5),
    paddingVertical: theme.spacing(1.25),
    marginBottom: theme.spacing(2),
  },
  label: {
    fontSize: 13,
    color: theme.color.textMuted,
    marginBottom: theme.spacing(0.5),
  },
  value: {
    fontSize: 16,
    color: theme.color.text,
  },
});

export default DateField;
