import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../api/api_client.dart';
import '../api/pdf_client.dart';
import '../storage/local_storage.dart';
import '../storage/secure_storage.dart';
import '../storage/timer_storage.dart';

/// Shared singletons - one ApiClient instance for the whole app's lifetime,
/// same as src/api/client.js's module-scope `client`.
final apiClientProvider = Provider<ApiClient>((ref) => ApiClient());

final pdfClientProvider = Provider<PdfClient>((ref) => PdfClient(ref.read(apiClientProvider)));

final secureStorageProvider = Provider<SecureStorage>((ref) => SecureStorage());

final localStorageProvider = Provider<LocalStorage>((ref) => LocalStorage());

final timerStorageProvider = Provider<TimerStorage>((ref) => TimerStorage());
