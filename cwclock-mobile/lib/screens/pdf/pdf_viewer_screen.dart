import 'package:flutter/material.dart';
import 'package:flutter_pdfview/flutter_pdfview.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:share_plus/share_plus.dart';

import '../../providers/locale_provider.dart';
import '../../theme.dart';
import '../../widgets/app_button.dart';
import '../../widgets/app_top_bar.dart';

/// The generic "show a locally-downloaded PDF" screen, reused by Reports,
/// Invoices' Preview/Generate and its previous-invoices list - all of them
/// just hand it a local file path and a title. Ported from
/// src/screens/pdf/PdfViewerScreen.js.
class PdfViewerScreen extends ConsumerStatefulWidget {
  final String path;
  final String? title;

  const PdfViewerScreen({super.key, required this.path, this.title});

  @override
  ConsumerState<PdfViewerScreen> createState() => _PdfViewerScreenState();
}

class _PdfViewerScreenState extends ConsumerState<PdfViewerScreen> {
  bool _failed = false;

  Future<void> _handleShare() async {
    // XFile wraps dart:io's File(path) directly - it wants a plain
    // filesystem path, not a file:// URI, so the share sheet used to open
    // an unreadable path (silently swallowed by the catch below) and never
    // show up. widget.path is already the raw path PdfClient wrote to.
    try {
      await SharePlus.instance.share(ShareParams(files: [XFile(widget.path)]));
    } catch (_) {
      // Matches the RN version's Share.open(...).catch(() => {}) - a
      // cancelled/failed share sheet isn't an error worth surfacing.
    }
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);

    return Scaffold(
      appBar: AppTopBar(title: widget.title ?? t('pdf.title')),
      body: Column(
        children: [
          Expanded(
            child: _failed
                ? Center(
                    child: Padding(
                      padding: EdgeInsets.all(AppSpacing.of(3)),
                      child: Text(
                        t('pdf.failedToLoad'),
                        textAlign: TextAlign.center,
                        style: TextStyle(color: AppColors.of(context).danger),
                      ),
                    ),
                  )
                : PDFView(
                    filePath: widget.path,
                    onError: (error) => setState(() => _failed = true),
                    onPageError: (page, error) => setState(() => _failed = true),
                  ),
          ),
          Padding(
            padding: EdgeInsets.all(AppSpacing.of(2)),
            child: AppButton(title: t('common.share'), onPressed: _handleShare),
          ),
        ],
      ),
    );
  }
}
