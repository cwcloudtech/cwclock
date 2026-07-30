import '../i18n/app_localizations.dart';

/// A failed request for display, whichever request path it came from - a
/// JSON error body ({i18n_code, message}, the same shape cwclock-api sends
/// everywhere) or a plain network failure. Ported from src/api/blobRequest.js
/// (PDF path) and the axios error shape src/i18n/translate.js's
/// apiErrorMessage reads for every other request.
class ApiException implements Exception, ApiErrorLike {
  @override
  final String? i18nCode;
  @override
  final String? message;
  final int? statusCode;

  const ApiException({this.i18nCode, this.message, this.statusCode});

  @override
  String toString() => message ?? 'Request failed${statusCode != null ? ' (HTTP $statusCode)' : ''}';
}
