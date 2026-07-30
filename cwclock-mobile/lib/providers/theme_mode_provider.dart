import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'api_providers.dart';

/// Persisted light/dark preference (ai-instruct-119) - defaults to the
/// system setting until the user picks one explicitly in Settings, then
/// sticks with that choice like the locale preference does.
class ThemeModeNotifier extends Notifier<ThemeMode> {
  @override
  ThemeMode build() => ThemeMode.system;

  Future<void> load() async {
    final stored = await ref.read(localStorageProvider).getThemeMode();
    switch (stored) {
      case 'light':
        state = ThemeMode.light;
        break;
      case 'dark':
        state = ThemeMode.dark;
        break;
    }
  }

  Future<void> setThemeMode(ThemeMode mode) async {
    if (mode == ThemeMode.system) {
      await ref.read(localStorageProvider).setThemeMode('system');
    } else {
      await ref.read(localStorageProvider).setThemeMode(mode == ThemeMode.dark ? 'dark' : 'light');
    }
    state = mode;
  }

  Future<void> toggle() {
    final current = state == ThemeMode.system
        ? (WidgetsBinding.instance.platformDispatcher.platformBrightness == Brightness.dark
            ? ThemeMode.dark
            : ThemeMode.light)
        : state;
    return setThemeMode(current == ThemeMode.dark ? ThemeMode.light : ThemeMode.dark);
  }
}

final themeModeProvider = NotifierProvider<ThemeModeNotifier, ThemeMode>(ThemeModeNotifier.new);
