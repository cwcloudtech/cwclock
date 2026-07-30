import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

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
import '../../widgets/form_field.dart';
import '../../widgets/select_field.dart';

/// The default tab: a running-timer header (start/stop) above a
/// day-grouped list of entries. All-day records are created from
/// AllDayRecordScreen (the "+" button); tapping a row opens EditRecordScreen.
/// Ported from src/screens/time/TimeTrackerScreen.js.
class TimeTrackerScreen extends ConsumerStatefulWidget {
  const TimeTrackerScreen({super.key});

  @override
  ConsumerState<TimeTrackerScreen> createState() => _TimeTrackerScreenState();
}

class _TimeTrackerScreenState extends ConsumerState<TimeTrackerScreen> {
  final _scrollController = ScrollController();
  String _projectId = '';
  String _text = '';
  int _elapsed = 0;
  bool _starting = false;
  Timer? _ticker;

  @override
  void initState() {
    super.initState();
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId != null) {
      Future.microtask(() {
        ref.read(projectsProvider.notifier).listProjects(orgId);
        ref.read(clientsProvider.notifier).listClients(orgId);
        ref.read(timeEntriesProvider.notifier).listTimeEntries(orgId, 1);
        ref.read(runningTimerProvider.notifier).load();
      });
    }
    _scrollController.addListener(_onScroll);
  }

  void _onScroll() {
    if (!_scrollController.hasClients) return;
    final threshold = _scrollController.position.maxScrollExtent * 0.6;
    if (_scrollController.position.pixels >= threshold) {
      final entriesState = ref.read(timeEntriesProvider);
      final orgId = ref.read(sessionProvider).orgId;
      if (orgId != null && entriesState.hasMore && !entriesState.isLoadingMore) {
        ref.read(timeEntriesProvider.notifier).listTimeEntries(orgId, entriesState.page + 1);
      }
    }
  }

  void _startTicking(DateTime startedAt) {
    _ticker?.cancel();
    void tick() {
      if (!mounted) return;
      final diff = DateTime.now().difference(startedAt).inSeconds;
      setState(() {
        _elapsed = diff < 0 ? 0 : diff;
      });
    }

    tick();
    _ticker = Timer.periodic(const Duration(seconds: 1), (_) => tick());
  }

  @override
  void dispose() {
    _ticker?.cancel();
    _scrollController.dispose();
    super.dispose();
  }

  Future<void> _handleStart() async {
    if (_projectId.isEmpty) return;
    final projects = ref.read(projectsProvider).items;
    final project = projects.where((p) => p.id == _projectId).firstOrNull;
    if (project == null) return;

    setState(() => _starting = true);
    await ref.read(runningTimerProvider.notifier).start(
      projectId: project.id,
      clientId: project.clientId,
      text: _text.trim().isEmpty ? project.name : _text.trim(),
    );
    setState(() => _starting = false);
  }

  Future<void> _handleStop() async {
    final timer = ref.read(runningTimerProvider);
    if (timer == null) return;
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;
    final started = DateTime.parse(timer.startedAt);
    final now = DateTime.now();
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);

    try {
      await ref.read(timeEntriesProvider.notifier).createTimeEntry(orgId, {
        'text': timer.text,
        'day': toDayString(started),
        'start': toHMS(started),
        'end': toHMS(now),
        'allDay': false,
        'half': false,
        'clientId': timer.clientId,
        'projectId': timer.projectId,
      });
      await ref.read(runningTimerProvider.notifier).clear();
      _ticker?.cancel();
      setState(() {
        _text = '';
        _projectId = '';
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(t('common.retry'))));
      }
    }
  }

  void _handleDelete(TimeEntry item) {
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(t('timeTracker.deleteRecordTitle')),
        content: Text(t('timeTracker.deleteRecordBody', {'text': item.text})),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: Text(t('common.cancel'))),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              ref.read(timeEntriesProvider.notifier).deleteTimeEntry(orgId, item.id);
            },
            child: Text(t('common.delete'), style: TextStyle(color: AppColors.of(context).danger)),
          ),
        ],
      ),
    );
  }

  String _timeLabel(TimeEntry item, String Function(String, [Map<String, String>?]) t) {
    if (item.allDay) return t('timeTracker.allDay');
    if (item.half) return t('timeTracker.half');
    return '${padTimeString(item.start).isEmpty ? '?' : padTimeString(item.start)} - '
        '${padTimeString(item.end).isEmpty ? '?' : padTimeString(item.end)}';
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);
    final session = ref.watch(sessionProvider);
    final projects = ref.watch(projectsProvider).items;
    final clients = ref.watch(clientsProvider).items;
    final entriesState = ref.watch(timeEntriesProvider);
    final timer = ref.watch(runningTimerProvider);

    ref.listen(runningTimerProvider, (previous, next) {
      if (next != null) _startTicking(DateTime.parse(next.startedAt));
    });

    final projectItems = [
      for (final p in projects) SelectItem(p.id, projectLabel(p, clients)),
    ];

    final groups = groupByDay(entriesState.items);

    return Scaffold(
      appBar: AppBar(title: Text(t('timeTracker.title'))),
      body: Column(
        children: [
          Container(
            padding: EdgeInsets.all(AppSpacing.of(2)),
            decoration: BoxDecoration(
              border: Border(bottom: BorderSide(color: AppColors.of(context).border)),
            ),
            child: timer != null
                ? Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        timer.text,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: AppColors.of(context).text),
                      ),
                      Padding(
                        padding: EdgeInsets.symmetric(vertical: AppSpacing.of(1)),
                        child: Text(
                          formatHMS(_elapsed),
                          style: TextStyle(fontSize: 28, fontWeight: FontWeight.w700, color: AppColors.of(context).primary),
                        ),
                      ),
                      AppButton(title: t('timeTracker.stop'), variant: AppButtonVariant.danger, onPressed: _handleStop),
                    ],
                  )
                : Column(
                    children: [
                      AppSelectField(
                        label: t('timeTracker.project'),
                        value: _projectId,
                        onChanged: (v) => setState(() => _projectId = v),
                        items: projectItems,
                        placeholder: t('timeTracker.project'),
                      ),
                      AppFormField(
                        value: _text,
                        onChanged: (v) => setState(() => _text = v),
                        placeholder: t('timeTracker.whatAreYouWorkingOn'),
                      ),
                      AppButton(
                        title: t('timeTracker.start'),
                        onPressed: _projectId.isEmpty ? null : _handleStart,
                        loading: _starting,
                      ),
                    ],
                  ),
          ),
          Expanded(
            child: Stack(
              children: [
                RefreshIndicator(
                  onRefresh: () async {
                    if (session.orgId != null) {
                      await ref.read(timeEntriesProvider.notifier).listTimeEntries(session.orgId!, 1);
                    }
                  },
                  child: groups.isEmpty && !entriesState.isLoading
                      ? ListView(
                          children: [
                            Padding(
                              padding: EdgeInsets.all(AppSpacing.of(3)),
                              child: Text(
                                t('timeTracker.noRecords'),
                                textAlign: TextAlign.center,
                                style: TextStyle(color: AppColors.of(context).textMuted),
                              ),
                            ),
                          ],
                        )
                      : ListView.builder(
                          controller: _scrollController,
                          padding: EdgeInsets.only(bottom: AppSpacing.of(10)),
                          itemCount: groups.fold<int>(0, (sum, g) => sum + 1 + g.items.length),
                          itemBuilder: (context, index) {
                            var remaining = index;
                            for (final group in groups) {
                              if (remaining == 0) {
                                return Container(
                                  color: AppColors.of(context).backgroundMuted,
                                  padding: EdgeInsets.symmetric(
                                    horizontal: AppSpacing.of(2),
                                    vertical: AppSpacing.of(1),
                                  ),
                                  child: Row(
                                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                    children: [
                                      Text(
                                        group.day,
                                        style: TextStyle(
                                            fontSize: 13, fontWeight: FontWeight.w600, color: AppColors.of(context).textMuted),
                                      ),
                                      Text(
                                        formatHMS(group.totalSecs),
                                        style: TextStyle(
                                            fontSize: 13, fontWeight: FontWeight.w600, color: AppColors.of(context).textMuted),
                                      ),
                                    ],
                                  ),
                                );
                              }
                              remaining -= 1;
                              if (remaining < group.items.length) {
                                final item = group.items[remaining];
                                final project = projects.where((p) => p.id == item.projectId).firstOrNull;
                                return ListTile(
                                  contentPadding: EdgeInsets.symmetric(horizontal: AppSpacing.of(2)),
                                  title: Text(item.text, maxLines: 1, overflow: TextOverflow.ellipsis),
                                  subtitle: Text(
                                    '${project != null ? projectLabel(project, clients) : ''} · ${_timeLabel(item, t)}',
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                  trailing: TextButton(
                                    onPressed: () => _handleDelete(item),
                                    child: Text(t('common.delete'), style: TextStyle(color: AppColors.of(context).danger)),
                                  ),
                                  onTap: () => context.push('/edit-record', extra: item),
                                );
                              }
                              remaining -= group.items.length;
                            }
                            return const SizedBox.shrink();
                          },
                        ),
                ),
              ],
            ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => context.push('/all-day-record'),
        backgroundColor: AppColors.of(context).primary,
        child: const Icon(Icons.add, color: kWhite),
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
