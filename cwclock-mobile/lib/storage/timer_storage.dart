import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';

/// A running timer's state (project/client/text/start time) - persisted so
/// backgrounding or killing the app doesn't lose an in-progress timer.
/// Mirrors the CLI's timer.json and src/storage/timer.js.
class RunningTimer {
  final String startedAt;
  final String projectId;
  final String clientId;
  final String text;

  const RunningTimer({
    required this.startedAt,
    required this.projectId,
    required this.clientId,
    required this.text,
  });

  factory RunningTimer.fromJson(Map<String, dynamic> json) {
    return RunningTimer(
      startedAt: json['startedAt'] as String,
      projectId: json['projectId'] as String,
      clientId: json['clientId'] as String,
      text: json['text'] as String,
    );
  }

  Map<String, dynamic> toJson() => {
        'startedAt': startedAt,
        'projectId': projectId,
        'clientId': clientId,
        'text': text,
      };
}

class TimerStorage {
  static const _timerKey = 'cwclock.timer';

  Future<void> saveTimer(RunningTimer timer) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_timerKey, jsonEncode(timer.toJson()));
  }

  Future<RunningTimer?> loadTimer() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_timerKey);
    if (raw == null) return null;
    return RunningTimer.fromJson(jsonDecode(raw) as Map<String, dynamic>);
  }

  Future<void> clearTimer() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_timerKey);
  }
}
