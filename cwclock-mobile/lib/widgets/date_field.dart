import 'package:flutter/material.dart';

import '../theme.dart';

enum DateFieldMode { date, time }

/// Shows a formatted date/time value; tapping it opens the native
/// picker dialog. Ported from src/components/DateField.js
/// (@react-native-community/datetimepicker) - Flutter's built-in
/// showDatePicker/showTimePicker cover this natively, no package needed.
class AppDateField extends StatelessWidget {
  final String? label;
  final DateTime value;
  final ValueChanged<DateTime> onChanged;
  final DateFieldMode mode;

  const AppDateField({
    super.key,
    this.label,
    required this.value,
    required this.onChanged,
    this.mode = DateFieldMode.date,
  });

  String get _formattedValue {
    if (mode == DateFieldMode.time) {
      final h = value.hour.toString().padLeft(2, '0');
      final m = value.minute.toString().padLeft(2, '0');
      return '$h:$m';
    }
    final y = value.year.toString().padLeft(4, '0');
    final m = value.month.toString().padLeft(2, '0');
    final d = value.day.toString().padLeft(2, '0');
    return '$y-$m-$d';
  }

  Future<void> _handleTap(BuildContext context) async {
    if (mode == DateFieldMode.time) {
      final picked = await showTimePicker(
        context: context,
        initialTime: TimeOfDay.fromDateTime(value),
      );
      if (picked != null) {
        onChanged(DateTime(value.year, value.month, value.day, picked.hour, picked.minute));
      }
    } else {
      final picked = await showDatePicker(
        context: context,
        initialDate: value,
        firstDate: DateTime(2000),
        lastDate: DateTime(2100),
      );
      if (picked != null) {
        onChanged(DateTime(picked.year, picked.month, picked.day, value.hour, value.minute, value.second));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: () => _handleTap(context),
      child: Container(
        margin: EdgeInsets.only(bottom: AppSpacing.of(2)),
        padding: EdgeInsets.symmetric(horizontal: AppSpacing.of(1.5), vertical: AppSpacing.of(1.25)),
        decoration: BoxDecoration(
          border: Border.all(color: AppColors.of(context).border),
          borderRadius: BorderRadius.circular(AppRadius.value),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (label != null)
              Padding(
                padding: EdgeInsets.only(bottom: AppSpacing.of(0.5)),
                child: Text(label!, style: TextStyle(fontSize: 13, color: AppColors.of(context).textMuted)),
              ),
            Text(_formattedValue, style: TextStyle(fontSize: 16, color: AppColors.of(context).text)),
          ],
        ),
      ),
    );
  }
}
