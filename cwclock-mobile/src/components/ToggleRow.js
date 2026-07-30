import React from "react";
import { StyleSheet, Switch, Text, View } from "react-native";
import theme from "../theme";

// ToggleRow is a labeled Switch row - used for the All day / Half day
// toggles (EditRecordScreen), which disable each other exactly like the web
// app's TaskComponent.jsx/ReportEntryRow.jsx (ai-instruct-104).
const ToggleRow = ({ label, value, onValueChange, disabled, style }) => (
  <View style={[styles.row, style]}>
    <Text style={[styles.label, disabled && styles.labelDisabled]}>{label}</Text>
    <Switch
      value={value}
      onValueChange={onValueChange}
      disabled={disabled}
      trackColor={{ true: theme.color.primary }}
    />
  </View>
);

const styles = StyleSheet.create({
  row: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingVertical: theme.spacing(1),
  },
  label: {
    fontSize: 16,
    color: theme.color.text,
  },
  labelDisabled: {
    color: theme.color.textMuted,
  },
});

export default ToggleRow;
