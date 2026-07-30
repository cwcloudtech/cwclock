# AI instruction 108

## Mobile CICD

I have this build error on my gitlab pipeline:

```shell
31 [mobile-build 8/8] RUN VERSION="$(cat VERSION)" &&     sed -i "s/versionName "[^"]*"/versionName "${VERSION}"/" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk "/out/cwclock-v${VERSION}.apk"
#31 1.034 
#31 1.036 Welcome to Gradle 8.10.2!
#31 1.036 
#31 1.036 Here are the highlights of this release:
#31 1.036  - Support for Java 23
#31 1.036  - Faster configuration cache
#31 1.036  - Better configuration cache reports
#31 1.036 
#31 1.036 For more details see https://docs.gradle.org/8.10.2/release-notes.html
#31 1.036 
#31 1.234 To honour the JVM settings for this build a single-use Daemon process will be forked. For more on this, please refer to https://docs.gradle.org/8.10.2/userguide/gradle_daemon.html#sec:disabling_the_daemon in the Gradle documentation.
#31 2.633 Daemon will be stopped at the end of the build 
#31 63.64 > Task :gradle-plugin:settings-plugin:checkKotlinGradlePluginConfigurationErrors
#31 63.64 > Task :gradle-plugin:shared:checkKotlinGradlePluginConfigurationErrors
#31 63.73 > Task :gradle-plugin:settings-plugin:pluginDescriptors
#31 63.73 > Task :gradle-plugin:settings-plugin:processResources
#31 67.43 > Task :gradle-plugin:shared:processResources NO-SOURCE
#31 75.73 > Task :gradle-plugin:shared:compileKotlin
#31 75.73 > Task :gradle-plugin:shared:compileJava NO-SOURCE
#31 75.73 > Task :gradle-plugin:shared:classes UP-TO-DATE
#31 75.73 > Task :gradle-plugin:shared:jar
#31 85.14 > Task :gradle-plugin:settings-plugin:compileKotlin
#31 85.14 > Task :gradle-plugin:settings-plugin:compileJava NO-SOURCE
#31 85.14 > Task :gradle-plugin:settings-plugin:classes
#31 85.14 > Task :gradle-plugin:settings-plugin:jar
#31 96.53 
#31 96.53 FAILURE: Build failed with an exception.
#31 96.53 
#31 96.53 * What went wrong:
#31 96.53 A problem occurred configuring root project 'CWClockMobile'.
#31 96.53 > Could not resolve all artifacts for configuration ':classpath'.
#31 96.53    > Could not find com.android.tools.build:gradle:.
#31 96.53      Required by:
#31 96.53          root project :
#31 96.53    > Could not find com.facebook.react:react-native-gradle-plugin:.
#31 96.53      Required by:
#31 96.53          root project :
#31 96.53    > Could not find org.jetbrains.kotlin:kotlin-gradle-plugin:.
#31 96.53      Required by:
#31 96.53          root project :
#31 96.53 
#31 96.53 * Try:
#31 96.53 > Run with --stacktrace option to get the stack trace.
#31 96.53 > Run with --info or --debug option to get more log output.
#31 96.53 > Run with --scan to get full insights.
#31 96.53 > Get more help at https://help.gradle.org.
#31 96.53 
#31 96.53 BUILD FAILED in 1m 36s
#31 96.54 8 actionable tasks: 8 executed
#31 ERROR: process "/bin/sh -c VERSION=\"$(cat VERSION)\" &&     sed -i \"s/versionName \\\"[^\\\"]*\\\"/versionName \\\"${VERSION}\\\"/\" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk \"/out/cwclock-v${VERSION}.apk\"" did not complete successfully: exit code: 1
------
 > [mobile-build 8/8] RUN VERSION="$(cat VERSION)" &&     sed -i "s/versionName "[^"]*"/versionName "${VERSION}"/" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk "/out/cwclock-v${VERSION}.apk":
96.53          root project :
96.53 
96.53 * Try:
96.53 > Run with --stacktrace option to get the stack trace.
96.53 > Run with --info or --debug option to get more log output.
96.53 > Run with --scan to get full insights.
96.53 > Get more help at https://help.gradle.org.
96.53 
96.53 BUILD FAILED in 1m 36s
96.54 8 actionable tasks: 8 executed
------
Dockerfile:56
--------------------
  55 |     COPY VERSION ./VERSION
  56 | >>> RUN VERSION="$(cat VERSION)" && \
  57 | >>>     sed -i "s/versionName \"[^\"]*\"/versionName \"${VERSION}\"/" android/app/build.gradle && \
  58 | >>>     cd android && gradle assembleRelease --no-daemon && \
  59 | >>>     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk "/out/cwclock-v${VERSION}.apk"
  60 |     
--------------------
ERROR: failed to solve: process "/bin/sh -c VERSION=\"$(cat VERSION)\" &&     sed -i \"s/versionName \\\"[^\\\"]*\\\"/versionName \\\"${VERSION}\\\"/\" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk \"/out/cwclock-v${VERSION}.apk\"" did not complete successfully: exit code: 1
Error response from daemon: No such image: [MASKED]/ui:1.14.2-mobile
The push refers to repository [[MASKED]/ui]
An image does not exist locally with the tag: [MASKED]/ui
Error response from daemon: No such image: [MASKED]/ui:1.14.2-mobile
The push refers to repository [[MASKED]/ui]
An image does not exist locally with the tag: [MASKED]/ui
The push refers to repository [[MASKED]/ui]
An image does not exist locally with the tag: [MASKED]/ui
Cleaning up project directory and file based variables
00:01
ERROR: Job failed: exit status 1
```

Find the cause and fix it.
