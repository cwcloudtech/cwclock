# AI instruction 110

## Mobile CICD

I want you to fix all the following warnings (coming from the CICD):

```shell
#28 5.591 npm warn deprecated sudo-prompt@9.2.1: Package no longer supported. Contact Support at https://www.npmjs.com/support for more info.
#28 5.764 npm warn deprecated rimraf@3.0.2: Rimraf versions prior to v4 are no longer supported
#28 7.522 npm warn deprecated inflight@1.0.6: This module is not supported, and leaks memory. Do not use it. Check out lru-cache if you want a good and tested way to coalesce async requests by a key value, which is much more comprehensive and powerful.
#28 7.880 npm warn deprecated glob@7.2.3: Old versions of glob are not supported, and contain widely publicized security vulnerabilities, which have been fixed in the current version. Please update. Support for old versions may be purchased (at exorbitant rates) by contacting i@izs.me
#28 9.775 npm warn deprecated @humanwhocodes/object-schema@2.0.3: Use @eslint/object-schema instead
#28 9.867 npm warn deprecated @humanwhocodes/config-array@0.13.0: Use @eslint/config-array instead
#28 10.71 npm warn deprecated @babel/plugin-proposal-class-properties@7.18.6: This proposal has been merged to the ECMAScript standard and thus this plugin is no longer maintained. Please use @babel/plugin-transform-class-properties instead.
#28 10.71 npm warn deprecated @babel/plugin-proposal-optional-chaining@7.21.0: This proposal has been merged to the ECMAScript standard and thus this plugin is no longer maintained. Please use @babel/plugin-transform-optional-chaining instead.
#28 10.83 npm warn deprecated @babel/plugin-proposal-nullish-coalescing-operator@7.18.6: This proposal has been merged to the ECMAScript standard and thus this plugin is no longer maintained. Please use @babel/plugin-transform-nullish-coalescing-operator instead.
#28 10.98 npm warn deprecated rimraf@2.6.3: Rimraf versions prior to v4 are no longer supported
#28 11.79 npm warn deprecated @react-navigation/routers@6.1.9: This version is no longer supported
#28 12.91 npm warn deprecated glob@10.5.0: Old versions of glob are not supported, and contain widely publicized security vulnerabilities, which have been fixed in the current version. Please update. Support for old versions may be purchased (at exorbitant rates) by contacting i@izs.me
#28 13.31 npm warn deprecated @react-navigation/bottom-tabs@6.6.1: This version is no longer supported
#28 13.75 npm warn deprecated @react-navigation/native@6.1.18: This version is no longer supported
#28 13.81 npm warn deprecated @react-navigation/elements@1.3.31: This version is no longer supported
#28 14.80 npm warn deprecated @react-navigation/core@6.4.17: This version is no longer supported
#28 15.34 npm warn deprecated eslint@8.57.1: This version is no longer supported. Please see https://eslint.org/version-support for other options.
#28 95.32 added 987 packages, and audited 988 packages in 2m
#28 95.32 
#28 95.32 182 packages are looking for funding
#28 95.32   run `npm fund` for details
#28 96.12 
#28 96.12 55 vulnerabilities (6 moderate, 49 high)
#28 96.12 
#28 96.12 To address issues that do not require attention, run:
#28 96.12   npm audit fix
#28 96.12 
#28 96.12 To address all issues possible (including breaking changes), run:
#28 96.12   npm audit fix --force
#28 96.12 
#28 96.12 Some issues need review, and may require choosing
#28 96.12 a different dependency.
#28 96.12 
#28 96.12 Run `npm audit` for details.
#28 96.12 npm notice
#28 96.12 npm notice New major version of npm available! 10.8.2 -> 12.0.2
#28 96.12 npm notice Changelog: https://github.com/npm/cli/releases/tag/v12.0.2
#28 96.12 npm notice To update run: npm install -g npm@12.0.2
#28 96.12 npm notice
```

Update also npm in the dockerfile.

Fix also this warning:

