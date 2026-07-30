import 'package:flutter/material.dart';

import '../theme.dart';

/// The shared page wrapper: safe-area + keyboard-avoiding + a scrollable,
/// consistently-padded body. scroll=false opts out for screens that manage
/// their own scrolling (e.g. a ListView-based list screen). Ported from
/// src/components/Screen.js.
class AppScreen extends StatelessWidget {
  final Widget child;
  final bool scroll;

  const AppScreen({super.key, required this.child, this.scroll = true});

  @override
  Widget build(BuildContext context) {
    final body = scroll
        ? SingleChildScrollView(
            padding: EdgeInsets.all(AppSpacing.of(2)),
            child: child,
          )
        : child;

    return Scaffold(
      backgroundColor: AppColors.of(context).background,
      body: SafeArea(child: body),
    );
  }
}
