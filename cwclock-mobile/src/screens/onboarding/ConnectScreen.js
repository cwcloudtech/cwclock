import React from "react";
import { StyleSheet, Text, View } from "react-native";
import Screen from "../../components/Screen";
import Button from "../../components/Button";
import { useI18n } from "../../i18n/I18nContext";
import theme from "../../theme";

// ConnectScreen is the onboarding entry point (see the ai-instruct-105 plan
// - "ask to scan the QR code or the api key"): two ways into the same
// underlying flow, both ending at redux/session/session.actions.js's
// connectApi. Session status (src/redux/session/session.reducer.js) drives
// which screen the root navigator shows next (see src/App.js) - neither
// button here navigates directly past onboarding itself.
const ConnectScreen = ({ navigation }) => {
  const { t } = useI18n();

  return (
    <Screen>
      <View style={styles.header}>
        <Text style={styles.title}>{t("onboarding.title")}</Text>
        <Text style={styles.intro}>{t("onboarding.intro")}</Text>
      </View>
      <Button title={t("onboarding.scanQr")} onPress={() => navigation.navigate("ScanQr")} style={styles.button} />
      <Button
        title={t("onboarding.enterManually")}
        variant="secondary"
        onPress={() => navigation.navigate("ManualEntry")}
      />
    </Screen>
  );
};

const styles = StyleSheet.create({
  header: {
    marginTop: theme.spacing(6),
    marginBottom: theme.spacing(4),
  },
  title: {
    fontSize: 24,
    fontWeight: "700",
    color: theme.color.text,
    marginBottom: theme.spacing(1),
  },
  intro: {
    fontSize: 15,
    color: theme.color.textMuted,
  },
  button: {
    marginBottom: theme.spacing(2),
  },
});

export default ConnectScreen;