```shell
#31 309.0 /app/node_modules/@react-native-async-storage/async-storage/android/src/main/java/com/reactnativecommunity/asyncstorage/AsyncStorageModule.java:84: warning: [removal] onCatalystInstanceDestroy() in NativeModule has been deprecated and marked for removal
#31 309.0   public void onCatalystInstanceDestroy() {
#31 309.0               ^
```

And this error:

```shell
31 348.1 error node_modules/react-native-screens/src/fabric/ScreenStackHeaderConfigNativeComponent.ts: /app/node_modules/react-native-screens/src/fabric/ScreenStackHeaderConfigNativeComponent.ts: Unknown prop type for "onAttached": "undefined".
#31 348.3 
#31 348.3 > Task :app:createBundleReleaseJsAndAssets FAILED
#31 351.0 > Task :react-native-gesture-handler:verifyReleaseResources
#31 351.9 
#31 351.9 > Task :react-native-keychain:compileReleaseKotlin
#31 351.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:171:11 Variable 'instance' is never used
#31 351.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:290:40 Variable 'cipher' initializer is redundant
#31 351.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:308:53 Unnecessary safe call on a non-null receiver of type CipherStorage?
#31 351.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:574:33 Elvis operator (?:) always returns the left operand of non-nullable type String
#31 351.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:576:33 Elvis operator (?:) always returns the left operand of non-nullable type String
#31 351.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageBase.kt:320:42 'getter for isInsideSecureHardware: Boolean' is deprecated. Deprecated in Java
#31 351.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageBase.kt:449:59 Unnecessary non-null assertion (!!) on a non-null receiver of type Key
#31 351.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageBase.kt:544:55 Parameter 'output' is never used, could be renamed to _
#31 351.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageBase.kt:547:54 Parameter 'input' is never used, could be renamed to _
#31 351.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageKeystoreAesGcm.kt:191:44 'setUserAuthenticationValidityDurationSeconds(Int): KeyGenParameterSpec.Builder' is deprecated. Deprecated in Java
#31 351.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageKeystoreRsaEcb.kt:211:34 'setUserAuthenticationValidityDurationSeconds(Int): KeyGenParameterSpec.Builder' is deprecated. Deprecated in Java
#31 356.2 e: Supertypes of the following classes cannot be resolved. Please make sure you have the required dependencies in the classpath:
#31 356.2     class com.facebook.react.viewmanagers.RNGestureHandlerButtonManagerInterface, unresolved supertypes: ViewManagerWithGeneratedInterface
#31 356.2     class com.facebook.react.viewmanagers.RNGestureHandlerRootViewManagerInterface, unresolved supertypes: ViewManagerWithGeneratedInterface
#31 356.2 Adding -Xextended-compiler-checks argument might provide additional information.
#31 356.2 
#31 356.2 e: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:42:1 Cannot access 'ViewManagerWithGeneratedInterface' which is a supertype of 'com.swmansion.gesturehandler.react.RNGestureHandlerButtonViewManager'. Check your module classpath for missing or conflicting dependencies
#31 356.2 e: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:49:5 Cannot access 'ViewManagerWithGeneratedInterface' which is a supertype of 'com.swmansion.gesturehandler.react.RNGestureHandlerButtonViewManager'. Check your module classpath for missing or conflicting dependencies
#31 356.2 e: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:49:107 Cannot access 'ViewManagerWithGeneratedInterface' which is a supertype of 'com.swmansion.gesturehandler.react.RNGestureHandlerButtonViewManager'. Check your module classpath for missing or conflicting dependencies
#31 356.2 e: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:149:5 Cannot access 'ViewManagerWithGeneratedInterface' which is a supertype of 'com.swmansion.gesturehandler.react.RNGestureHandlerButtonViewManager'. Check your module classpath for missing or conflicting dependencies
#31 356.2 e: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:154:71 Cannot access 'ViewManagerWithGeneratedInterface' which is a supertype of 'com.swmansion.gesturehandler.react.RNGestureHandlerButtonViewManager'. Check your module classpath for missing or conflicting dependencies
#31 356.2 e: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:156:3 Class 'ButtonViewGroup' is not abstract and does not implement abstract member public abstract fun getPointerEvents(): PointerEvents! defined in com.facebook.react.uimanager.ReactPointerEventsView
#31 356.2 e: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:217:5 'pointerEvents' overrides nothing
#31 356.2 e: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerRootViewManager.kt:16:1 Cannot access 'ViewManagerWithGeneratedInterface' which is a supertype of 'com.swmansion.gesturehandler.react.RNGestureHandlerRootViewManager'. Check your module classpath for missing or conflicting dependencies
#31 356.2 e: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerRootViewManager.kt:23:5 Cannot access 'ViewManagerWithGeneratedInterface' which is a supertype of 'com.swmansion.gesturehandler.react.RNGestureHandlerRootViewManager'. Check your module classpath for missing or conflicting dependencies
#31 356.2 e: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerRootViewManager.kt:23:116 Cannot access 'ViewManagerWithGeneratedInterface' which is a supertype of 'com.swmansion.gesturehandler.react.RNGestureHandlerRootViewManager'. Check your module classpath for missing or conflicting dependencies
#31 356.2 e: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerRootViewManager.kt:26:79 Cannot access 'ViewManagerWithGeneratedInterface' which is a supertype of 'com.swmansion.gesturehandler.react.RNGestureHandlerRootViewManager'. Check your module classpath for missing or conflicting dependencies
#31 356.2 
#31 356.2 > Task :react-native-gesture-handler:compileReleaseKotlin FAILED
#31 356.3 
#31 356.3 FAILURE: Build completed with 2 failures.
#31 356.3 
#31 356.3 1: Task failed with an exception.
#31 356.3 -----------
#31 356.3 * What went wrong:
#31 356.3 Execution failed for task ':app:createBundleReleaseJsAndAssets'.
#31 356.3 > Process 'command 'node'' finished with non-zero exit value 1
#31 356.3 
#31 356.3 * Try:
#31 356.3 > Run with --stacktrace option to get the stack trace.
#31 356.3 > Run with --info or --debug option to get more log output.
#31 356.3 > Run with --scan to get full insights.
#31 356.3 > Get more help at https://help.gradle.org.
#31 356.3 ==============================================================================
#31 356.3 
#31 356.3 2: Task failed with an exception.
#31 356.3 -----------
#31 356.3 * What went wrong:
#31 356.3 Execution failed for task ':react-native-gesture-handler:compileReleaseKotlin'.
#31 356.3 > A failure occurred while executing org.jetbrains.kotlin.compilerRunner.GradleCompilerRunnerWithWorkers$GradleKotlinCompilerWorkAction
#31 356.3    > Compilation error. See log for more details
#31 356.3 
#31 356.3 * Try:
#31 356.3 > Run with --stacktrace option to get the stack trace.
#31 356.3 > Run with --info or --debug option to get more log output.
#31 356.3 > Run with --scan to get full insights.
#31 356.3 > Get more help at https://help.gradle.org.
#31 356.3 ==============================================================================
#31 356.3 
#31 356.3 BUILD FAILED in 5m 55s
#31 356.3 131 actionable tasks: 131 executed
#31 ERROR: process "/bin/sh -c VERSION=\"$(cat VERSION)\" &&     sed -i \"s/versionName \\\"[^\\\"]*\\\"/versionName \\\"${VERSION}\\\"/\" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk \"/out/cwclock-v${VERSION}.apk\"" did not complete successfully: exit code: 1
------
 > [mobile-build 8/8] RUN VERSION="$(cat VERSION)" &&     sed -i "s/versionName "[^"]*"/versionName "${VERSION}"/" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk "/out/cwclock-v${VERSION}.apk":
356.3 
356.3 * Try:
356.3 > Run with --stacktrace option to get the stack trace.
356.3 > Run with --info or --debug option to get more log output.
356.3 > Run with --scan to get full insights.
356.3 > Get more help at https://help.gradle.org.
356.3 ==============================================================================
356.3 
356.3 BUILD FAILED in 5m 55s
356.3 131 actionable tasks: 131 executed
------
Dockerfile:56
--------------------
```
