import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:path_provider/path_provider.dart';

import 'api_client.dart';
import 'api_exception.dart';

/// Issues a binary request for a report/invoice PDF and returns the local
/// file path it was cached to - ported from src/api/blobRequest.js's
/// fetchPdf, which used react-native-blob-util instead of axios/dio for the
/// same reason this uses dio's `responseType: bytes` rather than a generic
/// JSON call: streaming a large PDF response straight to a file, not through
/// JSON parsing.
class PdfClient {
  final ApiClient apiClient;

  PdfClient(this.apiClient);

  Future<String> fetchPdf(String method, String path, [Map<String, dynamic>? body]) async {
    late Response<List<int>> response;
    try {
      response = await apiClient.dio.request<List<int>>(
        '${apiClient.baseUrl}$path',
        data: body,
        options: Options(
          method: method,
          headers: {'Content-Type': 'application/json', ...apiClient.headers},
          responseType: ResponseType.bytes,
          validateStatus: (status) => true,
        ),
      );
    } on DioException catch (e) {
      throw asApiException(e);
    }

    final status = response.statusCode ?? 0;
    if (status >= 400) {
      String message = 'Request failed (HTTP $status)';
      String? i18nCode;
      try {
        final parsed = jsonDecode(utf8.decode(response.data ?? []));
        i18nCode = parsed?['i18n_code'] as String?;
        message = parsed?['message'] as String? ?? message;
      } catch (_) {
        // Response wasn't JSON (unexpected for an error) - keep the generic message.
      }
      throw ApiException(i18nCode: i18nCode, message: message, statusCode: status);
    }

    final dir = await getTemporaryDirectory();
    final file = File('${dir.path}/${DateTime.now().microsecondsSinceEpoch}.pdf');
    await file.writeAsBytes(response.data ?? []);
    return file.path;
  }
}
