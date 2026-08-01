import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../common/format.dart';
import '../../common/member_label.dart';
import '../../common/project_label.dart';
import '../../providers/clients_provider.dart';
import '../../providers/locale_provider.dart';
import '../../providers/organizations_provider.dart';
import '../../providers/projects_provider.dart';
import '../../providers/reports_provider.dart';
import '../../providers/session_provider.dart';
import '../../theme.dart';
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

/// A report type and date range, plus client/project/member filters
/// (ai-instruct-127) mirroring the web app's Reports page - single-select
/// here rather than the web's MultiSelect, matching the rest of the mobile
/// app's simplified pickers (eg. Invoices' client select). Ported from
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
  String _clientId = '';
  String _projectId = '';
  String _memberId = '';
  String? _error;
  bool _generating = false;

  @override
  void initState() {
    super.initState();
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId != null) {
      Future.microtask(() {
        ref.read(clientsProvider.notifier).listClients(orgId);
        ref.read(projectsProvider.notifier).listProjects(orgId);
        ref.read(organizationsProvider.notifier).listMembers(orgId);
      });
    }
  }

  // Narrowing the client filter narrows which projects make sense too, so
  // drop an already-selected project that no longer belongs to it - mirrors
  // cwclock-ui's Reports.jsx effect.
  void _handleClientChanged(String clientId) {
    final projects = ref.read(projectsProvider).items;
    final selected = projects.where((p) => p.id == _projectId).firstOrNull;
    setState(() {
      _clientId = clientId;
      if (clientId.isNotEmpty && selected != null && selected.clientId != clientId) {
        _projectId = '';
      }
    });
  }

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
      final path = await ref.read(reportsServiceProvider).generateReportPdf(
        orgId,
        _reportType,
        toDayString(_start),
        toDayString(_end),
        clientId: _clientId,
        projectId: _projectId,
        memberId: _memberId,
      );
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
    final clients = ref.watch(clientsProvider).items;
    final projects = ref.watch(projectsProvider).items;
    final members = ref.watch(organizationsProvider).members;

    final clientItems = [
      SelectItem('', t('reports.allClients')),
      for (final c in clients) SelectItem(c.id, c.name),
    ];
    final projectItems = [
      SelectItem('', t('reports.allProjects')),
      for (final p in projects.where((p) => _clientId.isEmpty || p.clientId == _clientId))
        SelectItem(p.id, projectLabel(p, clients)),
    ];
    final memberItems = [
      SelectItem('', t('reports.allMembers')),
      for (final m in members) SelectItem(m.userId, memberLabel(m)),
    ];

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
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Expanded(
                  child: AppDateField(
                    label: t('reports.startDate'),
                    value: _start,
                    onChanged: (v) => setState(() => _start = v),
                  ),
                ),
                SizedBox(width: AppSpacing.of(1.5)),
                Expanded(
                  child: AppDateField(
                    label: t('reports.endDate'),
                    value: _end,
                    onChanged: (v) => setState(() => _end = v),
                  ),
                ),
              ],
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Expanded(
                  child: AppSelectField(
                    label: t('reports.client'),
                    value: _clientId,
                    onChanged: _handleClientChanged,
                    items: clientItems,
                  ),
                ),
                SizedBox(width: AppSpacing.of(1.5)),
                Expanded(
                  child: AppSelectField(
                    label: t('reports.project'),
                    value: _projectId,
                    onChanged: (v) => setState(() => _projectId = v),
                    items: projectItems,
                  ),
                ),
              ],
            ),
            AppSelectField(
              label: t('reports.member'),
              value: _memberId,
              onChanged: (v) => setState(() => _memberId = v),
              items: memberItems,
            ),
            ErrorBanner(message: _error),
            AppButton(title: t('reports.generate'), onPressed: _handleGenerate, loading: _generating),
          ],
        ),
      ),
    );
  }
}

extension _FirstOrNull<T> on Iterable<T> {
  T? get firstOrNull {
    final it = iterator;
    return it.moveNext() ? it.current : null;
  }
}
