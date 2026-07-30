import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../i18n/app_localizations.dart';
import '../../providers/locale_provider.dart';
import '../../providers/organizations_provider.dart';
import '../../providers/permissions.dart' as perm;
import '../../providers/session_provider.dart';
import '../../providers/theme_mode_provider.dart';
import '../../theme.dart';
import '../../widgets/app_button.dart';
import '../../widgets/app_screen.dart';
import '../../widgets/data_uri_image.dart';
import '../../widgets/app_top_bar.dart';

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
            child: Text(t('settings.disconnect'), style: TextStyle(color: AppColors.of(context).danger)),
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
    // Shows the language tapping switches TO, not the current one - e.g.
    // "English" while French is loaded (ai-instruct-121), same convention as
    // the dark/light mode button below.
    final otherLocaleCode = locale == 'en' ? 'fr' : 'en';
    final otherLocaleLabel =
        supportedLocales.where((l) => l.code == otherLocaleCode).firstOrNull?.label ?? 'English';
    ref.watch(themeModeProvider);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      appBar: AppTopBar(title: t('settings.title')),
      body: AppScreen(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(t('settings.connectedTo'), style: TextStyle(fontSize: 13, color: AppColors.of(context).textMuted)),
            Text(session.apiUrl ?? '', style: TextStyle(fontSize: 17, color: AppColors.of(context).text)),
            SizedBox(height: AppSpacing.of(1.5)),
            Text(t('settings.organization'), style: TextStyle(fontSize: 13, color: AppColors.of(context).textMuted)),
            Padding(
              padding: EdgeInsets.only(top: 4),
              child: Row(
                children: [
                  DataUriImage(
                    picture: currentOrg?.picture,
                    pictureX: currentOrg?.pictureX ?? 50,
                    pictureY: currentOrg?.pictureY ?? 50,
                    size: 40,
                    borderRadius: 8,
                    fallback: Container(
                      width: 40,
                      height: 40,
                      decoration: BoxDecoration(
                        color: AppColors.of(context).backgroundMuted,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Icon(Icons.business_outlined, color: AppColors.of(context).textMuted),
                    ),
                  ),
                  SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      currentOrg?.name ?? session.orgId ?? '',
                      style: TextStyle(fontSize: 17, color: AppColors.of(context).text),
                    ),
                  ),
                ],
              ),
            ),
            if (session.user?.email != null) ...[
              SizedBox(height: AppSpacing.of(1.5)),
              Text(session.user!.email, style: TextStyle(fontSize: 13, color: AppColors.of(context).textMuted)),
            ],
            AppButton(
              title: t('settings.switchOrganization'),
              variant: AppButtonVariant.secondary,
              onPressed: () => context.push('/switch-organization'),
              margin: EdgeInsets.only(top: AppSpacing.of(1.5)),
            ),
            AppButton(
              title: otherLocaleLabel,
              variant: AppButtonVariant.secondary,
              onPressed: () => ref.read(localeProvider.notifier).setLocale(otherLocaleCode),
              margin: EdgeInsets.only(top: AppSpacing.of(1.5)),
            ),
            AppButton(
              // Shows the mode tapping switches TO, not the current one -
              // e.g. "Light mode" while dark mode is active (ai-instruct-120).
              title: isDark ? t('settings.lightMode') : t('settings.darkMode'),
              icon: isDark ? Icons.light_mode_outlined : Icons.dark_mode_outlined,
              variant: AppButtonVariant.secondary,
              onPressed: () => ref.read(themeModeProvider.notifier).toggle(),
              margin: EdgeInsets.only(top: AppSpacing.of(1.5)),
            ),
            if (canManage) ...[
              Container(
                margin: EdgeInsets.only(top: AppSpacing.of(3)),
                padding: EdgeInsets.only(top: AppSpacing.of(2)),
                decoration: BoxDecoration(border: Border(top: BorderSide(color: AppColors.of(context).border))),
                child: Text(
                  t('management.title'),
                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: AppColors.of(context).textMuted),
                ),
              ),
              AppButton(
                // Icons match cwclock-ui's SidebarNav.jsx (FaBuilding,
                // FaRegUserCircle, FaFileAlt - ai-instruct-120); there's no
                // dedicated "members" nav icon on the web app (membership is
                // managed inside the Organization page there), so this one
                // uses the closest generic "people" icon instead.
                title: t('management.organization'),
                icon: Icons.business_outlined,
                variant: AppButtonVariant.secondary,
                onPressed: () => context.push('/organization'),
                margin: EdgeInsets.only(top: AppSpacing.of(1.5)),
              ),
              AppButton(
                title: t('management.members'),
                icon: Icons.people_outline,
                variant: AppButtonVariant.secondary,
                onPressed: () => context.push('/members'),
                margin: EdgeInsets.only(top: AppSpacing.of(1.5)),
              ),
              AppButton(
                title: t('management.clients'),
                icon: Icons.account_circle_outlined,
                variant: AppButtonVariant.secondary,
                onPressed: () => context.push('/clients'),
                margin: EdgeInsets.only(top: AppSpacing.of(1.5)),
              ),
              AppButton(
                title: t('management.projects'),
                icon: Icons.description_outlined,
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
