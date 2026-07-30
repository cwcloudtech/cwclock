import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/locale_provider.dart';
import '../providers/main_tab_provider.dart';
import '../providers/organizations_provider.dart';
import '../providers/permissions.dart' as perm;
import '../providers/session_provider.dart';
import '../theme.dart';
import 'invoices/invoices_screen.dart';
import 'reports/reports_screen.dart';
import 'settings/settings_screen.dart';
import 'time/time_tracker_screen.dart';

/// The connected app's bottom tab bar. Invoices is only shown to an
/// admin/owner of the current org (mirrors cwclock-ui's showInvoices gate) -
/// members/reader roles get Time tracker/Reports/Settings only. Ported from
/// src/App.js's MainTabs. Unlike the RN stack-of-tabs, screens pushed "on
/// top" (EditRecord, PdfViewer, etc.) are separate top-level go_router
/// routes (see router.dart) rather than nested within this shell, so they
/// cover the whole screen including this bottom bar - same visual result as
/// React Navigation's stack-over-tabs.
///
/// The active tab lives in [mainTabProvider] rather than local State, so
/// AppTopBar's avatar action (ai-instruct-121) can jump straight to Settings
/// from any pushed screen too.
class MainTabsScreen extends ConsumerStatefulWidget {
  const MainTabsScreen({super.key});

  @override
  ConsumerState<MainTabsScreen> createState() => _MainTabsScreenState();
}

class _MainTabsScreenState extends ConsumerState<MainTabsScreen> {
  @override
  void initState() {
    super.initState();
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId != null) {
      Future.microtask(() => ref.read(organizationsProvider.notifier).listMembers(orgId));
    }
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);
    final session = ref.watch(sessionProvider);
    final members = ref.watch(organizationsProvider).members;
    final showInvoices = perm.isAdminOrOwner(session.user, members);
    final activeTab = ref.watch(mainTabProvider);

    final tabs = <_TabDef>[
      _TabDef(MainTab.timeTracker, t('timeTracker.title'), Icons.access_time, const TimeTrackerScreen()),
      _TabDef(MainTab.reports, t('reports.title'), Icons.description_outlined, const ReportsScreen()),
      if (showInvoices)
        _TabDef(MainTab.invoices, t('invoices.title'), Icons.receipt_long_outlined, const InvoicesScreen()),
      _TabDef(MainTab.settings, t('settings.title'), Icons.settings_outlined, const SettingsScreen()),
    ];

    final resolvedIndex = tabs.indexWhere((tab) => tab.id == activeTab);
    final index = resolvedIndex == -1 ? 0 : resolvedIndex;

    return Scaffold(
      body: IndexedStack(index: index, children: [for (final tab in tabs) tab.screen]),
      bottomNavigationBar: BottomNavigationBar(
        type: BottomNavigationBarType.fixed,
        currentIndex: index,
        selectedItemColor: AppColors.of(context).primary,
        unselectedItemColor: AppColors.of(context).textMuted,
        onTap: (i) => ref.read(mainTabProvider.notifier).state = tabs[i].id,
        items: [
          for (final tab in tabs) BottomNavigationBarItem(icon: Icon(tab.icon), label: tab.label),
        ],
      ),
    );
  }
}

class _TabDef {
  final MainTab id;
  final String label;
  final IconData icon;
  final Widget screen;

  const _TabDef(this.id, this.label, this.icon, this.screen);
}
