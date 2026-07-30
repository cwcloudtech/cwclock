import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/time_entry.dart';
import '../storage/timer_storage.dart';
import 'api_providers.dart';

const _pageSize = 50;

class TimeEntriesState {
  final List<TimeEntry> items;
  final bool isLoading;
  final bool isLoadingMore;
  final bool isError;
  final int page;
  final bool hasMore;

  const TimeEntriesState({
    this.items = const [],
    this.isLoading = false,
    this.isLoadingMore = false,
    this.isError = false,
    this.page = 1,
    this.hasMore = false,
  });

  TimeEntriesState copyWith({
    List<TimeEntry>? items,
    bool? isLoading,
    bool? isLoadingMore,
    bool? isError,
    int? page,
    bool? hasMore,
  }) {
    return TimeEntriesState(
      items: items ?? this.items,
      isLoading: isLoading ?? this.isLoading,
      isLoadingMore: isLoadingMore ?? this.isLoadingMore,
      isError: isError ?? this.isError,
      page: page ?? this.page,
      hasMore: hasMore ?? this.hasMore,
    );
  }
}

/// Matches the API's own ORDER BY day DESC, start DESC, so an entry
/// inserted/edited client-side (rather than re-fetched) lands in the same
/// slot the server would put it in - ported from
/// src/redux/timeEntries/timeEntries.reducer.js's compareEntries.
int _compareEntries(TimeEntry a, TimeEntry b) {
  if (a.day != b.day) return a.day.compareTo(b.day) > 0 ? -1 : 1;
  final aStart = a.start ?? '';
  final bStart = b.start ?? '';
  if (aStart == bStart) return 0;
  return aStart.compareTo(bStart) > 0 ? -1 : 1;
}

/// Ported from src/redux/timeEntries/timeEntries.actions.js +
/// timeEntries.reducer.js - page 1 replaces the list (initial load /
/// pull-to-refresh), page > 1 appends (infinite scroll).
class TimeEntriesNotifier extends Notifier<TimeEntriesState> {
  @override
  TimeEntriesState build() => const TimeEntriesState();

  Future<void> listTimeEntries(String orgId, [int page = 1]) async {
    state = page == 1 ? state.copyWith(isLoading: true, isError: false) : state.copyWith(isLoadingMore: true);
    try {
      final response = await ref.read(apiClientProvider).dio.get(
        '/organizations/$orgId/time-entries/',
        queryParameters: {'page': page, 'pageSize': _pageSize},
      );
      final data = response.data as Map<String, dynamic>;
      final items = (data['items'] as List? ?? [])
          .map((e) => TimeEntry.fromJson(e as Map<String, dynamic>))
          .toList();
      final hasMore = data['hasMore'] as bool? ?? false;
      final resolvedPage = data['page'] as int? ?? page;
      state = page == 1
          ? state.copyWith(items: items, isLoading: false, page: resolvedPage, hasMore: hasMore)
          : state.copyWith(
              items: [...state.items, ...items],
              isLoadingMore: false,
              page: resolvedPage,
              hasMore: hasMore,
            );
    } catch (e) {
      state = state.copyWith(isLoading: false, isLoadingMore: false, isError: true);
      rethrow;
    }
  }

  Future<TimeEntry> createTimeEntry(String orgId, Map<String, dynamic> fields) async {
    final response =
        await ref.read(apiClientProvider).dio.post('/organizations/$orgId/time-entries/', data: fields);
    final entry = TimeEntry.fromJson(response.data as Map<String, dynamic>);
    final items = [...state.items, entry]..sort(_compareEntries);
    state = state.copyWith(items: items);
    return entry;
  }

  Future<TimeEntry> updateTimeEntry(String orgId, TimeEntry entry) async {
    final response = await ref
        .read(apiClientProvider)
        .dio
        .put('/organizations/$orgId/time-entries/${entry.id}', data: entry.toJson());
    final updated = TimeEntry.fromJson(response.data as Map<String, dynamic>);
    final items = [for (final e in state.items) e.id == updated.id ? updated : e]..sort(_compareEntries);
    state = state.copyWith(items: items);
    return updated;
  }

  Future<void> deleteTimeEntry(String orgId, String id) async {
    await ref.read(apiClientProvider).dio.delete('/organizations/$orgId/time-entries/$id');
    state = state.copyWith(items: state.items.where((e) => e.id != id).toList());
  }
}

final timeEntriesProvider = NotifierProvider<TimeEntriesNotifier, TimeEntriesState>(TimeEntriesNotifier.new);

/// The running timer (start/stop header on the Time tracker screen) -
/// persisted via [TimerStorage] so backgrounding/killing the app doesn't
/// lose it. Ported from TimeTrackerScreen.js's local `timer` state.
class RunningTimerNotifier extends Notifier<RunningTimer?> {
  @override
  RunningTimer? build() => null;

  Future<void> load() async {
    state = await ref.read(timerStorageProvider).loadTimer();
  }

  Future<void> start({required String projectId, required String clientId, required String text}) async {
    final timer = RunningTimer(
      startedAt: DateTime.now().toIso8601String(),
      projectId: projectId,
      clientId: clientId,
      text: text,
    );
    await ref.read(timerStorageProvider).saveTimer(timer);
    state = timer;
  }

  Future<void> clear() async {
    await ref.read(timerStorageProvider).clearTimer();
    state = null;
  }
}

final runningTimerProvider = NotifierProvider<RunningTimerNotifier, RunningTimer?>(RunningTimerNotifier.new);
