import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../common/format.dart';
import '../../common/project_label.dart';
import '../../providers/clients_provider.dart';
import '../../providers/locale_provider.dart';
import '../../providers/projects_provider.dart';
import '../../providers/session_provider.dart';
import '../../providers/time_entries_provider.dart';
import '../../theme.dart';
import '../../widgets/app_button.dart';
import '../../widgets/app_screen.dart';
import '../../widgets/date_field.dart';
import '../../widgets/error_banner.dart';
import '../../widgets/form_field.dart';
import '../../widgets/select_field.dart';
import '../../widgets/app_top_bar.dart';

/// Picking a project and a date creates the entry immediately. Ported from
/// src/screens/time/AllDayRecordScreen.js.
class AllDayRecordScreen extends ConsumerStatefulWidget {
  const AllDayRecordScreen({super.key});

  @override
  ConsumerState<AllDayRecordScreen> createState() => _AllDayRecordScreenState();
}

class _AllDayRecordScreenState extends ConsumerState<AllDayRecordScreen> {
  String _projectId = '';
  String _text = '';
  DateTime _day = DateTime.now();
  bool _half = false;
  String? _error;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId != null) {
      Future.microtask(() => ref.read(projectsProvider.notifier).listProjects(orgId));
    }
  }

  Future<void> _handleCreate() async {
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;
    final locale = ref.read(localeProvider);
    final projects = ref.read(projectsProvider).items;
    final project = projects.where((p) => p.id == _projectId).firstOrNull;
    if (project == null) return;

    setState(() {
      _error = null;
      _saving = true;
    });
    try {
      await ref.read(timeEntriesProvider.notifier).createTimeEntry(orgId, {
        'text': _text.trim().isEmpty ? project.name : _text.trim(),
        'day': toDayString(_day),
        'allDay': !_half,
        'half': _half,
        'clientId': project.clientId,
        'projectId': project.id,
      });
      if (mounted) Navigator.of(context).pop();
    } catch (e) {
      setState(() => _error = apiErrorMessage(asApiException(e), locale));
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);
    final projects = ref.watch(projectsProvider).items;
    final clients = ref.watch(clientsProvider).items;

    final projectItems = [
      for (final p in projects) SelectItem(p.id, projectLabel(p, clients)),
    ];

    return Scaffold(
      appBar: AppTopBar(title: t('timeTracker.addAllDayRecord')),
      body: AppScreen(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            AppSelectField(
              label: t('timeTracker.project'),
              value: _projectId,
              onChanged: (v) => setState(() => _projectId = v),
              items: projectItems,
              placeholder: t('timeTracker.project'),
            ),
            AppFormField(
              label: t('timeTracker.taskDescription'),
              value: _text,
              onChanged: (v) => setState(() => _text = v),
              placeholder: t('timeTracker.whatAreYouWorkingOn'),
            ),
            Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                Expanded(
                  child: AppDateField(
                    label: t('timeTracker.day'),
                    value: _day,
                    onChanged: (v) => setState(() => _day = v),
                  ),
                ),
                SizedBox(width: AppSpacing.of(1.5)),
                Padding(
                  padding: EdgeInsets.only(bottom: AppSpacing.of(2)),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        t('timeTracker.half'),
                        style: TextStyle(fontSize: 13, color: AppColors.of(context).textMuted),
                      ),
                      Switch(
                        value: _half,
                        onChanged: (v) => setState(() => _half = v),
                        activeTrackColor: AppColors.of(context).primary,
                      ),
                    ],
                  ),
                ),
              ],
            ),
            ErrorBanner(message: _error),
            AppButton(
              title: t('timeTracker.addAllDayRecord'),
              onPressed: _projectId.isEmpty ? null : _handleCreate,
              loading: _saving,
            ),
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
