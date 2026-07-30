import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../common/format.dart';
import '../../providers/locale_provider.dart';
import '../../providers/reports_provider.dart';
import '../../providers/session_provider.dart';
import '../../widgets/app_button.dart';
import '../../widgets/app_screen.dart';
import '../../widgets/date_field.dart';
import '../../widgets/error_banner.dart';
import '../../widgets/select_field.dart';
import '../../widgets/app_top_bar.dart';

DateTime _firstOfMonth() {
  final now = DateTime.now();
  return DateTime(now.year, now.month, 1);
}

/// The "simplified generation screen for summary and detailed reports in PDF
/// format only" - just a report type and a date range, no client/project/user
/// filters like the web Reports page has. Ported from
/// src/screens/reports/ReportsScreen.js.
class ReportsScreen extends ConsumerStatefulWidget {
  const ReportsScreen({super.key});

  @override
  ConsumerState<ReportsScreen> createState() => _ReportsScreenState();
}

class _ReportsScreenState extends ConsumerState<ReportsScreen> {
  String _reportType = 'summary';
  DateTime _start = _firstOfMonth();
  DateTime _end = DateTime.now();
  String? _error;
  bool _generating = false;

  Future<void> _handleGenerate() async {
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);

    setState(() {
      _error = null;
      _generating = true;
    });
    try {
      final path = await ref
          .read(reportsServiceProvider)
          .generateReportPdf(orgId, _reportType, toDayString(_start), toDayString(_end));
      if (mounted) {
        context.push('/pdf-viewer', extra: {'path': path, 'title': t('reports.$_reportType')});
      }
    } catch (e) {
      setState(() => _error = apiErrorMessage(asApiException(e), locale));
    } finally {
      if (mounted) setState(() => _generating = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);

    return Scaffold(
      appBar: AppTopBar(title: t('reports.title')),
      body: AppScreen(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            AppSelectField(
              label: t('reports.type'),
              value: _reportType,
              onChanged: (v) => setState(() => _reportType = v),
              items: [
                SelectItem('summary', t('reports.summary')),
                SelectItem('detailed', t('reports.detailed')),
              ],
            ),
            AppDateField(label: t('reports.startDate'), value: _start, onChanged: (v) => setState(() => _start = v)),
            AppDateField(label: t('reports.endDate'), value: _end, onChanged: (v) => setState(() => _end = v)),
            ErrorBanner(message: _error),
            AppButton(title: t('reports.generate'), onPressed: _handleGenerate, loading: _generating),
          ],
        ),
      ),
    );
  }
}
