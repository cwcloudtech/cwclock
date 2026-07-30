import React, { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import FormField from "../../components/FormField";
import SelectField from "../../components/SelectField";
import Button from "../../components/Button";
import ErrorBanner from "../../components/ErrorBanner";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import { addMemberApi } from "../../redux/organizations/organizations.actions";

const InviteMemberScreen = ({ navigation }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { orgId } = useSelector((state) => state.session);

  const [email, setEmail] = useState("");
  const [role, setRole] = useState("member");
  const [error, setError] = useState(null);
  const [saving, setSaving] = useState(false);

  const handleInvite = async () => {
    setError(null);
    setSaving(true);
    try {
      await dispatch(addMemberApi(orgId, email.trim(), role));
      navigation.goBack();
    } catch (e) {
      setError(apiErrorMessage(e, locale));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Screen>
      <FormField
        label={t("organizations.email")}
        value={email}
        onChangeText={setEmail}
        autoCapitalize="none"
        autoCorrect={false}
        keyboardType="email-address"
      />
      <SelectField
        label={t("organizations.role")}
        value={role}
        onValueChange={setRole}
        items={[
          { value: "admin", label: t("organizations.roleAdmin") },
          { value: "member", label: t("organizations.roleMember") },
          { value: "reader", label: t("organizations.roleReader") },
        ]}
      />
      <ErrorBanner message={error} />
      <Button title={t("organizations.inviteMember")} onPress={handleInvite} disabled={!email.trim()} loading={saving} />
    </Screen>
  );
};

export default InviteMemberScreen;
