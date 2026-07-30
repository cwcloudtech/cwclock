import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/currency.dart';
import 'api_providers.dart';

/// GET /v1/currencies is public (no auth, no org scope), returns
/// { currencies: [{ iso, name }] }. Ported from
/// src/redux/currencies/currencies.actions.js.
class CurrenciesNotifier extends Notifier<List<Currency>> {
  @override
  List<Currency> build() => const [];

  Future<void> listCurrencies() async {
    final response = await ref.read(apiClientProvider).dio.get('/currencies');
    final data = response.data as Map<String, dynamic>;
    state = (data['currencies'] as List? ?? []).map((e) => Currency.fromJson(e as Map<String, dynamic>)).toList();
  }
}

final currenciesProvider = NotifierProvider<CurrenciesNotifier, List<Currency>>(CurrenciesNotifier.new);
