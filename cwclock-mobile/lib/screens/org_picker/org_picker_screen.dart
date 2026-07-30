import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../providers/locale_provider.dart';
import '../../providers/organizations_provider.dart';
import '../../providers/session_provider.dart';
import '../../theme.dart';

/// Shown whenever session status is "needsOrg" - either onboarding didn't
/// have an org_id line (manual entry / an older QR), or the user picked
/// "Switch organization" from Settings. Selecting a row calls
/// SessionNotifier.selectOrg and the router's redirect (see router.dart)
/// reacts to the resulting status change itself. Ported from
/// src/screens/onboarding/OrgPickerScreen.js.
class OrgPickerScreen extends ConsumerStatefulWidget {
  const OrgPickerScreen({super.key});

  @override
  ConsumerState<OrgPickerScreen> createState() => _OrgPickerScreenState();
}

class _OrgPickerScreenState extends ConsumerState<OrgPickerScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() => ref.read(organizationsProvider.notifier).listOrganizations());
  }

  Future<void> _handleSelect(String orgId) async {
    await ref.read(sessionProvider.notifier).selectOrg(orgId);
    // No-op when there's nothing to go back to (first-run onboarding) - only
    // meaningful when reached from Settings' "Switch organization", where
    // session status was already "connected".
    if (mounted && Navigator.of(context).canPop()) {
      Navigator.of(context).pop();
    }
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);
    final orgsState = ref.watch(organizationsProvider);

    return Scaffold(
      appBar: AppBar(title: Text(t('onboarding.pickOrganization'))),
      body: SafeArea(
        child: orgsState.items.isEmpty && !orgsState.isLoading
            ? Padding(
                padding: EdgeInsets.all(AppSpacing.of(2)),
                child: Text(t('onboarding.noOrganizations'), style: const TextStyle(color: AppColors.textMuted)),
              )
            : ListView.separated(
                padding: EdgeInsets.symmetric(horizontal: AppSpacing.of(2)),
                itemCount: orgsState.items.length,
                separatorBuilder: (_, _) => const Divider(height: 1, color: AppColors.border),
                itemBuilder: (context, index) {
                  final org = orgsState.items[index];
                  return ListTile(
                    contentPadding: EdgeInsets.zero,
                    title: Text(org.name, style: const TextStyle(fontSize: 17, color: AppColors.text)),
                    onTap: () => _handleSelect(org.id),
                  );
                },
              ),
      ),
    );
  }
}
