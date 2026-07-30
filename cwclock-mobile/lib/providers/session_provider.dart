import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/session_config.dart';
import '../models/user.dart';
import 'api_providers.dart';

/// restoring (booting up, checking storage) | missing (no/invalid session,
/// show onboarding) | needsOrg (valid key, no org picked yet) | connected
/// (ready for the main tab navigator). Ported from
/// src/redux/session/session.reducer.js.
enum SessionStatus { restoring, missing, needsOrg, connected }

class SessionState {
  final SessionStatus status;
  final String? apiUrl;
  final String? orgId;
  final User? user;

  const SessionState({
    this.status = SessionStatus.restoring,
    this.apiUrl,
    this.orgId,
    this.user,
  });

  SessionState copyWith({
    SessionStatus? status,
    String? apiUrl,
    String? orgId,
    User? user,
  }) {
    return SessionState(
      status: status ?? this.status,
      apiUrl: apiUrl ?? this.apiUrl,
      orgId: orgId ?? this.orgId,
      user: user ?? this.user,
    );
  }
}

/// Ported from src/redux/session/session.actions.js + src/storage/session.js
/// combined - the Redux/AsyncStorage split doesn't need to be mirrored 1:1
/// in Riverpod.
class SessionNotifier extends Notifier<SessionState> {
  @override
  SessionState build() => const SessionState();

  /// Runs once at app boot: if a session was saved from a previous run, wire
  /// the shared API client back up and confirm the API key is still accepted
  /// server-side - a revoked/deleted key must send the user back to
  /// onboarding, not into the tab navigator where every screen would just
  /// 401.
  Future<void> restoreSession() async {
    state = state.copyWith(status: SessionStatus.restoring);
    final client = ref.read(apiClientProvider);
    final secure = ref.read(secureStorageProvider);
    final local = ref.read(localStorageProvider);

    final apiKey = await secure.getApiKey();
    final apiUrl = await local.getApiUrl();
    if (apiKey == null || apiKey.isEmpty || apiUrl == null || apiUrl.isEmpty) {
      state = const SessionState(status: SessionStatus.missing);
      return;
    }

    final orgId = await local.getOrgId();
    client.setSession(apiUrl: apiUrl, apiKey: apiKey, orgId: orgId ?? '');
    try {
      final response = await client.dio.get('/users/me');
      final user = User.fromJson(response.data as Map<String, dynamic>);
      state = SessionState(
        status: SessionStatus.connected,
        apiUrl: apiUrl,
        orgId: orgId,
        user: user,
      );
    } catch (_) {
      client.clearSession();
      state = const SessionState(status: SessionStatus.missing);
    }
  }

  /// Validates {apiUrl, apiKey} against GET /users/me - the one endpoint
  /// that accepts X-Api-Key without also requiring an org membership check -
  /// then either stores orgId as given (from a QR/pasted config with an
  /// org_id line) or reports back that none was resolved, routing to the
  /// org-picker screen.
  Future<User> connect(SessionConfig config) async {
    final client = ref.read(apiClientProvider);
    final secure = ref.read(secureStorageProvider);
    final local = ref.read(localStorageProvider);

    client.setSession(apiUrl: config.apiUrl, apiKey: config.apiKey, orgId: config.orgId);
    final response = await client.dio.get('/users/me');
    final user = User.fromJson(response.data as Map<String, dynamic>);

    await secure.setApiKey(config.apiKey);
    await local.setApiUrl(config.apiUrl);

    if (config.orgId.isNotEmpty) {
      await local.setOrgId(config.orgId);
      state = SessionState(
        status: SessionStatus.connected,
        apiUrl: config.apiUrl,
        orgId: config.orgId,
        user: user,
      );
    } else {
      await local.clearOrgId();
      state = SessionState(status: SessionStatus.needsOrg, apiUrl: config.apiUrl, user: user);
    }
    return user;
  }

  Future<void> selectOrg(String orgId) async {
    final client = ref.read(apiClientProvider);
    final local = ref.read(localStorageProvider);
    await local.setOrgId(orgId);
    client.setSession(orgId: orgId);
    state = state.copyWith(status: SessionStatus.connected, orgId: orgId);
  }

  Future<void> disconnect() async {
    final client = ref.read(apiClientProvider);
    final secure = ref.read(secureStorageProvider);
    final local = ref.read(localStorageProvider);
    await secure.clearApiKey();
    await local.clearAll();
    client.clearSession();
    state = const SessionState(status: SessionStatus.missing);
  }
}

final sessionProvider = NotifierProvider<SessionNotifier, SessionState>(SessionNotifier.new);
