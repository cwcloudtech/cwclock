import '../models/time_entry.dart';

/// Zero-pads each ":"-separated segment of a time string (e.g. "9:5:3" ->
/// "09:05:03"), matching what the API's "HH:MM:SS" fields expect.
String padTimeString(String? value) {
  if (value == null || value.isEmpty) return value ?? '';
  return value.split(':').map((part) => part.padLeft(2, '0')).join(':');
}

/// Converts a DateTime into the "YYYY-MM-DD" string the API expects for the
/// day field - local calendar date, no timezone conversion.
String toDayString(DateTime date) {
  final y = date.year.toString().padLeft(4, '0');
  final m = date.month.toString().padLeft(2, '0');
  final d = date.day.toString().padLeft(2, '0');
  return '$y-$m-$d';
}

/// Converts a DateTime into the "HH:MM:SS" string the API expects for
/// start/end fields - local wall-clock time, no timezone conversion.
String toHMS(DateTime date) {
  return [date.hour, date.minute, date.second]
      .map((v) => v.toString().padLeft(2, '0'))
      .join(':');
}

String formatHMS(int secs) {
  final h = secs ~/ 3600;
  final m = (secs % 3600) ~/ 60;
  final s = secs % 60;
  return [h, m, s].map((v) => v.toString().padLeft(2, '0')).join(':');
}

int _parseSecondsOfDay(String? hms) {
  if (hms == null || hms.isEmpty) return 0;
  final parts = hms.split(':').map((p) => int.tryParse(p) ?? 0).toList();
  final h = parts.isNotEmpty ? parts[0] : 0;
  final m = parts.length > 1 ? parts[1] : 0;
  final s = parts.length > 2 ? parts[2] : 0;
  return h * 3600 + m * 60 + s;
}

/// An all-day or half-day entry contributes nothing to this personal duration
/// total - that billing math needs the client's hoursPerDay, computed
/// server-side for Reports/Invoices instead.
int entryDurationSecs(TimeEntry item) {
  if (item.allDay || item.half || item.start == null || item.end == null) {
    return 0;
  }
  final secs = _parseSecondsOfDay(item.end) - _parseSecondsOfDay(item.start);
  return secs > 0 ? secs : 0;
}

class DayGroup {
  final String day;
  final List<TimeEntry> items;
  final int totalSecs;

  const DayGroup({required this.day, required this.items, required this.totalSecs});
}

/// Entries already arrive sorted day DESC from the API (and stay sorted
/// client-side), so a single pass grouping consecutive same-day items is
/// enough.
List<DayGroup> groupByDay(List<TimeEntry> items) {
  final groups = <_MutableDayGroup>[];
  for (final item in items) {
    final last = groups.isNotEmpty ? groups.last : null;
    if (last != null && last.day == item.day) {
      last.items.add(item);
      last.totalSecs += entryDurationSecs(item);
    } else {
      groups.add(_MutableDayGroup(
        day: item.day,
        items: [item],
        totalSecs: entryDurationSecs(item),
      ));
    }
  }
  return groups
      .map((g) => DayGroup(day: g.day, items: g.items, totalSecs: g.totalSecs))
      .toList();
}

class _MutableDayGroup {
  final String day;
  final List<TimeEntry> items;
  int totalSecs;

  _MutableDayGroup({required this.day, required this.items, required this.totalSecs});
}
