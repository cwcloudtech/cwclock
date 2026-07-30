import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../api/config_parser.dart';
import '../../models/session_config.dart';
import '../../providers/locale_provider.dart';
import '../../providers/session_provider.dart';
import '../../widgets/app_button.dart';
import '../../widgets/app_screen.dart';
import '../../widgets/error_banner.dart';
import '../../widgets/form_field.dart';

/// Covers onboarding without a camera: either paste the whole downloaded
/// config file's text to fill the three fields at once, or type them in
/// directly - both end at the same SessionNotifier.connect call as
/// ScanQrScreen. Ported from src/screens/onboarding/ManualEntryScreen.js.
class ManualEntryScreen extends ConsumerStatefulWidget {
  const ManualEntryScreen({super.key});

  @override
  ConsumerState<ManualEntryScreen> createState() => _ManualEntryScreenState();
}

class _ManualEntryScreenState extends ConsumerState<ManualEntryScreen> {
  final _pasteFocusNode = FocusNode();
  String _pasteText = '';
  String _apiUrl = '';
  String _apiKey = '';
  String _orgId = '';
  String? _error;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _pasteFocusNode.addListener(() {
      if (!_pasteFocusNode.hasFocus) _handleParsePaste();
    });
  }

  @override
  void dispose() {
    _pasteFocusNode.dispose();
    super.dispose();
  }

  void _handleParsePaste() {
    final parsed = parseConfigText(_pasteText);
    setState(() {
      if (parsed.apiUrl.isNotEmpty) _apiUrl = parsed.apiUrl;
      if (parsed.apiKey.isNotEmpty) _apiKey = parsed.apiKey;
      if (parsed.orgId.isNotEmpty) _orgId = parsed.orgId;
    });
  }

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

    return AppScreen(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          AppFormField(
            label: t('onboarding.pasteConfig'),
            value: _pasteText,
            onChanged: (v) => setState(() => _pasteText = v),
            placeholder: 'api_url = ...\napi_key = ...\norg_id = ...',
            maxLines: 3,
            textCapitalization: TextCapitalization.none,
            focusNode: _pasteFocusNode,
          ),
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
    );
  }
}
