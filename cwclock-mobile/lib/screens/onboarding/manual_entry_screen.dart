import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../models/session_config.dart';
import '../../providers/locale_provider.dart';
import '../../providers/session_provider.dart';
import '../../widgets/app_button.dart';
import '../../widgets/app_screen.dart';
import '../../widgets/error_banner.dart';
import '../../widgets/form_field.dart';

/// Covers onboarding without a camera: type the API URL/key (and optional
/// org ID) directly - ends at the same SessionNotifier.connect call as
/// ScanQrScreen. Ported from src/screens/onboarding/ManualEntryScreen.js;
/// per ai-instruct-119, the free-form "paste the whole config text" field
/// was dropped in favor of just the separate fields, with apiUrl pre-filled
/// to the production API.
class ManualEntryScreen extends ConsumerStatefulWidget {
  const ManualEntryScreen({super.key});

  @override
  ConsumerState<ManualEntryScreen> createState() => _ManualEntryScreenState();
}

class _ManualEntryScreenState extends ConsumerState<ManualEntryScreen> {
  String _apiUrl = 'https://api.cwclock.me';
  String _apiKey = '';
  String _orgId = '';
  String? _error;
  bool _loading = false;

  Future<void> _handleConnect() async {
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);
    setState(() {
      _error = null;
      _loading = true;
    });
    try {
      await ref.read(sessionProvider.notifier).connect(
        SessionConfig(apiUrl: _apiUrl.trim(), apiKey: _apiKey.trim(), orgId: _orgId.trim()),
      );
    } catch (e) {
      final message = apiErrorMessage(asApiException(e), locale);
      if (mounted) setState(() => _error = message.isEmpty ? t('onboarding.invalidCredentials') : message);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);

    return Scaffold(
      appBar: AppBar(title: Text(t('onboarding.enterManually'))),
      body: AppScreen(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            AppFormField(
              label: t('onboarding.apiUrl'),
              value: _apiUrl,
              onChanged: (v) => setState(() => _apiUrl = v),
              placeholder: 'https://api.cwclock.me',
              keyboardType: TextInputType.url,
              textCapitalization: TextCapitalization.none,
            ),
            AppFormField(
              label: t('onboarding.apiKey'),
              value: _apiKey,
              onChanged: (v) => setState(() => _apiKey = v),
              obscureText: true,
              textCapitalization: TextCapitalization.none,
            ),
            AppFormField(
              label: t('onboarding.orgIdOptional'),
              value: _orgId,
              onChanged: (v) => setState(() => _orgId = v),
              textCapitalization: TextCapitalization.none,
            ),
            ErrorBanner(message: _error),
            AppButton(
              title: t('onboarding.connect'),
              onPressed: _apiUrl.trim().isEmpty || _apiKey.trim().isEmpty ? null : _handleConnect,
              loading: _loading,
            ),
          ],
        ),
      ),
    );
  }
}
