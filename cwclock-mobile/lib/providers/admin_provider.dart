import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/organization.dart';
import '../models/user.dart';
import 'api_providers.dart';

/// Page size requested from the admin list endpoints - matches the backend's
/// default (handlers.defaultAdminPageSize) and cwclock-ui's ADMIN_PAGE_SIZE.
const _adminPageSize = 20;

/// Superuser-only management of every account and organization, backed by the
/// `/admin/*` endpoints (RequireSuperuser server-side). Ported from
/// cwclock-ui's src/Redux/Admin/Admin.{actions,reducer}.js + the org
/// delete/transfer actions it reuses from Org.actions.js.
///
/// The list endpoints return the paged SearchResult envelope
/// ({startIndex, maxResults, totalResults, results}) and accept a ?term=
/// filter (name/surname/e-mail/id) - orgs also filter on name/owner/id - and
/// ?startIndex (0-based page) / ?maxResults, so this provider tracks the term,
/// the current page and the total behind the filter to drive the screens'
/// search field + infinite scroll (mirrors the web Admin.jsx/Organizations.jsx).
class AdminState {
  final List<User> users;
  final List<Organization> organizations;
  final bool usersLoading;
  final bool orgsLoading;
  final bool usersLoadingMore;
  final bool orgsLoadingMore;
  final String usersTerm;
  final String orgsTerm;
  final int usersPage;
  final int orgsPage;
  final int usersTotal;
  final int orgsTotal;

  const AdminState({
    this.users = const [],
    this.organizations = const [],
    this.usersLoading = false,
    this.orgsLoading = false,
    this.usersLoadingMore = false,
    this.orgsLoadingMore = false,
    this.usersTerm = '',
    this.orgsTerm = '',
    this.usersPage = 0,
    this.orgsPage = 0,
    this.usersTotal = 0,
    this.orgsTotal = 0,
  });

  /// More pages to fetch while fewer rows are loaded than the filter's total.
  bool get usersHasMore => users.length < usersTotal;
  bool get orgsHasMore => organizations.length < orgsTotal;

  AdminState copyWith({
    List<User>? users,
    List<Organization>? organizations,
    bool? usersLoading,
    bool? orgsLoading,
    bool? usersLoadingMore,
    bool? orgsLoadingMore,
    String? usersTerm,
    String? orgsTerm,
    int? usersPage,
    int? orgsPage,
    int? usersTotal,
    int? orgsTotal,
  }) {
    return AdminState(
      users: users ?? this.users,
      organizations: organizations ?? this.organizations,
      usersLoading: usersLoading ?? this.usersLoading,
      orgsLoading: orgsLoading ?? this.orgsLoading,
      usersLoadingMore: usersLoadingMore ?? this.usersLoadingMore,
      orgsLoadingMore: orgsLoadingMore ?? this.orgsLoadingMore,
      usersTerm: usersTerm ?? this.usersTerm,
      orgsTerm: orgsTerm ?? this.orgsTerm,
      usersPage: usersPage ?? this.usersPage,
      orgsPage: orgsPage ?? this.orgsPage,
      usersTotal: usersTotal ?? this.usersTotal,
      orgsTotal: orgsTotal ?? this.orgsTotal,
    );
  }
}

class AdminNotifier extends Notifier<AdminState> {
  @override
  AdminState build() => const AdminState();

  /// Fetches one page of accounts. startIndex 0 replaces the list (initial
  /// load, pull-to-refresh or a new search term), startIndex > 0 appends
  /// (infinite scroll). A null [term] keeps the current filter so paging
  /// doesn't drop it.
  Future<List<User>> listUsers({String? term, int startIndex = 0}) async {
    final resolvedTerm = term ?? state.usersTerm;
    state = startIndex == 0
        ? state.copyWith(usersLoading: true, usersTerm: resolvedTerm)
        : state.copyWith(usersLoadingMore: true);
    try {
      final response = await ref.read(apiClientProvider).dio.get(
        '/admin/users/',
        queryParameters: {
          if (resolvedTerm.isNotEmpty) 'term': resolvedTerm,
          'startIndex': startIndex,
          'maxResults': _adminPageSize,
        },
      );
      final data = response.data as Map<String, dynamic>;
      final page = (data['results'] as List? ?? [])
          .map((e) => User.fromJson(e as Map<String, dynamic>))
          .toList();
      final total = (data['totalResults'] as num?)?.toInt() ?? page.length;
      final users = startIndex == 0 ? page : [...state.users, ...page];
      state = state.copyWith(
        users: users,
        usersLoading: false,
        usersLoadingMore: false,
        usersPage: startIndex,
        usersTotal: total,
      );
      return users;
    } catch (e) {
      state = state.copyWith(usersLoading: false, usersLoadingMore: false);
      rethrow;
    }
  }

