import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:permission_handler/permission_handler.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../api/config_parser.dart';
import '../../providers/locale_provider.dart';
import '../../providers/session_provider.dart';
import '../../theme.dart';
import '../../widgets/app_screen.dart';
import '../../widgets/error_banner.dart';

/// Reads the same "key = value" config text the CLI's `cwclock configure
/// import` and the web ApiKeys page's QR/file downloads use (see
/// api/config_parser.dart) - mobile_scanner fires with the raw decoded
/// string, no separate QR-format handling needed since a QR code is just a
/// container for that text. Ported from
/// src/screens/onboarding/ScanQrScreen.js.
class ScanQrScreen extends ConsumerStatefulWidget {
  const ScanQrScreen({super.key});

  @override
  ConsumerState<ScanQrScreen> createState() => _ScanQrScreenState();
}

class _ScanQrScreenState extends ConsumerState<ScanQrScreen> {
  bool _hasPermission = false;
  String? _error;
  bool _connecting = false;

  @override
  void initState() {
    super.initState();
    Permission.camera.request().then((status) {
      if (mounted) setState(() => _hasPermission = status.isGranted);
    });
  }

  Future<void> _handleDetect(BarcodeCapture capture) async {
    if (_connecting) return;
    final raw = capture.barcodes.isNotEmpty ? capture.barcodes.first.rawValue : null;
    if (raw == null) return;

    final locale = ref.read(localeProvider);
    final t = translateWith(locale);
    final parsed = parseConfigText(raw);
    if (!parsed.isComplete) {
      setState(() => _error = t('onboarding.scanFailed'));
      return;
    }

    setState(() {
      _connecting = true;
      _error = null;
    });
    try {
      // Session status flips to "needsOrg" or "connected" - the router's
      // redirect (see router.dart) reacts to that and moves on by itself,
      // nothing to navigate to here.
      await ref.read(sessionProvider.notifier).connect(parsed);
    } catch (e) {
      final message = apiErrorMessage(asApiException(e), locale);
      if (mounted) setState(() => _error = message);
    } finally {
      if (mounted) setState(() => _connecting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);

    if (!_hasPermission) {
      return AppScreen(child: Text(t('onboarding.scanHint')));
    }

    return Stack(
      children: [
        MobileScanner(onDetect: _handleDetect),
        Positioned(
          left: 0,
          right: 0,
          bottom: 0,
          child: Container(
            padding: EdgeInsets.all(AppSpacing.of(2)),
            color: Colors.black.withValues(alpha: 0.55),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  t('onboarding.scanHint'),
                  textAlign: TextAlign.center,
                  style: const TextStyle(color: AppColors.white, fontSize: 14),
                ),
                ErrorBanner(message: _error),
              ],
            ),
          ),
        ),
      ],
    );
  }
}
