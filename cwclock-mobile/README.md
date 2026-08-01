# CWClock mobile

Android Flutter app covering a subset of CWClock: time recording (start/stop,
all-day records, edit, delete), simplified summary/detailed PDF report
generation, and invoice preview/generation/history, plus organization/member/
client/project management.

## Setup

```sh
flutter pub get
flutter run
```

Onboarding needs a running `cwclock-api` instance and an API key generated
from the web app's API Keys page (Organization -> API Keys -> Create, then
"Download config" or "Show QR"). Scan that QR code, or paste its contents
into the manual-entry screen.

## Architecture

- **State management**: Riverpod (`lib/providers/`), one `Notifier` per
  concern - session, organizations (+members), clients, projects, time
  entries, invoices, countries, currencies, locale. `lib/providers/reports_provider.dart`
  is stateless (nothing to keep beyond the PDF-generating call itself).
- **Navigation**: go_router (`lib/router.dart`). The whole route tree's
  `redirect` is driven by the session status
  (`restoring|missing|needsOrg|connected`, see `lib/providers/session_provider.dart`)
  - the same pattern the original RN app used to switch its entire
    navigator tree. The bottom tab bar (`lib/screens/main_tabs_screen.dart`)
    is a plain in-widget `IndexedStack`, not a nested router shell, since
    pushed screens (edit record, PDF viewer, etc.) need to cover the whole
    screen including the tab bar - they're separate top-level routes.
- **API client**: `lib/api/api_client.dart` (dio, `X-Api-Key` header
  injected per request) + `lib/api/pdf_client.dart` (binary PDF
  downloads, since report/invoice PDFs are fetched as bytes rather than
  JSON).
- **i18n**: `lib/i18n/` - a minimal hand-rolled `en`/`fr` dictionary lookup
  (not Flutter's `intl`/ARB tooling), so server error `i18n_code`s can be
  resolved through the same dictionaries as UI strings, matching the
  original app's `apiErrorMessage`.
- **Secure storage**: the API key lives in `flutter_secure_storage`
  (Android Keystore-backed); `apiUrl`/`orgId`/locale/the in-progress timer
  are plain `shared_preferences`.

## Known gaps / follow-ups (deliberately out of scope for this pass)

- **iOS is not configured.** Only the `android/` platform folder exists.
- Not implemented (outside the original feature subset): API-key creation
  on-device, invoice email sending, export jobs, calendar view.

## Release signing (Android)

`android/app/build.gradle.kts`'s `release` build type signs with a
persistent keystore supplied via the `MOBILE_KEYSTORE_PATH` /
`MOBILE_KEYSTORE_PASSWORD` / `MOBILE_KEY_ALIAS` / `MOBILE_KEY_PASSWORD` env
vars, falling back to Flutter's implicit debug keystore when they're unset
(so plain `flutter run --release` still works with no setup). CI supplies
these as build secrets - see the `mobile-build` stage in the root
`Dockerfile` and `secrets:` block in `docker-compose-build.yml`, sourced
from the `MOBILE_KEYSTORE_BASE64` / `MOBILE_KEYSTORE_PASSWORD` /
`MOBILE_KEY_ALIAS` / `MOBILE_KEY_PASSWORD` GitLab CI/CD variables (masked,
protected).

All release APKs must be signed with the *same* key, or Android refuses to
install one as an "update" over another ("App not installed" -
`ai-instruct-128`). Never let a real release build fall back to the debug
keystore.
