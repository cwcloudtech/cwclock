import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/ai.dart';
import 'api_providers.dart';

/// The AI assistant (ai-instruct-132): the user's subscriptions, the provider
/// catalogue, and their conversations. Mirrors cwclock-ui's Redux AI module.
class AiState {
  final List<AiProviderInfo> providers;
  final List<AiConnection> connections;
  final List<AiConversation> conversations;
  final bool isSending;

  const AiState({
    this.providers = const [],
    this.connections = const [],
    this.conversations = const [],
    this.isSending = false,
  });

  bool get enabled => connections.isNotEmpty;

  AiState copyWith({
    List<AiProviderInfo>? providers,
    List<AiConnection>? connections,
    List<AiConversation>? conversations,
    bool? isSending,
  }) {
    return AiState(
      providers: providers ?? this.providers,
      connections: connections ?? this.connections,
      conversations: conversations ?? this.conversations,
      isSending: isSending ?? this.isSending,
    );
  }
}

class AiNotifier extends Notifier<AiState> {
  @override
  AiState build() => const AiState();

  // --- providers + connections ---

  Future<void> listProviders() async {
    final response = await ref.read(apiClientProvider).dio.get('/ai/providers');
    final providers = (response.data as List)
        .map((e) => AiProviderInfo.fromJson(e as Map<String, dynamic>))
        .toList();
    state = state.copyWith(providers: providers);
  }

  Future<List<AiConnection>> listConnections() async {
    final response = await ref.read(apiClientProvider).dio.get('/ai/connections');
    final connections = (response.data as List)
        .map((e) => AiConnection.fromJson(e as Map<String, dynamic>))
        .toList();
    state = state.copyWith(connections: connections);
    return connections;
  }

  Future<void> createConnection(Map<String, dynamic> payload) async {
    await ref.read(apiClientProvider).dio.post('/ai/connections', data: payload);
    await listConnections();
  }

  Future<void> updateConnection(String id, Map<String, dynamic> payload) async {
    await ref.read(apiClientProvider).dio.put('/ai/connections/$id', data: payload);
    await listConnections();
  }

  Future<void> deleteConnection(String id) async {
    await ref.read(apiClientProvider).dio.delete('/ai/connections/$id');
    await listConnections();
  }

  /// Sets the order; the first is the default a question goes through.
  Future<void> setDefault(String id) async {
    await ref.read(apiClientProvider).dio.post('/ai/connections/order', data: {
      'connectionIds': [id],
    });
    await listConnections();
  }

  // --- sharing ---

  Future<List<AiConnectionShare>> listShares(String connectionId) async {
    final response = await ref.read(apiClientProvider).dio.get('/ai/connections/$connectionId/shares');
    return (response.data as List)
        .map((e) => AiConnectionShare.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<AiConnectionShare>> shareConnection(String connectionId, String email) async {
    final response = await ref.read(apiClientProvider).dio.post(
      '/ai/connections/$connectionId/shares',
      data: {'email': email, 'role': 'user'},
    );
    return (response.data as List)
        .map((e) => AiConnectionShare.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> unshareConnection(String connectionId, String userId) async {
    await ref.read(apiClientProvider).dio.delete('/ai/connections/$connectionId/shares/$userId');
  }

  // --- chat + conversations ---

  Future<void> listConversations() async {
    final response = await ref.read(apiClientProvider).dio.get('/ai/conversations');
    final conversations = (response.data as List)
        .map((e) => AiConversation.fromJson(e as Map<String, dynamic>))
        .toList();
    state = state.copyWith(conversations: conversations);
  }

  Future<AiConversation> getConversation(String id) async {
    final response = await ref.read(apiClientProvider).dio.get('/ai/conversations/$id');
    return AiConversation.fromJson(response.data as Map<String, dynamic>);
  }

  /// Asks the assistant a question. Returns the answer plus the (possibly new)
  /// conversation id and title.
  Future<Map<String, dynamic>> sendChat({
    required String message,
    String? conversationId,
    String? connectionId,
    String? orgId,
  }) async {
    state = state.copyWith(isSending: true);
    try {
      final response = await ref.read(apiClientProvider).dio.post('/ai/chat', data: {
        'message': message,
        if (conversationId != null && conversationId.isNotEmpty) 'conversationId': conversationId,
        if (connectionId != null && connectionId.isNotEmpty) 'connectionId': connectionId,
        if (orgId != null && orgId.isNotEmpty) 'orgId': orgId,
      });
      return response.data as Map<String, dynamic>;
    } finally {
      state = state.copyWith(isSending: false);
    }
  }

  Future<void> clearHistory() async {
    await ref.read(apiClientProvider).dio.delete('/ai/conversations', data: {'all': true});
    state = state.copyWith(conversations: const []);
  }
}

final aiProvider = NotifierProvider<AiNotifier, AiState>(AiNotifier.new);
