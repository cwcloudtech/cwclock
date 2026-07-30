/// The parsed `api_url = ...\napi_key = ...\norg_id = ...` config text - same
/// shape as the CLI's config file / the web app's QR-code config download.
class SessionConfig {
  final String apiUrl;
  final String apiKey;
  final String orgId;

  const SessionConfig({required this.apiUrl, required this.apiKey, this.orgId = ''});

  bool get isComplete => apiUrl.isNotEmpty && apiKey.isNotEmpty;
}
