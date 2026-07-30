import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../models/member.dart';
import '../../providers/locale_provider.dart';
import '../../providers/organizations_provider.dart';
import '../../providers/permissions.dart' as perm;
import '../../providers/session_provider.dart';
import '../../theme.dart';
import '../../widgets/app_button.dart';

const _roles = ['admin', 'member', 'reader'];

/// List members with their role; invite (owner-only), change an existing
/// member's role (owner-only), remove a member (admin/owner). The owner's
/// own row exposes neither action. Ported from
/// src/screens/organizations/MembersScreen.js.
class MembersScreen extends ConsumerStatefulWidget {
  const MembersScreen({super.key});

  @override
  ConsumerState<MembersScreen> createState() => _MembersScreenState();
}

class _MembersScreenState extends ConsumerState<MembersScreen> {
  @override
  void initState() {
    super.initState();
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId != null) {
      Future.microtask(() => ref.read(organizationsProvider.notifier).listMembers(orgId));
    }
  }

  String _roleLabel(String role, String Function(String, [Map<String, String>?]) t) {
    final capitalized = '${role[0].toUpperCase()}${role.substring(1)}';
    return t('organizations.role$capitalized');
  }

  void _handleChangeRole(Member member) {
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(t('organizations.changeRole')),
        content: Text(member.email),
        actions: [
          for (final role in _roles)
            TextButton(
              onPressed: () async {
                Navigator.pop(context);
                try {
                  await ref.read(organizationsProvider.notifier).updateMemberRole(orgId, member.userId, role);
                } catch (e) {
                  if (mounted) {
                    _showError(apiErrorMessage(asApiException(e), locale));
                  }
                }
              },
              child: Text(_roleLabel(role, t)),
            ),
          TextButton(onPressed: () => Navigator.pop(context), child: Text(t('common.cancel'))),
        ],
      ),
    );
  }

  void _showError(String message) {
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(t('common.retry')),
        content: Text(message),
        actions: [TextButton(onPressed: () => Navigator.pop(context), child: Text(t('common.close')))],
      ),
    );
  }

  void _handleRemove(Member member) {
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(t('organizations.removeMemberTitle')),
        content: Text(t('organizations.removeMemberBody', {'name': member.email})),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: Text(t('common.cancel'))),
          TextButton(
            onPressed: () async {
              Navigator.pop(context);
              try {
                await ref.read(organizationsProvider.notifier).removeMember(orgId, member.userId);
              } catch (e) {
                if (mounted) _showError(apiErrorMessage(asApiException(e), locale));
              }
            },
            child: Text(t('organizations.removeMember'), style: TextStyle(color: AppColors.of(context).danger)),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);
    final session = ref.watch(sessionProvider);
    final orgsState = ref.watch(organizationsProvider);

    final org = orgsState.items.where((o) => o.id == session.orgId).firstOrNull;
    final canInvite = perm.isOrgOwner(session.user, org);
    final canRemove = perm.isAdminOrOwner(session.user, orgsState.members);
    final members = orgsState.members;

    return Scaffold(
      appBar: AppBar(title: Text(t('organizations.membersTitle'))),
      body: SafeArea(
        child: Column(
          children: [
            if (canInvite)
              Padding(
                padding: EdgeInsets.all(AppSpacing.of(2)),
                child: AppButton(
                  title: t('organizations.inviteMember'),
                  onPressed: () => context.push('/members/invite'),
                ),
              ),
            Expanded(
              child: members.isEmpty
                  ? Padding(
                      padding: EdgeInsets.symmetric(horizontal: AppSpacing.of(2)),
                      child: Text(t('organizations.noMembers'), style: TextStyle(color: AppColors.of(context).textMuted)),
                    )
                  : ListView.separated(
                      padding: EdgeInsets.symmetric(horizontal: AppSpacing.of(2)),
                      itemCount: members.length,
                      separatorBuilder: (_, _) => Divider(height: 1, color: AppColors.of(context).border),
                      itemBuilder: (context, index) {
                        final member = members[index];
                        final isOwnerRow = member.role == 'owner';
                        final displayName = (member.name?.isNotEmpty ?? false) || (member.surname?.isNotEmpty ?? false)
                            ? '${member.name ?? ''} ${member.surname ?? ''}'.trim()
                            : member.email;
                        return ListTile(
                          contentPadding: EdgeInsets.zero,
                          title: Text(displayName, maxLines: 1, overflow: TextOverflow.ellipsis),
                          subtitle: Text(
                            '${member.email} · ${_roleLabel(member.role, t)}',
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                          trailing: isOwnerRow
                              ? null
                              : Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    if (canInvite)
                                      TextButton(
                                        onPressed: () => _handleChangeRole(member),
                                        child: Text(t('organizations.changeRole')),
                                      ),
                                    if (canRemove)
                                      TextButton(
                                        onPressed: () => _handleRemove(member),
                                        child: Text(
                                          t('organizations.removeMember'),
                                          style: TextStyle(color: AppColors.of(context).danger),
                                        ),
                                      ),
                                  ],
                                ),
                        );
                      },
                    ),
            ),
          ],
        ),
      ),
    );
  }
}

extension _FirstOrNull<T> on Iterable<T> {
  T? get firstOrNull {
    final it = iterator;
    return it.moveNext() ? it.current : null;
  }
}
