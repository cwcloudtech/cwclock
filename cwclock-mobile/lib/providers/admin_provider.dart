import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/organization.dart';
import '../models/user.dart';
import 'api_providers.dart';

/// Superuser-only management of every account and organization, backed by the
/// `/admin/*` endpoints (RequireSuperuser server-side). Ported from
/// cwclock-ui's src/Redux/Admin/Admin.{actions,reducer}.js + the org
/// delete/transfer actions it reuses from Org.actions.js.
class AdminState {
  final List<User> users;
  final List<Organization> organizations;
  final bool usersLoading;
  final bool orgsLoading;

  const AdminState({
    this.users = const [],
    this.organizations = const [],
    this.usersLoading = false,
    this.orgsLoading = false,
  });

  AdminState copyWith({
    List<User>? users,
    List<Organization>? organizations,
    bool? usersLoading,
    bool? orgsLoading,
  }) {
    return AdminState(
      users: users ?? this.users,
      organizations: organizations ?? this.organizations,
      usersLoading: usersLoading ?? this.usersLoading,
      orgsLoading: orgsLoading ?? this.orgsLoading,
    );
  }
}

class AdminNotifier extends Notifier<AdminState> {
  @override
  AdminState build() => const AdminState();

  Future<List<User>> listUsers() async {
    state = state.copyWith(usersLoading: true);
    try {
      final response = await ref
          .read(apiClientProvider)
          .dio
          .get('/admin/users/');
      final users = (response.data as List)
          .map((e) => User.fromJson(e as Map<String, dynamic>))
          .toList();
      state = state.copyWith(users: users, usersLoading: false);
      return users;
    } catch (e) {
      state = state.copyWith(usersLoading: false);
      rethrow;
    }
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
    state = state.copyWith(
      users: state.users.where((u) => u.id != id).toList(),
    );
  }

  Future<List<Organization>> listOrganizations() async {
    state = state.copyWith(orgsLoading: true);
    try {
      final response = await ref
          .read(apiClientProvider)
          .dio
          .get('/admin/organizations/');
      final orgs = (response.data as List)
          .map((e) => Organization.fromJson(e as Map<String, dynamic>))
          .toList();
      state = state.copyWith(organizations: orgs, orgsLoading: false);
      return orgs;
    } catch (e) {
      state = state.copyWith(orgsLoading: false);
      rethrow;
    }
  }

  /// Deletes an organization via the regular owner-only endpoint - a superuser
  /// is granted an implicit owner role in every org server-side.
  Future<void> deleteOrganization(String id) async {
    await ref.read(apiClientProvider).dio.delete('/organizations/$id');
    state = state.copyWith(
      organizations: state.organizations.where((o) => o.id != id).toList(),
    );
  }

  /// Transfers an org to a new owner (by email), then re-lists so the row picks
  /// up the fresh ownerEmail - the transfer response is a plain Organization
  /// without it.
  Future<void> transferOwnership(String orgId, String email) async {
    await ref
        .read(apiClientProvider)
        .dio
        .put('/organizations/$orgId/owner', data: {'email': email});
    await listOrganizations();
  }
}

final adminProvider = NotifierProvider<AdminNotifier, AdminState>(
  AdminNotifier.new,
);
