import '../models/member.dart';

/// Renders a member's display name, falling back to their email when no
/// name/surname has been set yet. Mirrors cwclock-ui's memberLabel.js.
String memberLabel(Member member) {
  return member.name?.isNotEmpty == true || member.surname?.isNotEmpty == true
      ? '${member.name ?? ''} ${member.surname ?? ''}'.trim()
      : member.email;
}
