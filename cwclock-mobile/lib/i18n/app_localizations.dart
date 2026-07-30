import 'en.dart';
import 'fr.dart';

/// Minimal i18n: a flat-ish nested dictionary per locale plus a lookup with
/// `{{var}}` interpolation and an English fallback for missing keys - ported
/// from the RN app's src/i18n/translate.js, itself ported from cwclock-ui.
const Map<String, Map<String, dynamic>> dictionaries = {'en': en, 'fr': fr};

const List<Locale> supportedLocales = [
  Locale('en', 'English'),
  Locale('fr', 'Français'),
];

class Locale {
  final String code;
  final String label;

  const Locale(this.code, this.label);
}

dynamic resolve(Map<String, dynamic>? dict, String key) {
  dynamic acc = dict;
  for (final part in key.split('.')) {
    if (acc is Map<String, dynamic> && acc.containsKey(part)) {
      acc = acc[part];
    } else {
      return null;
    }
  }
  return acc;
}

String translate(String locale, String key, [Map<String, String>? vars]) {
  var str = (resolve(dictionaries[locale], key) ?? resolve(dictionaries['en'], key) ?? key)
      .toString();
  if (vars != null) {
    for (final entry in vars.entries) {
      str = str.replaceAll('{{${entry.key}}}', entry.value);
    }
  }
  return str;
}

/// Implemented by [ApiException] (api/api_exception.dart) - kept here rather
/// than importing that file, to avoid a circular dependency between api/ and
/// i18n/.
abstract class ApiErrorLike {
  String? get i18nCode;
  String? get message;
}

/// Translates a failed request for display, whichever request path it came
/// from: an API JSON error body (`i18n_code`/`message`, same shape
/// cwclock-api sends everywhere) or a plain exception message.
String apiErrorMessage(Object error, String locale) {
  String? i18nCode;
  String? message;

  if (error is ApiErrorLike) {
    i18nCode = error.i18nCode;
    message = error.message;
  } else {
    message = error.toString();
  }

  if (i18nCode != null) {
    final translated = resolve(dictionaries[locale] ?? dictionaries['en'], i18nCode);
    if (translated != null) return translated.toString();
  }

  return message ?? translate(locale, 'errors.network');
}
