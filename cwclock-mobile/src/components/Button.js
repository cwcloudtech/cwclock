import React from "react";
import { ActivityIndicator, StyleSheet, Text, TouchableOpacity } from "react-native";
import theme from "../theme";

const VARIANT_STYLES = {
  primary: { backgroundColor: theme.color.primary, textColor: theme.color.white },
  secondary: { backgroundColor: theme.color.backgroundMuted, textColor: theme.color.text },
  danger: { backgroundColor: theme.color.danger, textColor: theme.color.white },
};

const Button = ({ title, onPress, variant = "primary", disabled, loading, style }) => {
  const { backgroundColor, textColor } = VARIANT_STYLES[variant] || VARIANT_STYLES.primary;

  return (
    <TouchableOpacity
      onPress={onPress}
      disabled={disabled || loading}
      style={[styles.button, { backgroundColor }, (disabled || loading) && styles.disabled, style]}
    >
      {loading ? <ActivityIndicator color={textColor} /> : <Text style={[styles.text, { color: textColor }]}>{title}</Text>}
    </TouchableOpacity>
  );
};

const styles = StyleSheet.create({
  button: {
    paddingVertical: theme.spacing(1.5),
    paddingHorizontal: theme.spacing(2),
    borderRadius: theme.radius,
    alignItems: "center",
    justifyContent: "center",
  },
  disabled: {
    opacity: 0.5,
  },
  text: {
    fontSize: 16,
    fontWeight: "600",
  },
});

export default Button;
