import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../i18n/app_localizations.dart';
import 'api_providers.dart';

/// Ported from src/i18n/I18nContext.jsx - starts as "en" for the very first
/// render, then loads the persisted locale asynchronously (shared_preferences
/// is async, unlike the web app's synchronous localStorage).
class LocaleNotifier extends Notifier<String> {
  @override
  String build() => 'en';

  Future<void> load() async {
    final stored = await ref.read(localStorageProvider).getLocale();
    if (stored != null && dictionaries.containsKey(stored)) {
      state = stored;
    }
  }

  Future<void> setLocale(String next) async {
    if (!dictionaries.containsKey(next)) return;
    await ref.read(localStorageProvider).setLocale(next);
    state = next;
  }
}

final localeProvider = NotifierProvider<LocaleNotifier, String>(LocaleNotifier.new);

/// Shorthand: `ref.watch(tProvider)('key.path', {'var': 'value'})`.
String Function(String, [Map<String, String>?]) translateWith(String locale) {
  return (key, [vars]) => translate(locale, key, vars);
}
