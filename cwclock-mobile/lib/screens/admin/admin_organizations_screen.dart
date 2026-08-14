import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../models/organization.dart';
import '../../providers/admin_provider.dart';
import '../../providers/locale_provider.dart';
import '../../theme.dart';
import '../../widgets/app_top_bar.dart';
import '../../widgets/data_uri_image.dart';

/// Superuser-only management of every organization: transfer ownership
/// (exchange icon) and delete one (confirm modal). Ported from cwclock-ui's
/// src/Components/Admin/Organizations.jsx.
class AdminOrganizationsScreen extends ConsumerStatefulWidget {
  const AdminOrganizationsScreen({super.key});

  @override
  ConsumerState<AdminOrganizationsScreen> createState() =>
      _AdminOrganizationsScreenState();
}

class _AdminOrganizationsScreenState
    extends ConsumerState<AdminOrganizationsScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(
      () => ref.read(adminProvider.notifier).listOrganizations(),
    );
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

  Future<void> _handleTransfer(Organization org) async {
    final locale = ref.read(localeProvider);

    final email = await showDialog<String>(
      context: context,
      builder: (context) => _TransferOwnershipDialog(orgName: org.name),
    );
    if (email == null || email.isEmpty) return;

    try {
      await ref.read(adminProvider.notifier).transferOwnership(org.id, email);
    } catch (e) {
      if (mounted) _showError(apiErrorMessage(asApiException(e), locale));
    }
  }

  void _handleDelete(Organization org) {
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(t('admin.deleteOrgTitle')),
        content: Text(t('admin.deleteOrgBody', {'name': org.name})),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(t('common.cancel')),
          ),
          TextButton(
            onPressed: () async {
              Navigator.pop(context);
              try {
                await ref
                    .read(adminProvider.notifier)
                    .deleteOrganization(org.id);
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
    final orgs = adminState.organizations;

    return Scaffold(
      appBar: AppTopBar(title: t('admin.organizationsTitle')),
      body: SafeArea(
        child: adminState.orgsLoading && orgs.isEmpty
            ? const Center(child: CircularProgressIndicator())
            : orgs.isEmpty
            ? Padding(
                padding: EdgeInsets.symmetric(horizontal: AppSpacing.of(2)),
                child: Text(
                  t('admin.noOrganizations'),
                  style: TextStyle(color: AppColors.of(context).textMuted),
                ),
              )
            : ListView.separated(
                padding: EdgeInsets.symmetric(horizontal: AppSpacing.of(2)),
                itemCount: orgs.length,
                separatorBuilder: (_, _) =>
                    Divider(height: 1, color: AppColors.of(context).border),
                itemBuilder: (context, index) {
                  final org = orgs[index];
                  return ListTile(
                    contentPadding: EdgeInsets.zero,
                    leading: DataUriImage(
                      picture: org.picture,
                      pictureX: org.pictureX,
                      pictureY: org.pictureY,
                      size: 40,
                      borderRadius: 8,
                      fallback: Container(
                        width: 40,
                        height: 40,
                        decoration: BoxDecoration(
                          color: AppColors.of(context).backgroundMuted,
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Icon(
                          Icons.business_outlined,
                          color: AppColors.of(context).textMuted,
                        ),
                      ),
                    ),
                    title: Text(
                      org.name,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    subtitle: org.ownerEmail?.isNotEmpty == true
                        ? Text(
                            org.ownerEmail!,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          )
                        : null,
                    trailing: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        IconButton(
                          onPressed: () => _handleTransfer(org),
                          icon: const Icon(Icons.swap_horiz),
                          tooltip: t('admin.transferOwnership'),
                        ),
                        IconButton(
                          onPressed: () => _handleDelete(org),
                          icon: Icon(
                            Icons.delete_outline,
                            color: AppColors.of(context).danger,
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

/// Prompts for the new owner's email and pops it back (or null on cancel).
/// Kept stateful so the text controller is disposed with the dialog.
class _TransferOwnershipDialog extends ConsumerStatefulWidget {
  final String orgName;

  const _TransferOwnershipDialog({required this.orgName});

  @override
  ConsumerState<_TransferOwnershipDialog> createState() =>
      _TransferOwnershipDialogState();
}

class _TransferOwnershipDialogState
    extends ConsumerState<_TransferOwnershipDialog> {
  final _controller = TextEditingController();

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final t = translateWith(ref.watch(localeProvider));

    return AlertDialog(
      title: Text(t('admin.transferModalTitle', {'name': widget.orgName})),
      content: TextField(
        controller: _controller,
        autofocus: true,
        keyboardType: TextInputType.emailAddress,
        decoration: InputDecoration(labelText: t('admin.newOwnerEmail')),
        onSubmitted: (value) => Navigator.pop(context, value.trim()),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(t('common.cancel')),
        ),
        TextButton(
          onPressed: () => Navigator.pop(context, _controller.text.trim()),
          child: Text(
            t('admin.transfer'),
            style: TextStyle(color: AppColors.of(context).danger),
          ),
        ),
      ],
    );
  }
}
