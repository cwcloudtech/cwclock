import React, { useEffect } from "react";
import { Alert, StyleSheet, Text, View } from "react-native";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import Button from "../../components/Button";
import { useI18n, LANGUAGES } from "../../i18n/I18nContext";
import theme from "../../theme";
import { isAdminOrOwner } from "../../common/permissions";
import { listOrganizationsApi, listMembersApi } from "../../redux/organizations/organizations.actions";
import { disconnectApi } from "../../redux/session/session.actions";

const SettingsScreen = ({ navigation }) => {
  const { t, locale, setLocale } = useI18n();
  const dispatch = useDispatch();
  const { apiUrl, orgId, user } = useSelector((state) => state.session);
  const { items: orgs, members } = useSelector((state) => state.organizations);

  useEffect(() => {
    dispatch(listOrganizationsApi());
  }, [dispatch]);

  useEffect(() => {
    if (orgId) dispatch(listMembersApi(orgId));
  }, [orgId, dispatch]);

  const currentOrg = orgs.find((o) => o.id === orgId);
  const canManage = isAdminOrOwner(user, members);

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

      {/* Organization/client/project management (ai-instruct-106) - gated to
          admin/owner, matching the same bar the create/edit/delete endpoints
          themselves enforce server-side (see the *.actions.js comments in
          each screen). A plain member/reader wouldn't have any actionable
          capability behind these screens besides viewing, so the entry
          points are hidden rather than shown read-only. */}
      {canManage && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>{t("management.title")}</Text>
          <Button
            title={t("management.organization")}
            variant="secondary"
            onPress={() => navigation.navigate("Organization")}
            style={styles.button}
          />
          <Button
            title={t("management.members")}
            variant="secondary"
            onPress={() => navigation.navigate("Members")}
            style={styles.button}
          />
          <Button
            title={t("management.clients")}
            variant="secondary"
            onPress={() => navigation.navigate("Clients")}
            style={styles.button}
          />
          <Button
            title={t("management.projects")}
            variant="secondary"
            onPress={() => navigation.navigate("Projects")}
            style={styles.button}
          />
        </View>
      )}

      <Button title={t("settings.disconnect")} variant="danger" onPress={handleDisconnect} style={styles.disconnectButton} />
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
    marginTop: theme.spacing(1.5),
  },
  section: {
    marginTop: theme.spacing(3),
    paddingTop: theme.spacing(2),
    borderTopWidth: 1,
    borderTopColor: theme.color.border,
  },
  sectionTitle: {
    fontSize: 14,
    fontWeight: "600",
    color: theme.color.textMuted,
  },
  disconnectButton: {
    marginTop: theme.spacing(3),
  },
});

export default SettingsScreen;
