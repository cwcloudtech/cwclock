import React from "react";
import { StyleSheet, Text } from "react-native";
import theme from "../theme";

const ErrorBanner = ({ message }) => {
  if (!message) return null;
  return <Text style={styles.text}>{message}</Text>;
};

const styles = StyleSheet.create({
  text: {
    color: theme.color.danger,
    marginBottom: theme.spacing(1.5),
    fontSize: 14,
  },
});

export default ErrorBanner;
