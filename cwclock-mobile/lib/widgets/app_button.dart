import 'package:flutter/material.dart';

import '../theme.dart';

enum AppButtonVariant { primary, secondary, danger }

/// Ported from src/components/Button.js.
class AppButton extends StatelessWidget {
  final String title;
  final VoidCallback? onPressed;
  final AppButtonVariant variant;
  final bool loading;
  final EdgeInsetsGeometry? margin;

  const AppButton({
    super.key,
    required this.title,
    required this.onPressed,
    this.variant = AppButtonVariant.primary,
    this.loading = false,
    this.margin,
  });

  @override
  Widget build(BuildContext context) {
    final Color background;
    final Color foreground;
    switch (variant) {
      case AppButtonVariant.primary:
        background = AppColors.primary;
        foreground = AppColors.white;
        break;
      case AppButtonVariant.secondary:
        background = AppColors.backgroundMuted;
        foreground = AppColors.text;
        break;
      case AppButtonVariant.danger:
        background = AppColors.danger;
        foreground = AppColors.white;
        break;
    }

    final disabled = onPressed == null || loading;

    return Container(
      margin: margin,
      width: double.infinity,
      child: ElevatedButton(
        onPressed: disabled ? null : onPressed,
        style: ElevatedButton.styleFrom(
          backgroundColor: background,
          disabledBackgroundColor: background.withValues(alpha: 0.5),
          foregroundColor: foreground,
          padding: EdgeInsets.symmetric(vertical: AppSpacing.of(1.5)),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(AppRadius.value)),
          elevation: 0,
        ),
        child: loading
            ? SizedBox(
                height: 20,
                width: 20,
                child: CircularProgressIndicator(strokeWidth: 2, color: foreground),
              )
            : Text(title, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
      ),
    );
  }
}
