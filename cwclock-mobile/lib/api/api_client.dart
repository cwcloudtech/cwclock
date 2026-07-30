import 'package:dio/dio.dart';

import 'api_exception.dart';

/// A single shared HTTP client - base URL `${apiUrl}/v1`, `X-Api-Key` header
/// injected on every request (simple API-key auth, no OAuth/bearer refresh).
/// Ported from src/api/client.js; `getBaseUrl()`/`getClientHeaders()`'s role
/// there (letting the non-dio PDF fetch path share the same session) is
/// covered here by [baseUrl]/[headers], used by PdfClient.
class ApiClient {
  final Dio dio = Dio();

  String _apiUrl = '';
  String _apiKey = '';
  String _orgId = '';

  ApiClient() {
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          if (_apiKey.isNotEmpty) {
            options.headers['X-Api-Key'] = _apiKey;
          }
          handler.next(options);
        },
        onError: (error, handler) {
          handler.reject(_toApiError(error));
        },
      ),
    );
  }

  String get orgId => _orgId;

  String get baseUrl => dio.options.baseUrl;

  Map<String, String> get headers => _apiKey.isNotEmpty ? {'X-Api-Key': _apiKey} : {};

  void setSession({String? apiUrl, String? apiKey, String? orgId}) {
    if (apiUrl != null) _apiUrl = apiUrl;
    if (apiKey != null) _apiKey = apiKey;
    if (orgId != null) _orgId = orgId;
    dio.options.baseUrl = _apiUrl.isNotEmpty
        ? '${_apiUrl.replaceFirst(RegExp(r'/+$'), '')}/v1'
        : '';
  }

  void clearSession() {
    _apiUrl = '';
    _apiKey = '';
    _orgId = '';
    dio.options.baseUrl = '';
  }

  DioException _toApiError(DioException error) {
    final response = error.response;
    if (response == null) {
      return error.copyWith(
        error: ApiException(message: error.message),
      );
    }
    final data = response.data;
    String? i18nCode;
    String? message;
    if (data is Map) {
      i18nCode = data['i18n_code'] as String?;
      message = data['message'] as String?;
    }
    return error.copyWith(
      error: ApiException(
        i18nCode: i18nCode,
        message: message,
        statusCode: response.statusCode,
      ),
    );
  }
}

/// Unwraps the [ApiException] a request's failure carries, whether it came
/// from [ApiClient]'s interceptor (a [DioException] wrapping one) or was
/// thrown directly (PdfClient).
ApiException asApiException(Object error) {
  if (error is ApiException) return error;
  if (error is DioException && error.error is ApiException) {
    return error.error as ApiException;
  }
  return ApiException(message: error.toString());
}
