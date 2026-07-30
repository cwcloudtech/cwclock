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
  final IconData? icon;

  const AppButton({
    super.key,
    required this.title,
    required this.onPressed,
    this.variant = AppButtonVariant.primary,
    this.loading = false,
    this.margin,
    this.icon,
  });

  @override
  Widget build(BuildContext context) {
    final Color background;
    final Color foreground;
    switch (variant) {
      case AppButtonVariant.primary:
        background = AppColors.of(context).primary;
        foreground = kWhite;
        break;
      case AppButtonVariant.secondary:
        background = AppColors.of(context).backgroundMuted;
        foreground = AppColors.of(context).text;
        break;
      case AppButtonVariant.danger:
        background = AppColors.of(context).danger;
        foreground = kWhite;
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
            : icon == null
                ? Text(title, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600))
                : Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(icon, size: 18),
                      const SizedBox(width: 8),
                      Text(title, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                    ],
                  ),
      ),
    );
  }
}
