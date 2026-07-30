import React from "react";
import { StyleSheet, Text, TouchableOpacity, View } from "react-native";
import theme from "../theme";

// PRESET_COLORS mirrors cwclock-ui's project color picker's general palette
// intent (a handful of distinct, readable badge colors) - the web app uses a
// free-form <input type="color">, which has no RN equivalent without an
// extra native dependency, so this is a fixed swatch row instead.
const PRESET_COLORS = [
  "#1cb9f7",
  "#f76c1c",
  "#1cf7a0",
  "#f71c6c",
  "#7c1cf7",
  "#f7d11c",
  "#1c3ff7",
  "#8a8a8a",
];

const ColorSwatchPicker = ({ label, value, onChange }) => (
  <View style={styles.container}>
    {label ? <Text style={styles.label}>{label}</Text> : null}
    <View style={styles.row}>
      {PRESET_COLORS.map((color) => (
        <TouchableOpacity
          key={color}
          onPress={() => onChange(color)}
          style={[styles.swatch, { backgroundColor: color }, value === color && styles.swatchSelected]}
        />
      ))}
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
  row: {
    flexDirection: "row",
    flexWrap: "wrap",
    gap: theme.spacing(1),
  },
  swatch: {
    width: 32,
    height: 32,
    borderRadius: 16,
  },
  swatchSelected: {
    borderWidth: 3,
    borderColor: theme.color.text,
  },
});

export default ColorSwatchPicker;
