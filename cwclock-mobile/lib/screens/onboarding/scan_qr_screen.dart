import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../api/config_parser.dart';
import '../../providers/locale_provider.dart';
import '../../providers/session_provider.dart';
import '../../widgets/error_banner.dart';

/// Reads the same "key = value" config text the CLI's `cwclock configure
/// import` and the web ApiKeys page's QR/file downloads use (see
/// api/config_parser.dart) - mobile_scanner fires with the raw decoded
/// string, no separate QR-format handling needed since a QR code is just a
/// container for that text.
///
/// Uses [MobileScanner] directly with no separate permission_handler
/// pre-check, mirroring cwc-chat-mobile's `_QrScannerPage`
/// (~/cwc-chat-mobile/lib/settings.dart) - mobile_scanner requests/manages
/// camera permission internally when its controller starts. Layering a
/// second, separate permission_handler request in front of it (the previous
/// version of this screen) left the camera preview visibly running but
/// mobile_scanner's own controller never properly initialized, so
/// `onDetect` never fired - that's what "stays on camera without
/// recognition" (ai-instruct-119) was.
class ScanQrScreen extends ConsumerStatefulWidget {
  const ScanQrScreen({super.key});

  @override
  ConsumerState<ScanQrScreen> createState() => _ScanQrScreenState();
}

class _ScanQrScreenState extends ConsumerState<ScanQrScreen> {
  bool _hasScanned = false;
  String? _error;
  bool _connecting = false;

  Future<void> _handleDetect(BarcodeCapture capture) async {
    if (_hasScanned || _connecting) return;

    String? raw;
    for (final barcode in capture.barcodes) {
      if (barcode.rawValue != null && barcode.rawValue!.trim().isNotEmpty) {
        raw = barcode.rawValue;
        break;
      }
    }
    if (raw == null) return;

    final locale = ref.read(localeProvider);
    final t = translateWith(locale);
    final parsed = parseConfigText(raw);
    if (!parsed.isComplete) {
      setState(() => _error = t('onboarding.scanFailed'));
      return;
    }

    _hasScanned = true;
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
      if (mounted) {
        setState(() {
          _error = message;
          _hasScanned = false;
        });
      }
    } finally {
      if (mounted) setState(() => _connecting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);

    return Scaffold(
      appBar: AppBar(title: Text(t('onboarding.scanQr'))),
      body: Stack(
        children: [
          MobileScanner(onDetect: _handleDetect),
          Positioned(
            left: 0,
            right: 0,
            bottom: 0,
            child: Container(
              padding: const EdgeInsets.all(16),
              color: Colors.black.withValues(alpha: 0.55),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    t('onboarding.scanHint'),
                    textAlign: TextAlign.center,
                    style: const TextStyle(color: Colors.white, fontSize: 14),
                  ),
                  ErrorBanner(message: _error),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
