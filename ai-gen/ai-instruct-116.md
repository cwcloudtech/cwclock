# AI instruction 116

## Mobile CICD

Now I have this in the CICD pipeline:

```shell
#31 306.6 > Task :react-native-safe-area-context:compileReleaseKotlin FAILED
#31 306.6 e: file:///app/node_modules/react-native-safe-area-context/android/src/main/java/com/th3rdwave/safeareacontext/SafeAreaProviderManager.kt:14:27 Cannot infer type for this parameter. Please specify it explicitly.
#31 306.6 e: file:///app/node_modules/react-native-safe-area-context/android/src/main/java/com/th3rdwave/safeareacontext/SafeAreaProviderManager.kt:14:62 Argument type mismatch: actual type is 'com.th3rdwave.safeareacontext.SafeAreaProviderManager', but 'U!' was expected.
#31 306.6 e: file:///app/node_modules/react-native-safe-area-context/android/src/main/java/com/th3rdwave/safeareacontext/SafeAreaProviderManager.kt:16:16 Return type of 'getDelegate' is not a subtype of the return type of the overridden member 'fun getDelegate(): ViewManagerDelegate<SafeAreaProvider!>?' defined in 'com/th3rdwave/safeareacontext/SafeAreaProviderManager'.
#31 306.7 
#31 306.7 > Task :react-native-share:generateReleaseBuildConfig
#31 313.7 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenContainerViewManager.kt:20:71 Type argument is not within its bounds: should be subtype of 'it(BaseViewManagerInterface & com.facebook.react.viewmanagers.RNSScreenContainerManagerInterface<com.swmansion.rnscreens.ScreenContainer!>)!'.
#31 313.7 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenContainerViewManager.kt:20:99 Argument type mismatch: actual type is 'com.swmansion.rnscreens.ScreenContainerViewManager', but 'U!' was expected.
#31 313.7 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenContentWrapperManager.kt:14:110 Argument type mismatch: actual type is 'com.swmansion.rnscreens.ScreenContentWrapperManager', but 'U!' was expected.
#31 313.7 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenFooterManager.kt:14:94 Argument type mismatch: actual type is 'com.swmansion.rnscreens.ScreenFooterManager', but 'U!' was expected.
#31 313.7 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenStackHeaderConfigViewManager.kt:29:87 Type argument is not within its bounds: should be subtype of 'it(BaseViewManagerInterface & com.facebook.react.viewmanagers.RNSScreenStackHeaderConfigManagerInterface<com.swmansion.rnscreens.ScreenStackHeaderConfig!>)!'.
#31 313.7 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenStackHeaderConfigViewManager.kt:29:123 Argument type mismatch: actual type is 'com.swmansion.rnscreens.ScreenStackHeaderConfigViewManager', but 'U!' was expected.
#31 313.7 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenStackHeaderSubviewManager.kt:19:89 Type argument is not within its bounds: should be subtype of 'it(BaseViewManagerInterface & com.facebook.react.viewmanagers.RNSScreenStackHeaderSubviewManagerInterface<com.swmansion.rnscreens.ScreenStackHeaderSubview!>)!'.
#31 313.7 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenStackHeaderSubviewManager.kt:19:122 Argument type mismatch: actual type is 'com.swmansion.rnscreens.ScreenStackHeaderSubviewManager', but 'U!' was expected.
#31 313.7 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenStackViewManager.kt:21:63 Type argument is not within its bounds: should be subtype of 'it(BaseViewManagerInterface & com.facebook.react.viewmanagers.RNSScreenStackManagerInterface<com.swmansion.rnscreens.ScreenStack!>)!'.
#31 313.7 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenStackViewManager.kt:21:87 Argument type mismatch: actual type is 'com.swmansion.rnscreens.ScreenStackViewManager', but 'U!' was expected.
#31 313.7 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenViewManager.kt:34:53 Type argument is not within its bounds: should be subtype of 'it(BaseViewManagerInterface & com.facebook.react.viewmanagers.RNSScreenManagerInterface<com.swmansion.rnscreens.Screen!>)!'.
#31 313.7 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenViewManager.kt:34:72 Argument type mismatch: actual type is 'com.swmansion.rnscreens.ScreenViewManager', but 'U!' was expected.
#31 313.7 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/SearchBarManager.kt:27:63 Type argument is not within its bounds: should be subtype of 'it(BaseViewManagerInterface & com.facebook.react.viewmanagers.RNSSearchBarManagerInterface<com.swmansion.rnscreens.SearchBarView!>)!'.
#31 313.7 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/SearchBarManager.kt:27:81 Argument type mismatch: actual type is 'com.swmansion.rnscreens.SearchBarManager', but 'U!' was expected.
#31 313.7 
#31 313.7 > Task :react-native-screens:compileReleaseKotlin
#31 313.8 
#31 313.8 FAILURE: Build completed with 2 failures.
#31 313.8 
#31 313.8 1: Task failed with an exception.
#31 313.8 -----------
#31 313.8 * What went wrong:
#31 313.8 Execution failed for task ':react-native-safe-area-context:compileReleaseKotlin'.
#31 313.8 > A failure occurred while executing org.jetbrains.kotlin.compilerRunner.GradleCompilerRunnerWithWorkers$GradleKotlinCompilerWorkAction
#31 313.8    > Compilation error. See log for more details
#31 313.8 
#31 313.8 * Try:
#31 313.8 > Run with --stacktrace option to get the stack trace.
#31 313.8 > Run with --info or --debug option to get more log output.
#31 313.8 > Run with --scan to get full insights.
#31 313.8 > Get more help at https://help.gradle.org.
#31 313.8 ==============================================================================
#31 313.8 
#31 313.8 2: Task failed with an exception.
#31 313.8 -----------
#31 313.8 * What went wrong:
#31 313.8 Execution failed for task ':react-native-screens:compileReleaseKotlin'.
#31 313.8 > A failure occurred while executing org.jetbrains.kotlin.compilerRunner.GradleCompilerRunnerWithWorkers$GradleKotlinCompilerWorkAction
#31 313.8    > Compilation error. See log for more details
#31 313.8 
#31 313.8 * Try:
#31 313.8 > Run with --stacktrace option to get the stack trace.
#31 313.8 > Run with --info or --debug option to get more log output.
#31 313.8 > Run with --scan to get full insights.
#31 313.8 > Get more help at https://help.gradle.org.
#31 313.8 ==============================================================================
#31 313.8 
#31 313.8 BUILD FAILED in 5m 13s
#31 313.8 
#31 313.8 > Task :react-native-screens:compileReleaseKotlin FAILED
#31 313.8 156 actionable tasks: 156 executed
#31 ERROR: process "/bin/sh -c VERSION=\"$(cat VERSION)\" &&     sed -i \"s/versionName \\\"[^\\\"]*\\\"/versionName \\\"${VERSION}\\\"/\" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk \"/out/cwclock-v${VERSION}.apk\"" did not complete successfully: exit code: 1
------
 > [mobile-build 8/8] RUN VERSION="$(cat VERSION)" &&     sed -i "s/versionName "[^"]*"/versionName "${VERSION}"/" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk "/out/cwclock-v${VERSION}.apk":
313.8 > Run with --stacktrace option to get the stack trace.
313.8 > Run with --info or --debug option to get more log output.
313.8 > Run with --scan to get full insights.
313.8 > Get more help at https://help.gradle.org.
313.8 ==============================================================================
313.8 
313.8 BUILD FAILED in 5m 13s
313.8 
313.8 > Task :react-native-screens:compileReleaseKotlin FAILED
313.8 156 actionable tasks: 156 executed
------
Dockerfile:59
--------------------
```

Fix the docker build once and for all.
