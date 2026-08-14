import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../api/api_client.dart';
import '../../common/user_label.dart';
import '../../i18n/app_localizations.dart';
import '../../models/user.dart';
import '../../providers/admin_provider.dart';
import '../../providers/locale_provider.dart';
import '../../providers/session_provider.dart';
import '../../theme.dart';
import '../../widgets/app_top_bar.dart';
import '../../widgets/data_uri_image.dart';

const _statuses = ['superuser', 'confirmed', 'disabled', 'ban'];

/// Superuser-only management of every account: change a user's global status
/// (edit icon) and delete one (confirm modal). Ported from cwclock-ui's
/// src/Components/Admin/Admin.jsx.
class AdminUsersScreen extends ConsumerStatefulWidget {
  const AdminUsersScreen({super.key});

  @override
  ConsumerState<AdminUsersScreen> createState() => _AdminUsersScreenState();
}

class _AdminUsersScreenState extends ConsumerState<AdminUsersScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() => ref.read(adminProvider.notifier).listUsers());
  }

  String _statusLabel(
    String role,
    String Function(String, [Map<String, String>?]) t,
  ) {
    final key = role.isEmpty ? 'confirmed' : role;
    final capitalized = '${key[0].toUpperCase()}${key.substring(1)}';
    return t('admin.globalRole$capitalized');
  }

  void _showError(String message) {
    final t = translateWith(ref.read(localeProvider));
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(t('common.retry')),
        content: Text(message),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(t('common.close')),
          ),
        ],
      ),
    );
  }

  void _handleChangeStatus(User user) {
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(t('admin.changeStatus')),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              user.email,
              style: TextStyle(color: AppColors.of(context).textMuted),
            ),
            SizedBox(height: AppSpacing.of(1)),
            for (final status in _statuses)
              ListTile(
                contentPadding: EdgeInsets.zero,
                title: Text(_statusLabel(status, t)),
                trailing: user.role == status
                    ? Icon(Icons.check, color: AppColors.of(context).primary)
                    : null,
                onTap: () async {
                  Navigator.pop(context);
                  if (user.role == status) return;
                  try {
                    await ref
                        .read(adminProvider.notifier)
                        .updateUserStatus(user, status);
                  } catch (e) {
                    if (mounted) {
                      _showError(apiErrorMessage(asApiException(e), locale));
                    }
                  }
                },
              ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(t('common.cancel')),
          ),
        ],
      ),
    );
  }

  void _handleDelete(User user) {
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(t('admin.deleteUserTitle')),
        content: Text(t('admin.deleteUserBody', {'email': user.email})),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(t('common.cancel')),
          ),
          TextButton(
            onPressed: () async {
              Navigator.pop(context);
              try {
                await ref.read(adminProvider.notifier).deleteUser(user.id);
              } catch (e) {
                if (mounted) {
                  _showError(apiErrorMessage(asApiException(e), locale));
                }
              }
            },
            child: Text(
              t('common.delete'),
              style: TextStyle(color: AppColors.of(context).danger),
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);
    final adminState = ref.watch(adminProvider);
    final currentUserId = ref.watch(sessionProvider).user?.id;
    final users = adminState.users;

    return Scaffold(
      appBar: AppTopBar(title: t('admin.usersTitle')),
      body: SafeArea(
        child: adminState.usersLoading && users.isEmpty
            ? const Center(child: CircularProgressIndicator())
            : users.isEmpty
            ? Padding(
                padding: EdgeInsets.symmetric(horizontal: AppSpacing.of(2)),
                child: Text(
                  t('admin.noUsers'),
                  style: TextStyle(color: AppColors.of(context).textMuted),
                ),
              )
            : ListView.separated(
                padding: EdgeInsets.symmetric(horizontal: AppSpacing.of(2)),
                itemCount: users.length,
                separatorBuilder: (_, _) =>
                    Divider(height: 1, color: AppColors.of(context).border),
                itemBuilder: (context, index) {
                  final user = users[index];
                  final isSelf = user.id == currentUserId;
                  return ListTile(
                    contentPadding: EdgeInsets.zero,
                    leading: DataUriImage(
                      picture: user.picture,
                      pictureX: user.pictureX,
                      pictureY: user.pictureY,
                      size: 40,
                      fallback: CircleAvatar(
                        radius: 20,
                        backgroundColor: AppColors.of(context).backgroundMuted,
                        child: Icon(
                          Icons.person_outline,
                          color: AppColors.of(context).textMuted,
                        ),
                      ),
                    ),
                    title: Text(
                      userLabel(user),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    subtitle: Text(
                      '${user.email} · ${_statusLabel(user.role, t)}',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    trailing: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        IconButton(
                          onPressed: () => _handleChangeStatus(user),
                          icon: const Icon(Icons.edit_outlined),
                          tooltip: t('admin.changeStatus'),
                        ),
                        IconButton(
                          // The superuser can't delete their own account
                          // (server-side too) - disable rather than hide so
                          // the row stays visually consistent.
                          onPressed: isSelf ? null : () => _handleDelete(user),
                          icon: Icon(
                            Icons.delete_outline,
                            color: isSelf ? null : AppColors.of(context).danger,
                          ),
                          tooltip: t('common.delete'),
                        ),
                      ],
                    ),
                  );
                },
              ),
      ),
    );
  }
}
