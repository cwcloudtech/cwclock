import React, { useEffect } from "react";
import { FlatList, StyleSheet, Text, TouchableOpacity } from "react-native";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import { useI18n } from "../../i18n/I18nContext";
import { listOrganizationsApi } from "../../redux/organizations/organizations.actions";
import { selectOrgApi } from "../../redux/session/session.actions";
import theme from "../../theme";

// OrgPickerScreen shows whenever session.status is "needsOrg" (see
// session.reducer.js) - either onboarding didn't have an org_id line
// (manual entry / an older QR), or the user picked "Switch organization"
// from SettingsScreen. Selecting a row dispatches selectOrgApi and the root
// navigator (src/App.js) reacts to the resulting status change itself.
const OrgPickerScreen = ({ navigation }) => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const { items, isLoading } = useSelector((state) => state.organizations);

  useEffect(() => {
    dispatch(listOrganizationsApi());
  }, [dispatch]);

  const handleSelect = async (orgId) => {
    await dispatch(selectOrgApi(orgId));
    // No-op when there's nothing to go back to (first-run onboarding) -
    // only meaningful when reached from SettingsScreen's "Switch
    // organization", where session.status was already "connected".
    navigation.goBack();
  };

  return (
    <Screen scroll={false}>
      <Text style={styles.title}>{t("onboarding.pickOrganization")}</Text>
      <FlatList
        data={items}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.list}
        ListEmptyComponent={!isLoading ? <Text style={styles.empty}>{t("onboarding.noOrganizations")}</Text> : null}
        renderItem={({ item }) => (
          <TouchableOpacity style={styles.row} onPress={() => handleSelect(item.id)}>
            <Text style={styles.rowText}>{item.name}</Text>
          </TouchableOpacity>
        )}
      />
    </Screen>
  );
};

const styles = StyleSheet.create({
  title: {
    fontSize: 20,
    fontWeight: "700",
    color: theme.color.text,
    padding: theme.spacing(2),
  },
  list: {
    paddingHorizontal: theme.spacing(2),
  },
  row: {
    paddingVertical: theme.spacing(2),
    borderBottomWidth: 1,
    borderBottomColor: theme.color.border,
  },
  rowText: {
    fontSize: 17,
    color: theme.color.text,
  },
  empty: {
    color: theme.color.textMuted,
    padding: theme.spacing(2),
  },
});

export default OrgPickerScreen;
