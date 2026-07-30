import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../common/format.dart';
import '../../common/project_label.dart';
import '../../models/time_entry.dart';
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
import '../../widgets/toggle_row.dart';

DateTime _timeStringToDate(String? hms) {
  final padded = padTimeString(hms).isEmpty ? '09:00:00' : padTimeString(hms);
  final parts = padded.split(':').map((p) => int.tryParse(p) ?? 0).toList();
  final now = DateTime.now();
  return DateTime(
    now.year,
    now.month,
    now.day,
    parts.isNotEmpty ? parts[0] : 9,
    parts.length > 1 ? parts[1] : 0,
    parts.length > 2 ? parts[2] : 0,
  );
}

/// Mirrors cwclock-ui's edit form: text/day/project plus All day / Half day
/// toggles that disable each other (mutually exclusive server-side), with
/// start/end time fields only shown when neither is set. Ported from
/// src/screens/time/EditRecordScreen.js.
class EditRecordScreen extends ConsumerStatefulWidget {
  final TimeEntry entry;

  const EditRecordScreen({super.key, required this.entry});

  @override
  ConsumerState<EditRecordScreen> createState() => _EditRecordScreenState();
}

class _EditRecordScreenState extends ConsumerState<EditRecordScreen> {
  late String _text = widget.entry.text;
  late DateTime _day = DateTime.parse('${widget.entry.day}T00:00:00');
  late String _projectId = widget.entry.projectId;
  late bool _allDay = widget.entry.allDay;
  late bool _half = widget.entry.half;
  late DateTime _start = _timeStringToDate(widget.entry.start);
  late DateTime _end = _timeStringToDate(widget.entry.end);
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

  Future<void> _handleSave() async {
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;
    final locale = ref.read(localeProvider);
    final projects = ref.read(projectsProvider).items;
    final project = projects.where((p) => p.id == _projectId).firstOrNull;

    setState(() {
      _error = null;
      _saving = true;
    });
    try {
      final updated = widget.entry.copyWith(
        text: _text,
        day: toDayString(_day),
        allDay: _allDay,
        half: _half,
        start: _allDay || _half ? null : toHMS(_start),
        end: _allDay || _half ? null : toHMS(_end),
        projectId: _projectId,
        clientId: project?.clientId ?? widget.entry.clientId,
      );
      await ref.read(timeEntriesProvider.notifier).updateTimeEntry(orgId, updated);
      if (mounted) Navigator.of(context).pop();
    } catch (e) {
      setState(() => _error = apiErrorMessage(asApiException(e), locale));
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  void _handleDelete() {
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;

    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: Text(t('timeTracker.deleteRecordTitle')),
        content: Text(t('timeTracker.deleteRecordBody', {'text': widget.entry.text})),
        actions: [
          TextButton(onPressed: () => Navigator.pop(dialogContext), child: Text(t('common.cancel'))),
          TextButton(
            onPressed: () async {
              Navigator.pop(dialogContext);
              await ref.read(timeEntriesProvider.notifier).deleteTimeEntry(orgId, widget.entry.id);
              if (!mounted) return;
              Navigator.of(context).pop();
            },
            child: Text(t('common.delete'), style: const TextStyle(color: AppColors.danger)),
          ),
        ],
      ),
    );
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
      appBar: AppBar(title: Text(t('timeTracker.editRecord'))),
      body: AppScreen(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            AppFormField(
              label: t('timeTracker.taskDescription'),
              value: _text,
              onChanged: (v) => setState(() => _text = v),
            ),
            AppSelectField(
              label: t('timeTracker.project'),
              value: _projectId,
              onChanged: (v) => setState(() => _projectId = v),
              items: projectItems,
            ),
            AppDateField(
              label: t('timeTracker.day'),
              value: _day,
              onChanged: (v) => setState(() => _day = v),
            ),
            AppToggleRow(
              label: t('timeTracker.allDay'),
              value: _allDay,
              disabled: _half,
              onChanged: (v) => setState(() => _allDay = v),
            ),
            AppToggleRow(
              label: t('timeTracker.half'),
              value: _half,
              disabled: _allDay,
              onChanged: (v) => setState(() => _half = v),
            ),
            if (!_allDay && !_half) ...[
              AppDateField(
                label: t('timeTracker.startTime'),
                value: _start,
                mode: DateFieldMode.time,
                onChanged: (v) => setState(() => _start = v),
              ),
              AppDateField(
                label: t('timeTracker.endTime'),
                value: _end,
                mode: DateFieldMode.time,
                onChanged: (v) => setState(() => _end = v),
              ),
            ],
            ErrorBanner(message: _error),
            AppButton(
              title: t('common.save'),
              onPressed: _handleSave,
              loading: _saving,
              margin: EdgeInsets.only(bottom: AppSpacing.of(1.5)),
            ),
            AppButton(title: t('common.delete'), variant: AppButtonVariant.danger, onPressed: _handleDelete),
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
