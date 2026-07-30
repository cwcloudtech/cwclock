import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// The API key is a bearer secret (same trust level as a password), so it
/// lives in Android Keystore-backed secure storage - ported from
/// src/storage/session.js's Keychain.setGenericPassword/getGenericPassword/
/// resetGenericPassword calls under a named service.
class SecureStorage {
  static const _apiKeyKey = 'me.cwclock.mobile.apiKey';

  final FlutterSecureStorage _storage = const FlutterSecureStorage();

  Future<void> setApiKey(String apiKey) => _storage.write(key: _apiKeyKey, value: apiKey);

  Future<String?> getApiKey() => _storage.read(key: _apiKeyKey);

  Future<void> clearApiKey() => _storage.delete(key: _apiKeyKey);
}
