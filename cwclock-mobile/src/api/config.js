// parseConfigText parses the "key = value" line format shared by the CLI's
// ~/.cwclock/config file and the QR code / config file the web app's API
// Keys screen generates (POST /v1/users/me/config/qr and /file - see
// cwclock-api/internal/handlers/config_handler.go's buildConfigText: `api_url
// = %s\napi_key = %s\n`, plus an optional `org_id = %s`), so scanning that QR
// or pasting that file's contents here parses identically to the CLI's own
// config.GetValueFromFile (cwclock-cli/config/config.go). When a key appears
// more than once, the last matching line wins - same behavior as the Go
// implementation.
export const parseConfigText = (raw) => {
  const lines = String(raw ?? "").split("\n");

  const get = (key) => {
    let match = null;
    for (const line of lines) {
      if (line.startsWith(`${key} =`)) {
        match = line;
      }
    }
    if (match === null) return "";
    const value = match.split(" = ")[1];
    return value ? value.trim() : "";
  };

  return {
    apiUrl: get("api_url"),
    apiKey: get("api_key"),
    orgId: get("org_id"),
  };
};

// isCompleteConfig reports whether a parsed config has at least the two
// required fields (org_id is optional - onboarding falls back to picking an
// organization when it's missing, see screens/onboarding/OrgPickerScreen).
export const isCompleteConfig = (parsed) => Boolean(parsed?.apiUrl && parsed?.apiKey);
