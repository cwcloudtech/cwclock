import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/invoice.dart';
import 'api_providers.dart';

class InvoicesState {
  final List<Invoice> items;

  const InvoicesState({this.items = const []});

  InvoicesState copyWith({List<Invoice>? items}) => InvoicesState(items: items ?? this.items);
}

/// Ported from src/redux/invoices/invoices.actions.js + invoices.reducer.js.
class InvoicesNotifier extends Notifier<InvoicesState> {
  @override
  InvoicesState build() => const InvoicesState();

  /// One client + date range, all three required by the API.
  Future<List<Invoice>> listInvoices(String orgId, String clientId, String start, String end) async {
    final response = await ref.read(apiClientProvider).dio.get(
      '/organizations/$orgId/invoices/',
      queryParameters: {'clientId': clientId, 'start': start, 'end': end},
    );
    final items = (response.data as List).map((e) => Invoice.fromJson(e as Map<String, dynamic>)).toList();
    state = state.copyWith(items: items);
    return items;
  }

  /// Renders the PDF without saving anything - safe to call repeatedly while
  /// adjusting the client/date range before committing to Generate.
  Future<String> previewInvoicePdf(String orgId, String clientId, String start, String end) {
    return ref.read(pdfClientProvider).fetchPdf(
      'POST',
      '/organizations/$orgId/invoices/preview',
      {'clientId': clientId, 'dateRangeStart': start, 'dateRangeEnd': end},
    );
  }

  /// Allocates a real invoice number and saves it - callers should confirm
  /// with the user first, since unlike Preview this can't be undone from the
  /// app.
  Future<String> generateInvoicePdf(String orgId, String clientId, String start, String end) {
    return ref.read(pdfClientProvider).fetchPdf(
      'POST',
      '/organizations/$orgId/invoices',
      {'clientId': clientId, 'dateRangeStart': start, 'dateRangeEnd': end},
    );
  }

  /// Fetches an already-generated invoice's stored PDF.
  Future<String> downloadInvoicePdf(String orgId, String invoiceId) {
    return ref.read(pdfClientProvider).fetchPdf('GET', '/organizations/$orgId/invoices/$invoiceId/pdf');
  }
}

final invoicesProvider = NotifierProvider<InvoicesNotifier, InvoicesState>(InvoicesNotifier.new);
