import 'dart:typed_data';

import 'package:flutter/material.dart';

/// Renders an organization/user `picture` field - a self-contained
/// `data:image/...;base64,...` URI (cwclock-api never returns a bare
/// path/URL for these, see Organization.Picture/User.Picture), cropped the
/// same way cwclock-ui does via CSS `object-position` (Slidebar.jsx):
/// `pictureX`/`pictureY` are 0-100 percentages, 50/50 being centered.
/// Falls back to [fallback] (or a plain circle) when there's no picture or
/// it fails to decode.
class DataUriImage extends StatelessWidget {
  final String? picture;
  final double pictureX;
  final double pictureY;
  final double size;
  final double? borderRadius;
  final Widget? fallback;

  const DataUriImage({
    super.key,
    required this.picture,
    this.pictureX = 50,
    this.pictureY = 50,
    required this.size,
    this.borderRadius,
    this.fallback,
  });

  Uint8List? get _bytes {
    final value = picture;
    if (value == null || value.isEmpty) return null;
    try {
      return UriData.parse(value).contentAsBytes();
    } catch (_) {
      return null;
    }
  }

  @override
  Widget build(BuildContext context) {
    final bytes = _bytes;
    final radius = borderRadius ?? size / 2;

    if (bytes == null) {
      return fallback ?? SizedBox(width: size, height: size);
    }

    return ClipRRect(
      borderRadius: BorderRadius.circular(radius),
      child: Image.memory(
        bytes,
        width: size,
        height: size,
        fit: BoxFit.cover,
        alignment: Alignment((pictureX / 50) - 1, (pictureY / 50) - 1),
        errorBuilder: (context, error, stackTrace) => fallback ?? SizedBox(width: size, height: size),
      ),
    );
  }
}
