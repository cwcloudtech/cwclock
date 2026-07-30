import 'package:flutter/material.dart';

import '../theme.dart';

class SelectItem {
  final String value;
  final String label;

  const SelectItem(this.value, this.label);
}

/// Ported from src/components/SelectField.js (which wrapped
/// @react-native-picker/picker) - uses Flutter's built-in
/// DropdownButtonFormField, no extra package needed.
class AppSelectField extends StatelessWidget {
  final String? label;
  final String value;
  final ValueChanged<String> onChanged;
  final List<SelectItem> items;
  final String? placeholder;
  final bool enabled;

  const AppSelectField({
    super.key,
    this.label,
    required this.value,
    required this.onChanged,
    required this.items,
    this.placeholder,
    this.enabled = true,
  });

  @override
  Widget build(BuildContext context) {
    final currentValue = value.isEmpty || items.every((i) => i.value != value) ? null : value;
    return Padding(
      padding: EdgeInsets.only(bottom: AppSpacing.of(2)),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (label != null)
            Padding(
              padding: EdgeInsets.only(bottom: AppSpacing.of(0.5)),
              child: Text(label!, style: TextStyle(fontSize: 13, color: AppColors.of(context).textMuted)),
            ),
          DropdownButtonFormField<String>(
            initialValue: currentValue,
            isExpanded: true,
            decoration: InputDecoration(
              contentPadding: EdgeInsets.symmetric(
                horizontal: AppSpacing.of(1.5),
                vertical: AppSpacing.of(1),
              ),
            ),
            hint: placeholder != null ? Text(placeholder!) : null,
            items: [
              for (final item in items)
                DropdownMenuItem(value: item.value, child: Text(item.label, overflow: TextOverflow.ellipsis)),
            ],
            onChanged: enabled ? (v) => onChanged(v ?? '') : null,
          ),
        ],
      ),
    );
  }
}
