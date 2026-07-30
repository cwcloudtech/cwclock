import 'package:flutter/material.dart';

import '../theme.dart';

/// Ported from src/components/ErrorBanner.js.
class ErrorBanner extends StatelessWidget {
  final String? message;

  const ErrorBanner({super.key, this.message});

  @override
  Widget build(BuildContext context) {
    if (message == null || message!.isEmpty) return const SizedBox.shrink();
    return Padding(
      padding: EdgeInsets.only(bottom: AppSpacing.of(1.5)),
      child: Text(message!, style: TextStyle(color: AppColors.of(context).danger, fontSize: 14)),
    );
  }
}
