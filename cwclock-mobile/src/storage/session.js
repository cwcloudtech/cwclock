import * as Keychain from "react-native-keychain";
import AsyncStorage from "@react-native-async-storage/async-storage";

const KEYCHAIN_SERVICE = "me.cwclock.mobile.apiKey";
const API_URL_STORAGE_KEY = "cwclock.apiUrl";
const ORG_ID_STORAGE_KEY = "cwclock.orgId";

// The API key is a bearer secret (same trust level as a password), so it
// lives in the Android Keystore-backed Keychain; apiUrl/orgId are just local
// preferences (same split cwclock-cli makes by keeping everything in one
// plain ~/.cwclock/config file, but mobile has a real secure-storage option
// the CLI's plaintext file doesn't).
export const saveSession = async ({ apiUrl, apiKey, orgId }) => {
  await Keychain.setGenericPassword("cwclock", apiKey, { service: KEYCHAIN_SERVICE });
  await AsyncStorage.setItem(API_URL_STORAGE_KEY, apiUrl);
  if (orgId) {
    await AsyncStorage.setItem(ORG_ID_STORAGE_KEY, orgId);
  } else {
    await AsyncStorage.removeItem(ORG_ID_STORAGE_KEY);
  }
};

export const saveOrgId = async (orgId) => {
  await AsyncStorage.setItem(ORG_ID_STORAGE_KEY, orgId);
};

// loadSession restores a previously saved session on app boot, or returns
// null when nothing (or an incomplete pair) is stored - callers should
// treat that as "show onboarding".
export const loadSession = async () => {
  const credentials = await Keychain.getGenericPassword({ service: KEYCHAIN_SERVICE });
  if (!credentials) return null;

  const apiUrl = await AsyncStorage.getItem(API_URL_STORAGE_KEY);
  if (!apiUrl) return null;

  const orgId = await AsyncStorage.getItem(ORG_ID_STORAGE_KEY);
  return { apiUrl, apiKey: credentials.password, orgId: orgId || null };
};

export const clearSession = async () => {
  await Keychain.resetGenericPassword({ service: KEYCHAIN_SERVICE });
  await AsyncStorage.multiRemove([API_URL_STORAGE_KEY, ORG_ID_STORAGE_KEY]);
};
