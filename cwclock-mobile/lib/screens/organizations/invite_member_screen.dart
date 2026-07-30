import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../providers/locale_provider.dart';
import '../../providers/organizations_provider.dart';
import '../../providers/session_provider.dart';
import '../../widgets/app_button.dart';
import '../../widgets/app_screen.dart';
import '../../widgets/error_banner.dart';
import '../../widgets/form_field.dart';
import '../../widgets/select_field.dart';
import '../../widgets/app_top_bar.dart';

/// Ported from src/screens/organizations/InviteMemberScreen.js.
class InviteMemberScreen extends ConsumerStatefulWidget {
  const InviteMemberScreen({super.key});

  @override
  ConsumerState<InviteMemberScreen> createState() => _InviteMemberScreenState();
}

class _InviteMemberScreenState extends ConsumerState<InviteMemberScreen> {
  String _email = '';
  String _role = 'member';
  String? _error;
  bool _saving = false;

  Future<void> _handleInvite() async {
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;
    final locale = ref.read(localeProvider);

    setState(() {
      _error = null;
      _saving = true;
    });
    try {
      await ref.read(organizationsProvider.notifier).addMember(orgId, _email.trim(), _role);
      if (mounted) Navigator.of(context).pop();
    } catch (e) {
      setState(() => _error = apiErrorMessage(asApiException(e), locale));
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);

    return Scaffold(
      appBar: AppTopBar(title: t('organizations.inviteMember')),
      body: AppScreen(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            AppFormField(
              label: t('organizations.email'),
              value: _email,
              onChanged: (v) => setState(() => _email = v),
              keyboardType: TextInputType.emailAddress,
              textCapitalization: TextCapitalization.none,
            ),
            AppSelectField(
              label: t('organizations.role'),
              value: _role,
              onChanged: (v) => setState(() => _role = v),
              items: [
                SelectItem('admin', t('organizations.roleAdmin')),
                SelectItem('member', t('organizations.roleMember')),
                SelectItem('reader', t('organizations.roleReader')),
              ],
            ),
            ErrorBanner(message: _error),
            AppButton(
              title: t('organizations.inviteMember'),
              onPressed: _email.trim().isEmpty ? null : _handleInvite,
              loading: _saving,
            ),
          ],
        ),
      ),
    );
  }
}
