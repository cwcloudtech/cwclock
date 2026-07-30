import React from "react";
import { StyleSheet, Text, View } from "react-native";
import { Picker } from "@react-native-picker/picker";
import theme from "../theme";

// SelectField wraps the community Picker with the same label styling as
// FormField/DateField, for project/client pickers and the report-type /
// invoice-status style single-choice fields.
const SelectField = ({ label, value, onValueChange, items, placeholder, enabled = true }) => (
  <View style={styles.container}>
    {label ? <Text style={styles.label}>{label}</Text> : null}
    <View style={[styles.pickerWrap, !enabled && styles.pickerWrapDisabled]}>
      <Picker selectedValue={value} onValueChange={onValueChange} enabled={enabled}>
        {placeholder ? <Picker.Item label={placeholder} value="" /> : null}
        {items.map((item) => (
          <Picker.Item key={item.value} label={item.label} value={item.value} />
        ))}
      </Picker>
    </View>
  </View>
);

const styles = StyleSheet.create({
  container: {
    marginBottom: theme.spacing(2),
  },
  label: {
    fontSize: 13,
    color: theme.color.textMuted,
    marginBottom: theme.spacing(0.5),
  },
  pickerWrap: {
    borderWidth: 1,
    borderColor: theme.color.border,
    borderRadius: theme.radius,
    overflow: "hidden",
  },
  pickerWrapDisabled: {
    opacity: 0.5,
  },
});

export default SelectField;
