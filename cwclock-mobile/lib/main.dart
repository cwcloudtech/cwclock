import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'providers/locale_provider.dart';
import 'providers/session_provider.dart';
import 'providers/theme_mode_provider.dart';
import 'router.dart';
import 'theme.dart';

void main() {
  runApp(const ProviderScope(child: CwclockApp()));
}

/// Ported from src/App.js. On mount, restores a previously saved session
/// (see SessionNotifier.restoreSession) and loads the persisted locale/theme
/// mode - the router's redirect (see router.dart) reacts to the resulting
/// session status, same as the RN app's RootNavigator.
class CwclockApp extends ConsumerStatefulWidget {
  const CwclockApp({super.key});

  @override
  ConsumerState<CwclockApp> createState() => _CwclockAppState();
}

class _CwclockAppState extends ConsumerState<CwclockApp> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      ref.read(sessionProvider.notifier).restoreSession();
      ref.read(localeProvider.notifier).load();
      ref.read(themeModeProvider.notifier).load();
    });
  }

  @override
  Widget build(BuildContext context) {
    final router = ref.watch(routerProvider);
    final themeMode = ref.watch(themeModeProvider);
    return MaterialApp.router(
      title: 'CWClock',
      debugShowCheckedModeBanner: false,
      theme: buildAppTheme(Brightness.light),
      darkTheme: buildAppTheme(Brightness.dark),
      themeMode: themeMode,
      routerConfig: router,
    );
  }
}
