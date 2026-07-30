import React, { useCallback, useEffect, useState } from "react";
import { PermissionsAndroid, StyleSheet, Text, View } from "react-native";
import { Camera } from "react-native-camera-kit";
import { useDispatch } from "react-redux";
import Screen from "../../components/Screen";
import ErrorBanner from "../../components/ErrorBanner";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import { parseConfigText, isCompleteConfig } from "../../api/config";
import { connectApi } from "../../redux/session/session.actions";
import theme from "../../theme";

// ScanQrScreen reads the same "key = value" config text the CLI's
// `cwclock configure import` and the web ApiKeys page's QR/file downloads
// use (see src/api/config.js's parseConfigText) - react-native-camera-kit's
// onReadCode fires with the raw decoded string, no separate QR library
// needed since a QR code is just a container for that text.
const ScanQrScreen = () => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const [hasPermission, setHasPermission] = useState(false);
  const [error, setError] = useState(null);
  const [connecting, setConnecting] = useState(false);

  useEffect(() => {
    PermissionsAndroid.request(PermissionsAndroid.PERMISSIONS.CAMERA).then((result) => {
      setHasPermission(result === PermissionsAndroid.RESULTS.GRANTED);
    });
  }, []);

  const handleReadCode = useCallback(
    async (event) => {
      if (connecting) return;
      const raw = event?.nativeEvent?.codeStringValue;
      const parsed = parseConfigText(raw);
      if (!isCompleteConfig(parsed)) {
        setError(t("onboarding.scanFailed"));
        return;
      }

      setConnecting(true);
      setError(null);
      try {
        // Session status flips to "needsOrg" or "connected"
        // (session.reducer.js) - the root navigator (src/App.js) reacts to
        // that and moves on by itself, nothing to navigate to here.
        await dispatch(connectApi(parsed));
      } catch (e) {
        setError(apiErrorMessage(e, locale) || t("onboarding.invalidCredentials"));
      } finally {
        setConnecting(false);
      }
    },
    [connecting, dispatch, locale, t]
  );

  if (!hasPermission) {
    return (
      <Screen>
        <Text style={styles.hint}>{t("onboarding.scanHint")}</Text>
      </Screen>
    );
  }

  return (
    <View style={styles.flex}>
      <Camera
        style={styles.flex}
        scanBarcode
        onReadCode={handleReadCode}
        showFrame
        frameColor={theme.color.primary}
        laserColor={theme.color.primary}
      />
      <View style={styles.overlay}>
        <Text style={styles.hint}>{t("onboarding.scanHint")}</Text>
        <ErrorBanner message={error} />
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  flex: { flex: 1 },
  overlay: {
    position: "absolute",
    bottom: 0,
    left: 0,
    right: 0,
    padding: theme.spacing(2),
    backgroundColor: "rgba(0,0,0,0.55)",
  },
  hint: {
    color: theme.color.white,
    textAlign: "center",
    fontSize: 14,
  },
});

export default ScanQrScreen;
