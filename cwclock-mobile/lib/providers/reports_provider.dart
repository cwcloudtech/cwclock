import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'api_providers.dart';

/// Stateless - ported from src/redux/reports/reports.actions.js, which has
/// no reducer (nothing to keep in state, just a PDF-generating function).
/// clientId/projectId/memberId mirror the web app's clientIds/projectIds/
/// userIds MultiSelect filters (ai-instruct-127), single-valued here to
/// match the rest of the mobile app's simplified single-select pickers.
class ReportsService {
  final Ref ref;

  ReportsService(this.ref);

  Future<String> generateReportPdf(
    String orgId,
    String reportType,
    String start,
    String end, {
    String? clientId,
    String? projectId,
    String? memberId,
  }) {
    return ref.read(pdfClientProvider).fetchPdf(
      'POST',
      '/organizations/$orgId/reports/$reportType',
      {
        'exportType': 'PDF',
        'dateRangeStart': start,
        'dateRangeEnd': end,
        if (clientId != null && clientId.isNotEmpty) 'clients': {'ids': [clientId]},
        if (projectId != null && projectId.isNotEmpty) 'projects': {'ids': [projectId]},
        if (memberId != null && memberId.isNotEmpty) 'users': {'ids': [memberId]},
      },
    );
  }
}

final reportsServiceProvider = Provider<ReportsService>((ref) => ReportsService(ref));
