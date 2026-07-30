import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../i18n/app_localizations.dart';
import '../../providers/locale_provider.dart';
import '../../providers/organizations_provider.dart';
import '../../providers/permissions.dart' as perm;
import '../../providers/session_provider.dart';
import '../../theme.dart';
import '../../widgets/app_button.dart';
import '../../widgets/app_screen.dart';

/// Ported from src/screens/settings/SettingsScreen.js.
class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({super.key});

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      ref.read(organizationsProvider.notifier).listOrganizations();
      final orgId = ref.read(sessionProvider).orgId;
      if (orgId != null) ref.read(organizationsProvider.notifier).listMembers(orgId);
    });
  }

  void _handleDisconnect() {
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(t('settings.disconnect')),
        content: Text(t('settings.disconnectConfirm')),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: Text(t('common.cancel'))),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              ref.read(sessionProvider.notifier).disconnect();
            },
            child: Text(t('settings.disconnect'), style: const TextStyle(color: AppColors.danger)),
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

    final currentOrg = orgsState.items.where((o) => o.id == session.orgId).firstOrNull;
    final canManage = perm.isAdminOrOwner(session.user, orgsState.members);
    final currentLocaleLabel =
        supportedLocales.where((l) => l.code == locale).firstOrNull?.label ?? 'English';

    return Scaffold(
      appBar: AppBar(title: Text(t('settings.title'))),
      body: AppScreen(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(t('settings.connectedTo'), style: const TextStyle(fontSize: 13, color: AppColors.textMuted)),
            Text(session.apiUrl ?? '', style: const TextStyle(fontSize: 17, color: AppColors.text)),
            SizedBox(height: AppSpacing.of(1.5)),
            Text(t('settings.organization'), style: const TextStyle(fontSize: 13, color: AppColors.textMuted)),
            Text(
              currentOrg?.name ?? session.orgId ?? '',
              style: const TextStyle(fontSize: 17, color: AppColors.text),
            ),
            if (session.user?.email != null) ...[
              SizedBox(height: AppSpacing.of(1.5)),
              Text(session.user!.email, style: const TextStyle(fontSize: 13, color: AppColors.textMuted)),
            ],
            AppButton(
              title: t('settings.switchOrganization'),
              variant: AppButtonVariant.secondary,
              onPressed: () => context.push('/switch-organization'),
              margin: EdgeInsets.only(top: AppSpacing.of(1.5)),
            ),
            AppButton(
              title: currentLocaleLabel,
              variant: AppButtonVariant.secondary,
              onPressed: () => ref.read(localeProvider.notifier).setLocale(locale == 'en' ? 'fr' : 'en'),
              margin: EdgeInsets.only(top: AppSpacing.of(1.5)),
            ),
            if (canManage) ...[
              Container(
                margin: EdgeInsets.only(top: AppSpacing.of(3)),
                padding: EdgeInsets.only(top: AppSpacing.of(2)),
                decoration: const BoxDecoration(border: Border(top: BorderSide(color: AppColors.border))),
                child: Text(
                  t('management.title'),
                  style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: AppColors.textMuted),
                ),
              ),
              AppButton(
                title: t('management.organization'),
                variant: AppButtonVariant.secondary,
                onPressed: () => context.push('/organization'),
                margin: EdgeInsets.only(top: AppSpacing.of(1.5)),
              ),
              AppButton(
                title: t('management.members'),
                variant: AppButtonVariant.secondary,
                onPressed: () => context.push('/members'),
                margin: EdgeInsets.only(top: AppSpacing.of(1.5)),
              ),
              AppButton(
                title: t('management.clients'),
                variant: AppButtonVariant.secondary,
                onPressed: () => context.push('/clients'),
                margin: EdgeInsets.only(top: AppSpacing.of(1.5)),
              ),
              AppButton(
                title: t('management.projects'),
                variant: AppButtonVariant.secondary,
                onPressed: () => context.push('/projects'),
                margin: EdgeInsets.only(top: AppSpacing.of(1.5)),
              ),
            ],
            AppButton(
              title: t('settings.disconnect'),
              variant: AppButtonVariant.danger,
              onPressed: _handleDisconnect,
              margin: EdgeInsets.only(top: AppSpacing.of(3)),
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
