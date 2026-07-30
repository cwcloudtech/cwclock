import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/country.dart';
import 'api_providers.dart';

/// GET /v1/countries is public (no auth, no org scope), returns
/// { countries: [{ iso, name, currency }] }. Ported from
/// src/redux/countries/countries.actions.js.
class CountriesNotifier extends Notifier<List<Country>> {
  @override
  List<Country> build() => const [];

  Future<void> listCountries() async {
    final response = await ref.read(apiClientProvider).dio.get('/countries');
    final data = response.data as Map<String, dynamic>;
    state = (data['countries'] as List? ?? []).map((e) => Country.fromJson(e as Map<String, dynamic>)).toList();
  }
}

final countriesProvider = NotifierProvider<CountriesNotifier, List<Country>>(CountriesNotifier.new);
