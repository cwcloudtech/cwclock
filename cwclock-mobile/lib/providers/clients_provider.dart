import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/client.dart';
import 'api_providers.dart';

class ClientsState {
  final List<Client> items;

  const ClientsState({this.items = const []});

  ClientsState copyWith({List<Client>? items}) => ClientsState(items: items ?? this.items);
}

/// Ported from src/redux/clients/clients.actions.js + clients.reducer.js.
class ClientsNotifier extends Notifier<ClientsState> {
  @override
  ClientsState build() => const ClientsState();

  Future<List<Client>> listClients(String orgId) async {
    final response = await ref.read(apiClientProvider).dio.get('/organizations/$orgId/clients/');
    final items = (response.data as List).map((e) => Client.fromJson(e as Map<String, dynamic>)).toList();
    state = state.copyWith(items: items);
    return items;
  }

  Future<Client> createClient(String orgId, Map<String, dynamic> fields) async {
    final response = await ref.read(apiClientProvider).dio.post('/organizations/$orgId/clients/', data: fields);
    final client = Client.fromJson(response.data as Map<String, dynamic>);
    state = state.copyWith(items: [...state.items, client]);
    return client;
  }

  Future<Client> updateClient(String orgId, String clientId, Map<String, dynamic> fields) async {
    final response =
        await ref.read(apiClientProvider).dio.put('/organizations/$orgId/clients/$clientId', data: fields);
    final updated = Client.fromJson(response.data as Map<String, dynamic>);
    state = state.copyWith(items: [for (final c in state.items) c.id == updated.id ? updated : c]);
    return updated;
  }

  Future<void> deleteClient(String orgId, String clientId) async {
    await ref.read(apiClientProvider).dio.delete('/organizations/$orgId/clients/$clientId');
    state = state.copyWith(items: state.items.where((c) => c.id != clientId).toList());
  }
}

final clientsProvider = NotifierProvider<ClientsNotifier, ClientsState>(ClientsNotifier.new);
