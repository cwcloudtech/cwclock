# AI instruction 123

## Mobile CICD

Fix this build error:

```shell
#28 334.7 warning: [options] source value 8 is obsolete and will be removed in a future release
#28 334.7 warning: [options] target value 8 is obsolete and will be removed in a future release
#28 334.7 warning: [options] To suppress warnings about obsolete options, use -Xlint:-options.
#28 334.7 3 warnings----- End of the daemon log -----
#28 334.7 
#28 334.7 JVM crash log found: file:///app/android/hs_err_pid180.log
#28 334.8 
#28 334.8 FAILURE: Build failed with an exception.
#28 334.8 
#28 334.8 * What went wrong:
#28 334.8 Gradle build daemon disappeared unexpectedly (it may have been killed or may have crashed)
#28 334.8 
#28 334.8 * Try:
#28 334.8 > Run with --stacktrace option to get the stack trace.
#28 334.8 > Run with --info or --debug option to get more log output.
#28 334.8 > Run with --scan to generate a Build Scan (Powered by Develocity).
#28 334.8 > Get more help at https://help.gradle.org.
#28 334.8 Running Gradle task 'assembleRelease'...                          328.5s
#28 334.8 Gradle task assembleRelease failed with exit code 1
#28 ERROR: process "/bin/sh -c VERSION=\"$(cat VERSION)\" &&     ANDROID_VERSION_CODE=\"$(echo \"${VERSION}\" | awk -F. '{printf \"%d%02d%02d\", $1, $2, $3}')\" &&     sed -i \"s/^version: .*/version: ${VERSION}+${ANDROID_VERSION_CODE}/\" pubspec.yaml &&     flutter pub get &&     flutter build apk --release &&     mv /app/build/app/outputs/flutter-apk/app-release.apk \"/app/build/app/outputs/flutter-apk/cwclock-v${VERSION}.apk\"" did not complete successfully: exit code: 1
```

I think copying the [gradle.properties](../.docker/android/gradle.properties) to the right place in the Dockerfile will solve this issue.
