import 'package:flutter/material.dart';

import '../theme.dart';

/// Ported from src/components/FormField.js. Uses an internal
/// [TextEditingController] kept in sync with [value] (rather than
/// [TextFormField]'s `initialValue`, which only applies on first build) so
/// programmatic updates - e.g. ManualEntryScreen filling apiUrl/apiKey/orgId
/// from parsed pasted text - are reflected, matching the RN version's fully
/// controlled `<TextInput value=.../>`.
class AppFormField extends StatefulWidget {
  final String? label;
  final String value;
  final ValueChanged<String> onChanged;
  final String? placeholder;
  final bool obscureText;
  final bool enabled;
  final TextInputType? keyboardType;
  final int? maxLines;
  final TextCapitalization textCapitalization;
  final FocusNode? focusNode;

  const AppFormField({
    super.key,
    this.label,
    required this.value,
    required this.onChanged,
    this.placeholder,
    this.obscureText = false,
    this.enabled = true,
    this.keyboardType,
    this.maxLines = 1,
    this.textCapitalization = TextCapitalization.sentences,
    this.focusNode,
  });

  @override
  State<AppFormField> createState() => _AppFormFieldState();
}

class _AppFormFieldState extends State<AppFormField> {
  late final TextEditingController _controller = TextEditingController(text: widget.value);

  @override
  void didUpdateWidget(AppFormField oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.value != _controller.text) {
      _controller.value = _controller.value.copyWith(
        text: widget.value,
        selection: TextSelection.collapsed(offset: widget.value.length),
      );
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(bottom: AppSpacing.of(2)),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (widget.label != null)
            Padding(
              padding: EdgeInsets.only(bottom: AppSpacing.of(0.5)),
              child: Text(widget.label!, style: const TextStyle(fontSize: 13, color: AppColors.textMuted)),
            ),
          TextFormField(
            controller: _controller,
            focusNode: widget.focusNode,
            onChanged: widget.onChanged,
            obscureText: widget.obscureText,
            enabled: widget.enabled,
            keyboardType: widget.keyboardType,
            maxLines: widget.obscureText ? 1 : widget.maxLines,
            textCapitalization: widget.textCapitalization,
            style: const TextStyle(fontSize: 16, color: AppColors.text),
            decoration: InputDecoration(
              hintText: widget.placeholder,
              hintStyle: const TextStyle(color: AppColors.textMuted),
              contentPadding: EdgeInsets.symmetric(
                horizontal: AppSpacing.of(1.5),
                vertical: AppSpacing.of(1.25),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
