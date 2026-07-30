import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/locale_provider.dart';
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
class MainTabsScreen extends ConsumerStatefulWidget {
  const MainTabsScreen({super.key});

  @override
  ConsumerState<MainTabsScreen> createState() => _MainTabsScreenState();
}

class _MainTabsScreenState extends ConsumerState<MainTabsScreen> {
  int _index = 0;

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

    final tabs = <_TabDef>[
      _TabDef(t('timeTracker.title'), Icons.access_time, const TimeTrackerScreen()),
      _TabDef(t('reports.title'), Icons.description_outlined, const ReportsScreen()),
      if (showInvoices) _TabDef(t('invoices.title'), Icons.receipt_long_outlined, const InvoicesScreen()),
      _TabDef(t('settings.title'), Icons.settings_outlined, const SettingsScreen()),
    ];

    final index = _index >= tabs.length ? 0 : _index;

    return Scaffold(
      body: IndexedStack(index: index, children: [for (final tab in tabs) tab.screen]),
      bottomNavigationBar: BottomNavigationBar(
        type: BottomNavigationBarType.fixed,
        currentIndex: index,
        selectedItemColor: AppColors.of(context).primary,
        unselectedItemColor: AppColors.of(context).textMuted,
        onTap: (i) => setState(() => _index = i),
        items: [
          for (final tab in tabs) BottomNavigationBarItem(icon: Icon(tab.icon), label: tab.label),
        ],
      ),
    );
  }
}

class _TabDef {
  final String label;
  final IconData icon;
  final Widget screen;

  const _TabDef(this.label, this.icon, this.screen);
}
