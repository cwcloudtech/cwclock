import React, { useEffect } from "react";
import { Alert, StyleSheet, Text } from "react-native";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import Button from "../../components/Button";
import { useI18n, LANGUAGES } from "../../i18n/I18nContext";
import theme from "../../theme";
import { listOrganizationsApi } from "../../redux/organizations/organizations.actions";
import { disconnectApi } from "../../redux/session/session.actions";

const SettingsScreen = ({ navigation }) => {
  const { t, locale, setLocale } = useI18n();
  const dispatch = useDispatch();
  const { apiUrl, orgId, user } = useSelector((state) => state.session);
  const { items: orgs } = useSelector((state) => state.organizations);

  useEffect(() => {
    dispatch(listOrganizationsApi());
  }, [dispatch]);

  const currentOrg = orgs.find((o) => o.id === orgId);

  const handleDisconnect = () => {
    Alert.alert(t("settings.disconnect"), t("settings.disconnectConfirm"), [
      { text: t("common.cancel"), style: "cancel" },
      { text: t("settings.disconnect"), style: "destructive", onPress: () => dispatch(disconnectApi()) },
    ]);
  };

  return (
    <Screen>
      <Text style={styles.label}>{t("settings.connectedTo")}</Text>
      <Text style={styles.value}>{apiUrl}</Text>

      <Text style={styles.label}>{t("settings.organization")}</Text>
      <Text style={styles.value}>{currentOrg?.name || orgId}</Text>

      {user?.email ? (
        <>
          <Text style={styles.label}>{user.email}</Text>
        </>
      ) : null}

      <Button
        title={t("settings.switchOrganization")}
        variant="secondary"
        onPress={() => navigation.navigate("SwitchOrganization")}
        style={styles.button}
      />

      <Button
        title={LANGUAGES.find((l) => l.code === locale)?.label || "English"}
        variant="secondary"
        onPress={() => setLocale(locale === "en" ? "fr" : "en")}
        style={styles.button}
      />

      <Button title={t("settings.disconnect")} variant="danger" onPress={handleDisconnect} />
    </Screen>
  );
};

const styles = StyleSheet.create({
  label: {
    fontSize: 13,
    color: theme.color.textMuted,
    marginTop: theme.spacing(1.5),
  },
  value: {
    fontSize: 17,
    color: theme.color.text,
    marginTop: 2,
  },
  button: {
    marginTop: theme.spacing(3),
  },
});

export default SettingsScreen;
