import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/member.dart';
import '../models/organization.dart';
import 'api_providers.dart';

class OrganizationsState {
  final List<Organization> items;
  final List<Member> members;
  final bool isLoading;

  const OrganizationsState({
    this.items = const [],
    this.members = const [],
    this.isLoading = false,
  });

  OrganizationsState copyWith({List<Organization>? items, List<Member>? members, bool? isLoading}) {
    return OrganizationsState(
      items: items ?? this.items,
      members: members ?? this.members,
      isLoading: isLoading ?? this.isLoading,
    );
  }
}

/// Ported from src/redux/organizations/organizations.actions.js +
/// organizations.reducer.js.
class OrganizationsNotifier extends Notifier<OrganizationsState> {
  @override
  OrganizationsState build() => const OrganizationsState();

  Future<List<Organization>> listOrganizations() async {
    state = state.copyWith(isLoading: true);
    try {
      final response = await ref.read(apiClientProvider).dio.get('/organizations/');
      final items = (response.data as List)
          .map((e) => Organization.fromJson(e as Map<String, dynamic>))
          .toList();
      state = state.copyWith(items: items, isLoading: false);
      return items;
    } catch (e) {
      state = state.copyWith(isLoading: false);
      rethrow;
    }
  }

  /// Loads the current organization's members - used both to resolve the
  /// connected user's own role (see providers/permissions.dart) and by the
  /// Members screen's RBAC list.
  Future<List<Member>> listMembers(String orgId) async {
    final response = await ref.read(apiClientProvider).dio.get('/organizations/$orgId/members/');
    final members = (response.data as List).map((e) => Member.fromJson(e as Map<String, dynamic>)).toList();
    state = state.copyWith(members: members);
    return members;
  }

  /// Owner-only server-side (see OrganizationHandler.Update's route gate).
  Future<Organization> updateOrganization(String orgId, Map<String, dynamic> fields) async {
    final response = await ref.read(apiClientProvider).dio.patch('/organizations/$orgId', data: fields);
    final updated = Organization.fromJson(response.data as Map<String, dynamic>);
    state = state.copyWith(items: [for (final o in state.items) o.id == updated.id ? updated : o]);
    return updated;
  }

  /// Invites an existing cwclock user (by email) into the organization with a
  /// starting role - owner-only server-side.
  Future<Member> addMember(String orgId, String email, String role) async {
    final response = await ref
        .read(apiClientProvider)
        .dio
        .post('/organizations/$orgId/members/', data: {'email': email, 'role': role});
    final member = Member.fromJson(response.data as Map<String, dynamic>);
    state = state.copyWith(members: [...state.members, member]);
    return member;
  }

  /// Changes an existing member's role - owner-only server-side.
  Future<Member> updateMemberRole(String orgId, String userId, String role) async {
    final response = await ref
        .read(apiClientProvider)
        .dio
        .put('/organizations/$orgId/members/$userId', data: {'role': role});
    final updated = Member.fromJson(response.data as Map<String, dynamic>);
    state = state.copyWith(
      members: [for (final m in state.members) m.userId == updated.userId ? updated : m],
    );
    return updated;
  }

  Future<void> removeMember(String orgId, String userId) async {
    await ref.read(apiClientProvider).dio.delete('/organizations/$orgId/members/$userId');
    state = state.copyWith(members: state.members.where((m) => m.userId != userId).toList());
  }
}

final organizationsProvider =
    NotifierProvider<OrganizationsNotifier, OrganizationsState>(OrganizationsNotifier.new);
