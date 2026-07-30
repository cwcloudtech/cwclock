import React, { useEffect } from "react";
import { Alert, FlatList, StyleSheet, Text, TouchableOpacity, View } from "react-native";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import Button from "../../components/Button";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import theme from "../../theme";
import { isOrgOwner, isAdminOrOwner } from "../../common/permissions";
import { listMembersApi, updateMemberRoleApi, removeMemberApi } from "../../redux/organizations/organizations.actions";

const ROLES = ["admin", "member", "reader"];

// MembersScreen: the "user RBAC in the organization" feature from
// ai-instruct-106 - list members with their role, invite (owner-only,
// InviteMemberScreen), change an existing member's role (owner-only - the
// API supports this via UpdateMember, but cwclock-ui's web app has no UI for
// it, only invite-time role, see organizations.actions.js's
// updateMemberRoleApi comment) and remove a member (admin/owner). The
// owner's own row exposes neither action - removing/demoting the owner
// isn't a normal RBAC operation.
const MembersScreen = ({ navigation }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { orgId, user } = useSelector((state) => state.session);
  const { items: orgs, members } = useSelector((state) => state.organizations);

  const org = orgs.find((o) => o.id === orgId);
  const canInvite = isOrgOwner(user, org);
  const canRemove = isAdminOrOwner(user, members);

  useEffect(() => {
    if (orgId) dispatch(listMembersApi(orgId));
  }, [orgId, dispatch]);

  const roleLabel = (role) => t(`organizations.role${role.charAt(0).toUpperCase()}${role.slice(1)}`);

  const handleChangeRole = (member) => {
    Alert.alert(
      t("organizations.changeRole"),
      member.email,
      ROLES.map((role) => ({
        text: roleLabel(role),
        onPress: () => dispatch(updateMemberRoleApi(orgId, member.userId, role)).catch((e) => Alert.alert(t("common.retry"), apiErrorMessage(e, locale))),
      })).concat([{ text: t("common.cancel"), style: "cancel" }])
    );
  };

  const handleRemove = (member) => {
    Alert.alert(t("organizations.removeMemberTitle"), t("organizations.removeMemberBody", { name: member.email }), [
      { text: t("common.cancel"), style: "cancel" },
      {
        text: t("organizations.removeMember"),
        style: "destructive",
        onPress: () =>
          dispatch(removeMemberApi(orgId, member.userId)).catch((e) =>
            Alert.alert(t("common.retry"), apiErrorMessage(e, locale))
          ),
      },
    ]);
  };

  return (
    <Screen scroll={false}>
      {canInvite && (
        <View style={styles.header}>
          <Button title={t("organizations.inviteMember")} onPress={() => navigation.navigate("InviteMember")} />
        </View>
      )}
      <FlatList
        data={members}
        keyExtractor={(m) => m.userId}
        contentContainerStyle={styles.list}
        ListEmptyComponent={<Text style={styles.empty}>{t("organizations.noMembers")}</Text>}
        renderItem={({ item }) => {
          const isOwnerRow = item.role === "owner";
          return (
            <View style={styles.row}>
              <View style={styles.rowMain}>
                <Text style={styles.rowText} numberOfLines={1}>
                  {item.name || item.surname ? `${item.name} ${item.surname}`.trim() : item.email}
                </Text>
                <Text style={styles.rowSub} numberOfLines={1}>
                  {item.email} · {roleLabel(item.role)}
                </Text>
              </View>
              {!isOwnerRow && (
                <View style={styles.rowActions}>
                  {canInvite && (
                    <TouchableOpacity onPress={() => handleChangeRole(item)} style={styles.actionBtn}>
                      <Text style={styles.actionText}>{t("organizations.changeRole")}</Text>
                    </TouchableOpacity>
                  )}
                  {canRemove && (
                    <TouchableOpacity onPress={() => handleRemove(item)} style={styles.actionBtn}>
                      <Text style={styles.actionTextDanger}>{t("organizations.removeMember")}</Text>
                    </TouchableOpacity>
                  )}
                </View>
              )}
            </View>
          );
        }}
      />
    </Screen>
  );
};

const styles = StyleSheet.create({
  header: {
    padding: theme.spacing(2),
  },
  list: {
    paddingHorizontal: theme.spacing(2),
  },
  row: {
    paddingVertical: theme.spacing(1.5),
    borderBottomWidth: 1,
    borderBottomColor: theme.color.border,
  },
  rowMain: {
    marginBottom: theme.spacing(0.5),
  },
  rowText: {
    fontSize: 16,
    color: theme.color.text,
  },
  rowSub: {
    fontSize: 13,
    color: theme.color.textMuted,
    marginTop: 2,
  },
  rowActions: {
    flexDirection: "row",
    gap: theme.spacing(2),
  },
  actionBtn: {
    paddingVertical: theme.spacing(0.5),
  },
  actionText: {
    color: theme.color.primary,
    fontSize: 13,
    fontWeight: "600",
  },
  actionTextDanger: {
    color: theme.color.danger,
    fontSize: 13,
    fontWeight: "600",
  },
  empty: {
    color: theme.color.textMuted,
    paddingVertical: theme.spacing(2),
  },
});

export default MembersScreen;
