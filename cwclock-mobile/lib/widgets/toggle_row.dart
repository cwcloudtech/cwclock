import 'package:flutter/material.dart';

import '../theme.dart';

/// A labeled Switch row - ported from src/components/ToggleRow.js. Used for
/// the All day / Half day toggles (EditRecordScreen), which uncheck each
/// other instead of disabling.
class AppToggleRow extends StatelessWidget {
  final String label;
  final bool value;
  final ValueChanged<bool> onChanged;
  final bool disabled;

  const AppToggleRow({
    super.key,
    required this.label,
    required this.value,
    required this.onChanged,
    this.disabled = false,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.symmetric(vertical: AppSpacing.of(1)),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            label,
            style: TextStyle(fontSize: 16, color: disabled ? AppColors.of(context).textMuted : AppColors.of(context).text),
          ),
          Switch(
            value: value,
            onChanged: disabled ? null : onChanged,
            activeTrackColor: AppColors.of(context).primary,
          ),
        ],
      ),
    );
  }
}