  /// Re-runs the user list from page 0 with [term] (debounced search field).
  Future<void> searchUsers(String term) => listUsers(term: term, startIndex: 0);

  /// Appends the next page of users (infinite scroll), a no-op while a page is
  /// already in flight or the full result set is loaded.
  Future<void> loadMoreUsers() async {
    if (state.usersLoadingMore || !state.usersHasMore) return;
    await listUsers(startIndex: state.usersPage + 1);
  }

  /// Changes a user's global status/role. The backend's admin update is a full
  /// PUT that also requires email/name/surname, so this resends the account's
  /// existing values alongside the new role (registration guarantees name and
  /// surname are set, so they're never blank here).
  Future<User> updateUserStatus(User user, String role) async {
    final response = await ref
        .read(apiClientProvider)
        .dio
        .put(
          '/admin/users/${user.id}',
          data: {
            'email': user.email,
            'name': user.name ?? '',
            'surname': user.surname ?? '',
            'role': role,
          },
        );
    final updated = User.fromJson(response.data as Map<String, dynamic>);
    state = state.copyWith(
      users: [for (final u in state.users) u.id == updated.id ? updated : u],
    );
    return updated;
  }

  Future<void> deleteUser(String id) async {
    await ref.read(apiClientProvider).dio.delete('/admin/users/$id');
    final remaining = state.users.where((u) => u.id != id).toList();
    state = state.copyWith(
      users: remaining,
      usersTotal: state.usersTotal > 0 ? state.usersTotal - 1 : 0,
    );
  }

  /// Fetches one page of organizations. Same paging/append semantics as
  /// [listUsers]; ?term= filters on name/owner e-mail/id server-side.
  Future<List<Organization>> listOrganizations({
    String? term,
    int startIndex = 0,
  }) async {
    final resolvedTerm = term ?? state.orgsTerm;
    state = startIndex == 0
        ? state.copyWith(orgsLoading: true, orgsTerm: resolvedTerm)
        : state.copyWith(orgsLoadingMore: true);
    try {
      final response = await ref.read(apiClientProvider).dio.get(
        '/admin/organizations/',
        queryParameters: {
          if (resolvedTerm.isNotEmpty) 'term': resolvedTerm,
          'startIndex': startIndex,
          'maxResults': _adminPageSize,
        },
      );
      final data = response.data as Map<String, dynamic>;
      final page = (data['results'] as List? ?? [])
          .map((e) => Organization.fromJson(e as Map<String, dynamic>))
          .toList();
      final total = (data['totalResults'] as num?)?.toInt() ?? page.length;
      final orgs = startIndex == 0 ? page : [...state.organizations, ...page];
      state = state.copyWith(
        organizations: orgs,
        orgsLoading: false,
        orgsLoadingMore: false,
        orgsPage: startIndex,
        orgsTotal: total,
      );
      return orgs;
    } catch (e) {
      state = state.copyWith(orgsLoading: false, orgsLoadingMore: false);
      rethrow;
    }
  }

  /// Re-runs the org list from page 0 with [term] (debounced search field).
  Future<void> searchOrganizations(String term) =>
      listOrganizations(term: term, startIndex: 0);

  /// Appends the next page of organizations (infinite scroll).
  Future<void> loadMoreOrganizations() async {
    if (state.orgsLoadingMore || !state.orgsHasMore) return;
    await listOrganizations(startIndex: state.orgsPage + 1);
  }

  /// Deletes an organization via the regular owner-only endpoint - a superuser
  /// is granted an implicit owner role in every org server-side.
  Future<void> deleteOrganization(String id) async {
    await ref.read(apiClientProvider).dio.delete('/organizations/$id');
    final remaining = state.organizations.where((o) => o.id != id).toList();
    state = state.copyWith(
      organizations: remaining,
      orgsTotal: state.orgsTotal > 0 ? state.orgsTotal - 1 : 0,
    );
  }

  /// Transfers an org to a new owner (by email), then reloads the current
  /// filter from the top so the row picks up the fresh ownerEmail - the
  /// transfer response is a plain Organization without it.
  Future<void> transferOwnership(String orgId, String email) async {
    await ref
        .read(apiClientProvider)
        .dio
        .put('/organizations/$orgId/owner', data: {'email': email});
    await listOrganizations(startIndex: 0);
  }
}

final adminProvider = NotifierProvider<AdminNotifier, AdminState>(
  AdminNotifier.new,
);
