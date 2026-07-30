import 'package:flutter/material.dart';

/// A small shared palette, not a full design system - matches cwclock-ui's
/// --cw-primary-500 (src/index.css) for basic brand consistency.
class AppColors {
  const AppColors._();

  static const primary = Color(0xFF1CB9F7);
  static const primaryDark = Color(0xFF0F8FC2);
  static const danger = Color(0xFFE5484D);
  static const text = Color(0xFF1A1A1A);
  static const textMuted = Color(0xFF6B7280);
  static const border = Color(0xFFE2E8F0);
  static const background = Color(0xFFFFFFFF);
  static const backgroundMuted = Color(0xFFF5F7FA);
  static const white = Color(0xFFFFFFFF);
}

class AppSpacing {
  const AppSpacing._();

  static double of(num n) => n * 8.0;
}

class AppRadius {
  const AppRadius._();

  static const value = 8.0;
}

/// Parses a "#rrggbb" hex string (project colors) into a [Color].
Color colorFromHex(String hex) {
  final cleaned = hex.replaceFirst('#', '');
  return Color(int.parse('FF$cleaned', radix: 16));
}

ThemeData buildAppTheme() {
  return ThemeData(
    useMaterial3: true,
    scaffoldBackgroundColor: AppColors.background,
    colorScheme: ColorScheme.fromSeed(
      seedColor: AppColors.primary,
      primary: AppColors.primary,
      error: AppColors.danger,
    ),
    appBarTheme: const AppBarTheme(
      backgroundColor: AppColors.background,
      foregroundColor: AppColors.text,
      elevation: 0,
    ),
    inputDecorationTheme: InputDecorationTheme(
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(AppRadius.value),
        borderSide: const BorderSide(color: AppColors.border),
      ),
      filled: true,
      fillColor: AppColors.background,
    ),
    textTheme: const TextTheme(
      bodyMedium: TextStyle(color: AppColors.text),
      bodySmall: TextStyle(color: AppColors.textMuted),
    ),
  );
}
