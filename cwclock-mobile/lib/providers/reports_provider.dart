import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'api_providers.dart';

/// Stateless - ported from src/redux/reports/reports.actions.js, which has
/// no reducer (nothing to keep in state, just a PDF-generating function).
/// "Simplified" per the original instruction: just a date range and Summary
/// vs Detailed, no client/project/user filters like the web app's Reports
/// page.
class ReportsService {
  final Ref ref;

  ReportsService(this.ref);

  Future<String> generateReportPdf(String orgId, String reportType, String start, String end) {
    return ref.read(pdfClientProvider).fetchPdf(
      'POST',
      '/organizations/$orgId/reports/$reportType',
      {'exportType': 'PDF', 'dateRangeStart': start, 'dateRangeEnd': end},
    );
  }
}

final reportsServiceProvider = Provider<ReportsService>((ref) => ReportsService(ref));
