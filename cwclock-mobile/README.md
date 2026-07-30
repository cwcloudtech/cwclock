# cwclock-mobile

Android-first React Native app (bare CLI, plain JavaScript) covering a subset
of CWClock: time recording (start/stop, all-day records, edit, delete),
simplified summary/detailed PDF report generation, and invoice
preview/generation. See `ai-gen/ai-instruct-105.md` for the original request
and `~/.claude/plans/fizzy-growing-backus.md` (or your own copy of that plan)
for the full implementation plan this was built from.

## Setup

```sh
npm install
npx react-native run-android
```

Onboarding needs a running cwclock-api instance and an API key generated
from the web app's API Keys page (Organization → API Keys → Create, then
"Download config" or "Show QR"). Scan that QR code, or paste its contents
into the manual-entry screen.

## Known gaps / follow-ups (deliberately out of scope for this pass)

- **iOS is not configured.** Only the `android/` native project exists.
  Adding iOS support means running `npx react-native` iOS scaffolding (or
  reusing an existing iOS shell) and wiring up an `ios` Dockerfile stage
  separately - not attempted here.
- **`android/gradle/wrapper/gradle-wrapper.jar` isn't checked in** - it's a
  binary file. Run `gradle wrapper` once with a system-installed Gradle
  (matching `gradle-wrapper.properties`' pinned version) to generate it for
  local `./gradlew` use. The Docker `mobile-build` stage doesn't need it - it
  invokes the image's own installed `gradle` directly.
- **Release signing** falls back to the Android Gradle Plugin's implicit
  debug keystore when `MOBILE_KEYSTORE_PATH`/`MOBILE_KEYSTORE_PASSWORD`/
  `MOBILE_KEY_ALIAS`/`MOBILE_KEY_PASSWORD` aren't set (see
  `android/app/build.gradle`), so `assembleRelease` always produces an
  installable APK. Provide real secrets before publishing anywhere.
- Not implemented (outside the ai-instruct-105 feature subset): API-key
  creation on-device, org/client/project/member management, invoice email
  sending, export jobs, calendar view.
