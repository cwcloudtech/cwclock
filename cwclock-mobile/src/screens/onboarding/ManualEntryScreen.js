import React, { useState } from "react";
import { useDispatch } from "react-redux";
import Screen from "../../components/Screen";
import FormField from "../../components/FormField";
import Button from "../../components/Button";
import ErrorBanner from "../../components/ErrorBanner";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import { parseConfigText } from "../../api/config";
import { connectApi } from "../../redux/session/session.actions";

// ManualEntryScreen covers onboarding without a camera: either paste the
// whole downloaded config file's text (parsed the same way as the QR, see
// src/api/config.js) to fill the three fields at once, or type them in
// directly - both end at the same connectApi call as ScanQrScreen.
const ManualEntryScreen = () => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const [pasteText, setPasteText] = useState("");
  const [apiUrl, setApiUrl] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [orgId, setOrgId] = useState("");
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  const handleParsePaste = () => {
    const parsed = parseConfigText(pasteText);
    if (parsed.apiUrl) setApiUrl(parsed.apiUrl);
    if (parsed.apiKey) setApiKey(parsed.apiKey);
    if (parsed.orgId) setOrgId(parsed.orgId);
  };

  const handleConnect = async () => {
    setError(null);
    setLoading(true);
    try {
      await dispatch(connectApi({ apiUrl: apiUrl.trim(), apiKey: apiKey.trim(), orgId: orgId.trim() }));
    } catch (e) {
      setError(apiErrorMessage(e, locale) || t("onboarding.invalidCredentials"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Screen>
      <FormField
        label={t("onboarding.pasteConfig")}
        value={pasteText}
        onChangeText={setPasteText}
        onBlur={handleParsePaste}
        multiline
        numberOfLines={3}
        autoCapitalize="none"
        placeholder={"api_url = ...\napi_key = ...\norg_id = ..."}
      />
      <FormField
        label={t("onboarding.apiUrl")}
        value={apiUrl}
        onChangeText={setApiUrl}
        autoCapitalize="none"
        autoCorrect={false}
        keyboardType="url"
        placeholder="https://api.cwclock.me"
      />
      <FormField
        label={t("onboarding.apiKey")}
        value={apiKey}
        onChangeText={setApiKey}
        autoCapitalize="none"
        autoCorrect={false}
        secureTextEntry
      />
      <FormField label={t("onboarding.orgIdOptional")} value={orgId} onChangeText={setOrgId} autoCapitalize="none" />
      <ErrorBanner message={error} />
      <Button
        title={t("onboarding.connect")}
        onPress={handleConnect}
        loading={loading}
        disabled={!apiUrl.trim() || !apiKey.trim()}
      />
    </Screen>
  );
};

export default ManualEntryScreen;
