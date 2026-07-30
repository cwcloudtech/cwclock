import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/main_tab_provider.dart';
import '../providers/session_provider.dart';
import '../theme.dart';
import 'data_uri_image.dart';

/// A drop-in replacement for a plain `AppBar(title: ...)` that also shows
/// the connected user's avatar on the right, tapping it jumps straight to
/// the Settings tab (ai-instruct-121: "in the sidebar... the avatar of the
/// connected user on the right of settings button" - this app has a bottom
/// tab bar rather than a sidebar, so the avatar lives in each screen's top
/// bar instead, consistently across the whole connected app shell).
class AppTopBar extends ConsumerWidget implements PreferredSizeWidget {
  final String title;
  final List<Widget>? actions;

  const AppTopBar({super.key, required this.title, this.actions});

  @override
  Size get preferredSize => const Size.fromHeight(kToolbarHeight);

  void _goToSettings(BuildContext context, WidgetRef ref) {
    ref.read(mainTabProvider.notifier).state = MainTab.settings;
    context.go('/main');
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final user = ref.watch(sessionProvider).user;

    return AppBar(
      title: Text(title),
      actions: [
        ...?actions,
        if (user != null)
          Padding(
            padding: const EdgeInsets.only(right: 12, left: 4),
            child: GestureDetector(
              onTap: () => _goToSettings(context, ref),
              child: DataUriImage(
                picture: user.picture,
                pictureX: user.pictureX,
                pictureY: user.pictureY,
                size: 32,
                fallback: CircleAvatar(
                  radius: 16,
                  backgroundColor: AppColors.of(context).backgroundMuted,
                  child: Icon(Icons.person_outline, color: AppColors.of(context).textMuted, size: 18),
                ),
              ),
            ),
          ),
      ],
    );
  }
}
