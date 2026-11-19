// The AI assistant (ai-instruct-132) wire types, mirroring cwclock-api's
// internal/models/ai.go and the web app's Redux shapes. Dates are ignored here
// (the UI doesn't show them); everything else is projected.

/// One entry of the provider catalogue (GET /v1/ai/providers).
class AiProviderInfo {
  final String name;
  final String label;
  final String defaultUrl;
  final String defaultModel;
  final List<String> models;
  final String keyHint;

  const AiProviderInfo({
    required this.name,
    required this.label,
    required this.defaultUrl,
    required this.defaultModel,
    this.models = const [],
    this.keyHint = '',
  });

  factory AiProviderInfo.fromJson(Map<String, dynamic> json) {
    return AiProviderInfo(
      name: json['name'] as String? ?? '',
      label: json['label'] as String? ?? '',
      defaultUrl: json['defaultUrl'] as String? ?? '',
      defaultModel: json['defaultModel'] as String? ?? '',
      models: (json['models'] as List?)?.map((e) => e.toString()).toList() ?? const [],
      keyHint: json['keyHint'] as String? ?? '',
    );
  }
}

/// One subscription the caller may spend (GET /v1/ai/connections).
class AiConnection {
  final String id;
  final String ownerId;
  final String name;
  final String provider;
  final String baseUrl;
  final String model;
  final int maxTokens;
  final bool hasToken;
  final String role; // owner | user

  const AiConnection({
    required this.id,
    required this.ownerId,
    required this.name,
    required this.provider,
    required this.baseUrl,
    required this.model,
    required this.maxTokens,
    required this.hasToken,
    required this.role,
  });

  bool get isOwner => role == 'owner';

  factory AiConnection.fromJson(Map<String, dynamic> json) {
    return AiConnection(
      id: json['id'] as String? ?? '',
      ownerId: json['ownerId'] as String? ?? '',
      name: json['name'] as String? ?? '',
      provider: json['provider'] as String? ?? '',
      baseUrl: json['baseUrl'] as String? ?? '',
      model: json['model'] as String? ?? '',
      maxTokens: (json['maxTokens'] as num?)?.toInt() ?? 0,
      hasToken: json['hasToken'] as bool? ?? false,
      role: json['role'] as String? ?? '',
    );
  }
}

/// One line of a subscription's access list (GET .../shares).
class AiConnectionShare {
  final String userId;
  final String email;
  final String name;
  final String surname;
  final String role;

  const AiConnectionShare({
    required this.userId,
    required this.email,
    required this.name,
    required this.surname,
    required this.role,
  });

  String get label {
    final full = '$name $surname'.trim();
    return full.isNotEmpty ? full : email;
  }

  factory AiConnectionShare.fromJson(Map<String, dynamic> json) {
    return AiConnectionShare(
      userId: json['userId'] as String? ?? '',
      email: json['email'] as String? ?? '',
      name: json['name'] as String? ?? '',
      surname: json['surname'] as String? ?? '',
      role: json['role'] as String? ?? '',
    );
  }
}

/// One stored exchange within a conversation.
class AiExchange {
  final String question;
  final String answer;

  const AiExchange({required this.question, required this.answer});

  factory AiExchange.fromJson(Map<String, dynamic> json) {
    return AiExchange(
      question: json['question'] as String? ?? '',
      answer: json['answer'] as String? ?? '',
    );
  }
}

/// A conversation thread (list form carries no exchanges; a read carries them).
class AiConversation {
  final String id;
  final String title;
  final int count;
  final List<AiExchange> exchanges;

  const AiConversation({
    required this.id,
    required this.title,
    required this.count,
    this.exchanges = const [],
  });

  factory AiConversation.fromJson(Map<String, dynamic> json) {
    return AiConversation(
      id: json['id'] as String? ?? '',
      title: json['title'] as String? ?? '',
      count: (json['count'] as num?)?.toInt() ?? 0,
      exchanges: (json['exchanges'] as List?)
              ?.map((e) => AiExchange.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
    );
  }
}
