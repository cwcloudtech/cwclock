import 'package:flutter/material.dart';

import '../models/project.dart';
import '../theme.dart';

/// A fixed swatch row - the web app uses a free-form color input, which has
/// no simple Flutter equivalent without an extra dependency, so this mirrors
/// src/components/ColorSwatchPicker.js's fixed palette instead.
class ColorSwatchPicker extends StatelessWidget {
  final String? label;
  final String value;
  final ValueChanged<String> onChanged;

  const ColorSwatchPicker({super.key, this.label, required this.value, required this.onChanged});

  @override
  Widget build(BuildContext context) {
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
          Wrap(
            spacing: AppSpacing.of(1),
            runSpacing: AppSpacing.of(1),
            children: [
              for (final color in presetProjectColors)
                GestureDetector(
                  onTap: () => onChanged(color),
                  child: Container(
                    width: 32,
                    height: 32,
                    decoration: BoxDecoration(
                      color: colorFromHex(color),
                      shape: BoxShape.circle,
                      border: value == color ? Border.all(color: AppColors.of(context).text, width: 3) : null,
                    ),
                  ),
                ),
            ],
          ),
        ],
      ),
    );
  }
}
