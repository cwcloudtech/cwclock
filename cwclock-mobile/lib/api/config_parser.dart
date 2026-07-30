import '../models/session_config.dart';

/// Parses the "key = value" line format shared by the CLI's ~/.cwclock/config
/// file and the QR code / config file the web app's API Keys screen
/// generates (`api_url = %s\napi_key = %s\n`, plus an optional `org_id = %s`)
/// - ported from src/api/config.js's parseConfigText. When a key appears more
/// than once, the last matching line wins.
SessionConfig parseConfigText(String raw) {
  final lines = raw.split('\n');

  String get(String key) {
    String? match;
    for (final line in lines) {
      if (line.startsWith('$key =')) match = line;
    }
    if (match == null) return '';
    final parts = match.split(' = ');
    return parts.length > 1 ? parts[1].trim() : '';
  }

  return SessionConfig(
    apiUrl: get('api_url'),
    apiKey: get('api_key'),
    orgId: get('org_id'),
  );
}
