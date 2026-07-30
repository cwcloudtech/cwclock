import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../providers/locale_provider.dart';
import '../../theme.dart';
import '../../widgets/app_button.dart';
import '../../widgets/app_screen.dart';

/// The onboarding entry point: two ways into the same underlying flow, both
/// ending at SessionNotifier.connect. Neither button navigates directly past
/// onboarding itself - the router's redirect (see router.dart) reacts to the
/// resulting session status change. Ported from
/// src/screens/onboarding/ConnectScreen.js.
class ConnectScreen extends ConsumerWidget {
  const ConnectScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);

    return AppScreen(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(height: AppSpacing.of(6)),
          Text(
            t('onboarding.title'),
            style: const TextStyle(fontSize: 24, fontWeight: FontWeight.w700, color: AppColors.text),
          ),
          SizedBox(height: AppSpacing.of(1)),
          Text(t('onboarding.intro'), style: const TextStyle(fontSize: 15, color: AppColors.textMuted)),
          SizedBox(height: AppSpacing.of(4)),
          AppButton(
            title: t('onboarding.scanQr'),
            onPressed: () => context.push('/onboarding/scan-qr'),
            margin: EdgeInsets.only(bottom: AppSpacing.of(2)),
          ),
          AppButton(
            title: t('onboarding.enterManually'),
            variant: AppButtonVariant.secondary,
            onPressed: () => context.push('/onboarding/manual-entry'),
          ),
        ],
      ),
    );
  }
}
