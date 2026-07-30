import 'package:shared_preferences/shared_preferences.dart';

/// apiUrl/orgId/locale are just local preferences (not secrets) - ported from
/// src/storage/session.js's AsyncStorage calls and src/i18n/translate.js's
/// stored-locale key.
class LocalStorage {
  static const _apiUrlKey = 'cwclock.apiUrl';
  static const _orgIdKey = 'cwclock.orgId';
  static const _localeKey = 'cwclock.locale';

  Future<String?> getApiUrl() async => (await SharedPreferences.getInstance()).getString(_apiUrlKey);

  Future<void> setApiUrl(String apiUrl) async =>
      (await SharedPreferences.getInstance()).setString(_apiUrlKey, apiUrl);

  Future<String?> getOrgId() async => (await SharedPreferences.getInstance()).getString(_orgIdKey);

  Future<void> setOrgId(String orgId) async =>
      (await SharedPreferences.getInstance()).setString(_orgIdKey, orgId);

  Future<void> clearOrgId() async => (await SharedPreferences.getInstance()).remove(_orgIdKey);

  Future<String?> getLocale() async => (await SharedPreferences.getInstance()).getString(_localeKey);

  Future<void> setLocale(String locale) async =>
      (await SharedPreferences.getInstance()).setString(_localeKey, locale);

  Future<void> clearAll() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_apiUrlKey);
    await prefs.remove(_orgIdKey);
  }
}
