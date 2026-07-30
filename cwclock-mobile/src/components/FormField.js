import React from "react";
import { StyleSheet, Text, TextInput, View } from "react-native";
import theme from "../theme";

const FormField = ({ label, style, inputStyle, ...inputProps }) => (
  <View style={[styles.container, style]}>
    {label ? <Text style={styles.label}>{label}</Text> : null}
    <TextInput placeholderTextColor={theme.color.textMuted} style={[styles.input, inputStyle]} {...inputProps} />
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
  input: {
    borderWidth: 1,
    borderColor: theme.color.border,
    borderRadius: theme.radius,
    paddingHorizontal: theme.spacing(1.5),
    paddingVertical: theme.spacing(1.25),
    fontSize: 16,
    color: theme.color.text,
  },
});

export default FormField;
