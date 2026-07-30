import '../models/member.dart';
import '../models/organization.dart';
import '../models/user.dart';

/// A global superuser is granted an implicit owner role in every organization
/// by the backend, so this mirrors that client-side rather than requiring an
/// actual organization_members row.
bool isSuperadmin(User? user) => user?.role == 'superuser';

String? memberRole(User? user, List<Member> members) {
  for (final m in members) {
    if (m.userId == user?.id) return m.role;
  }
  return null;
}

bool isAdminOrOwner(User? user, List<Member> members) {
  if (isSuperadmin(user)) return true;
  final role = memberRole(user, members);
  return role == 'admin' || role == 'owner';
}

/// Gates owner-only actions (invite a member, change a member's role, edit
/// the organization profile) - ownership is a property of the organization
/// itself (ownerId), not the membership row, so it takes the org rather than
/// the members list.
bool isOrgOwner(User? user, Organization? org) =>
    isSuperadmin(user) || (org != null && org.ownerId == user?.id);
