import React from "react";
import { KeyboardAvoidingView, Platform, ScrollView, StyleSheet, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import theme from "../theme";

// Screen is the shared page wrapper: safe-area + keyboard-avoiding + a
// scrollable, consistently-padded body. scroll=false opts out for screens
// that manage their own scrolling (e.g. a FlatList-based list screen).
const Screen = ({ children, scroll = true, style }) => {
  const body = scroll ? (
    <ScrollView contentContainerStyle={styles.scrollContent} keyboardShouldPersistTaps="handled">
      {children}
    </ScrollView>
  ) : (
    <View style={styles.flex}>{children}</View>
  );

  return (
    <SafeAreaView style={[styles.safeArea, style]} edges={["top", "left", "right"]}>
      <KeyboardAvoidingView style={styles.flex} behavior={Platform.OS === "ios" ? "padding" : undefined}>
        {body}
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: theme.color.background,
  },
  flex: {
    flex: 1,
  },
  scrollContent: {
    padding: theme.spacing(2),
    flexGrow: 1,
  },
});

export default Screen;
