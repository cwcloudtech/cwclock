import 'package:flutter/material.dart';

/// Always white regardless of theme - used as a foreground color on top of
/// a colored (primary/danger) background, in both light and dark mode.
const Color kWhite = Color(0xFFFFFFFF);

/// A small shared palette, not a full design system - matches cwclock-ui's
/// --cw-primary-500 (src/index.css) for basic brand consistency. A
/// [ThemeExtension] so every screen/widget reading a token via
/// `AppColors.of(context)` picks up the current light/dark theme
/// automatically (ai-instruct-119).
@immutable
class AppColors extends ThemeExtension<AppColors> {
  final Color primary;
  final Color primaryDark;
  final Color danger;
  final Color text;
  final Color textMuted;
  final Color border;
  final Color background;
  final Color backgroundMuted;

  const AppColors({
    required this.primary,
    required this.primaryDark,
    required this.danger,
    required this.text,
    required this.textMuted,
    required this.border,
    required this.background,
    required this.backgroundMuted,
  });

  static const light = AppColors(
    primary: Color(0xFF1CB9F7),
    primaryDark: Color(0xFF0F8FC2),
    danger: Color(0xFFE5484D),
    text: Color(0xFF1A1A1A),
    textMuted: Color(0xFF6B7280),
    border: Color(0xFFE2E8F0),
    background: Color(0xFFFFFFFF),
    backgroundMuted: Color(0xFFF5F7FA),
  );

  static const dark = AppColors(
    primary: Color(0xFF1CB9F7),
    primaryDark: Color(0xFF3FCBFF),
    danger: Color(0xFFFF6B6E),
    text: Color(0xFFF2F2F2),
    textMuted: Color(0xFF9CA3AF),
    border: Color(0xFF2D3748),
    background: Color(0xFF121212),
    backgroundMuted: Color(0xFF1E1E1E),
  );

  static AppColors of(BuildContext context) => Theme.of(context).extension<AppColors>()!;

  @override
  AppColors copyWith({
    Color? primary,
    Color? primaryDark,
    Color? danger,
    Color? text,
    Color? textMuted,
    Color? border,
    Color? background,
    Color? backgroundMuted,
  }) {
    return AppColors(
      primary: primary ?? this.primary,
      primaryDark: primaryDark ?? this.primaryDark,
      danger: danger ?? this.danger,
      text: text ?? this.text,
      textMuted: textMuted ?? this.textMuted,
      border: border ?? this.border,
      background: background ?? this.background,
      backgroundMuted: backgroundMuted ?? this.backgroundMuted,
    );
  }

  @override
  AppColors lerp(ThemeExtension<AppColors>? other, double t) {
    if (other is! AppColors) return this;
    return t < 0.5 ? this : other;
  }
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

ThemeData buildAppTheme(Brightness brightness) {
  final colors = brightness == Brightness.dark ? AppColors.dark : AppColors.light;
  return ThemeData(
    useMaterial3: true,
    brightness: brightness,
    scaffoldBackgroundColor: colors.background,
    extensions: [colors],
    colorScheme: ColorScheme.fromSeed(
      seedColor: colors.primary,
      brightness: brightness,
      primary: colors.primary,
      error: colors.danger,
    ),
    appBarTheme: AppBarTheme(
      backgroundColor: colors.background,
      foregroundColor: colors.text,
      elevation: 0,
    ),
    inputDecorationTheme: InputDecorationTheme(
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(AppRadius.value),
        borderSide: BorderSide(color: colors.border),
      ),
      filled: true,
      fillColor: colors.background,
    ),
    textTheme: TextTheme(
      bodyMedium: TextStyle(color: colors.text),
      bodySmall: TextStyle(color: colors.textMuted),
    ),
  );
}
