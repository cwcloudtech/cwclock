import 'package:flutter/material.dart';

import '../theme.dart';

/// Ported from src/App.js's LoadingScreen - shown while session status is
/// "restoring".
class LoadingScreen extends StatelessWidget {
  const LoadingScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.of(context).background,
      body: Center(child: CircularProgressIndicator(color: AppColors.of(context).primary)),
    );
  }
}
