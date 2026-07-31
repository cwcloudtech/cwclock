import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'models/client.dart';
import 'models/project.dart';
import 'models/time_entry.dart';
import 'providers/session_provider.dart';
import 'screens/clients/client_form_screen.dart';
import 'screens/clients/clients_screen.dart';
import 'screens/export_jobs/export_jobs_screen.dart';
import 'screens/loading_screen.dart';
import 'screens/main_tabs_screen.dart';
import 'screens/onboarding/connect_screen.dart';
import 'screens/onboarding/manual_entry_screen.dart';
import 'screens/onboarding/scan_qr_screen.dart';
import 'screens/org_picker/org_picker_screen.dart';
import 'screens/organizations/invite_member_screen.dart';
import 'screens/organizations/members_screen.dart';
import 'screens/organizations/organization_screen.dart';
import 'screens/pdf/pdf_viewer_screen.dart';
import 'screens/projects/project_form_screen.dart';
import 'screens/projects/projects_screen.dart';
import 'screens/time/all_day_record_screen.dart';
import 'screens/time/edit_record_screen.dart';

/// Bridges Riverpod's sessionProvider changes into a [Listenable] go_router
/// can watch via `refreshListenable`, so route redirects re-evaluate
/// whenever session status changes (restore/connect/selectOrg/disconnect).
class _RouterRefreshNotifier extends ChangeNotifier {
  _RouterRefreshNotifier(Ref ref) {
    ref.listen(sessionProvider, (previous, next) {
      if (previous?.status != next.status) notifyListeners();
    });
  }
}

final _routerRefreshProvider = Provider<_RouterRefreshNotifier>((ref) => _RouterRefreshNotifier(ref));

/// Root switches its whole tree based on session status (see
/// SessionStatus in providers/session_provider.dart), restored once at boot
/// - no route in any branch needs to know about the others or navigate
/// between them. Ported from src/App.js's RootNavigator.
final routerProvider = Provider<GoRouter>((ref) {
  final refresh = ref.watch(_routerRefreshProvider);

  return GoRouter(
    initialLocation: '/loading',
    refreshListenable: refresh,
    redirect: (context, state) {
      final status = ref.read(sessionProvider).status;
      final loc = state.matchedLocation;

      switch (status) {
        case SessionStatus.restoring:
          return loc == '/loading' ? null : '/loading';
        case SessionStatus.missing:
          return loc.startsWith('/onboarding') ? null : '/onboarding/connect';
        case SessionStatus.needsOrg:
          return loc == '/pick-organization' ? null : '/pick-organization';
        case SessionStatus.connected:
          final isOnboardingRoute =
              loc == '/loading' || loc.startsWith('/onboarding') || loc == '/pick-organization';
          return isOnboardingRoute ? '/main' : null;
      }
    },
    routes: [
      GoRoute(path: '/loading', builder: (context, state) => const LoadingScreen()),
      GoRoute(path: '/onboarding/connect', builder: (context, state) => const ConnectScreen()),
      GoRoute(path: '/onboarding/scan-qr', builder: (context, state) => const ScanQrScreen()),
      GoRoute(path: '/onboarding/manual-entry', builder: (context, state) => const ManualEntryScreen()),
      GoRoute(path: '/pick-organization', builder: (context, state) => const OrgPickerScreen()),
      GoRoute(path: '/switch-organization', builder: (context, state) => const OrgPickerScreen()),
      GoRoute(path: '/main', builder: (context, state) => const MainTabsScreen()),
      GoRoute(
        path: '/edit-record',
        builder: (context, state) => EditRecordScreen(entry: state.extra as TimeEntry),
      ),
      GoRoute(path: '/all-day-record', builder: (context, state) => const AllDayRecordScreen()),
      GoRoute(
        path: '/pdf-viewer',
        builder: (context, state) {
          final extra = state.extra as Map<String, dynamic>;
          return PdfViewerScreen(path: extra['path'] as String, title: extra['title'] as String?);
        },
      ),
      GoRoute(path: '/organization', builder: (context, state) => const OrganizationScreen()),
      GoRoute(path: '/members', builder: (context, state) => const MembersScreen()),
      GoRoute(path: '/members/invite', builder: (context, state) => const InviteMemberScreen()),
      GoRoute(path: '/clients', builder: (context, state) => const ClientsScreen()),
      GoRoute(
        path: '/clients/form',
        builder: (context, state) => ClientFormScreen(client: state.extra as Client?),
      ),
      GoRoute(path: '/projects', builder: (context, state) => const ProjectsScreen()),
      GoRoute(
        path: '/projects/form',
        builder: (context, state) => ProjectFormScreen(project: state.extra as Project?),
      ),
      GoRoute(path: '/export-jobs', builder: (context, state) => const ExportJobsScreen()),
    ],
  );
});
