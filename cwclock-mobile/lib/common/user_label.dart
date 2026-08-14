import '../models/user.dart';

/// Renders a user's display name, falling back to their email when no
/// name/surname has been set. Mirrors common/member_label.dart for the
/// [User]-shaped rows of the admin user-management screen.
String userLabel(User user) {
  return user.name?.isNotEmpty == true || user.surname?.isNotEmpty == true
      ? '${user.name ?? ''} ${user.surname ?? ''}'.trim()
      : user.email;
}
