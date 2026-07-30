import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Which bottom tab MainTabsScreen shows - a stable identity (rather than a
/// raw index) since the Invoices tab is conditionally present, shifting
/// what index "Settings" resolves to. Lets any screen (e.g. AppTopBar's
/// avatar action, ai-instruct-121) switch straight to Settings without
/// knowing the current tab layout.
enum MainTab { timeTracker, reports, invoices, settings }

final mainTabProvider = StateProvider<MainTab>((ref) => MainTab.timeTracker);
