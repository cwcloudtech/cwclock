# AI instruction 117

## Mobile CICD

Fix this build error:

```shell
#31 318.7 > Task :react-native-safe-area-context:compileReleaseKotlin FAILED
#31 318.7 e: file:///app/node_modules/react-native-safe-area-context/android/src/main/java/com/th3rdwave/safeareacontext/SafeAreaProviderManager.kt:14:27 Cannot infer type for this parameter. Please specify it explicitly.
#31 318.7 e: file:///app/node_modules/react-native-safe-area-context/android/src/main/java/com/th3rdwave/safeareacontext/SafeAreaProviderManager.kt:14:62 Argument type mismatch: actual type is 'com.th3rdwave.safeareacontext.SafeAreaProviderManager', but 'U!' was expected.
#31 318.7 e: file:///app/node_modules/react-native-safe-area-context/android/src/main/java/com/th3rdwave/safeareacontext/SafeAreaProviderManager.kt:16:16 Return type of 'getDelegate' is not a subtype of the return type of the overridden member 'fun getDelegate(): ViewManagerDelegate<SafeAreaProvider!>?' defined in 'com/th3rdwave/safeareacontext/SafeAreaProviderManager'.
#31 319.9 
#31 319.9 > Task :react-native-pdf:compileReleaseJavaWithJavac
#31 325.5 
#31 325.5 > Task :react-native-screens:compileReleaseKotlin
#31 325.5 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenContainerViewManager.kt:20:71 Type argument is not within its bounds: should be subtype of 'it(BaseViewManagerInterface & com.facebook.react.viewmanagers.RNSScreenContainerManagerInterface<com.swmansion.rnscreens.ScreenContainer!>)!'.
#31 325.5 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenContainerViewManager.kt:20:99 Argument type mismatch: actual type is 'com.swmansion.rnscreens.ScreenContainerViewManager', but 'U!' was expected.
#31 325.5 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenContentWrapperManager.kt:14:110 Argument type mismatch: actual type is 'com.swmansion.rnscreens.ScreenContentWrapperManager', but 'U!' was expected.
#31 325.5 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenFooterManager.kt:14:94 Argument type mismatch: actual type is 'com.swmansion.rnscreens.ScreenFooterManager', but 'U!' was expected.
#31 325.5 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenStackHeaderConfigViewManager.kt:29:87 Type argument is not within its bounds: should be subtype of 'it(BaseViewManagerInterface & com.facebook.react.viewmanagers.RNSScreenStackHeaderConfigManagerInterface<com.swmansion.rnscreens.ScreenStackHeaderConfig!>)!'.
#31 325.5 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenStackHeaderConfigViewManager.kt:29:123 Argument type mismatch: actual type is 'com.swmansion.rnscreens.ScreenStackHeaderConfigViewManager', but 'U!' was expected.
#31 325.5 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenStackHeaderSubviewManager.kt:19:89 Type argument is not within its bounds: should be subtype of 'it(BaseViewManagerInterface & com.facebook.react.viewmanagers.RNSScreenStackHeaderSubviewManagerInterface<com.swmansion.rnscreens.ScreenStackHeaderSubview!>)!'.
#31 325.5 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenStackHeaderSubviewManager.kt:19:122 Argument type mismatch: actual type is 'com.swmansion.rnscreens.ScreenStackHeaderSubviewManager', but 'U!' was expected.
#31 325.5 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenStackViewManager.kt:21:63 Type argument is not within its bounds: should be subtype of 'it(BaseViewManagerInterface & com.facebook.react.viewmanagers.RNSScreenStackManagerInterface<com.swmansion.rnscreens.ScreenStack!>)!'.
#31 325.5 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenStackViewManager.kt:21:87 Argument type mismatch: actual type is 'com.swmansion.rnscreens.ScreenStackViewManager', but 'U!' was expected.
#31 325.5 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenViewManager.kt:34:53 Type argument is not within its bounds: should be subtype of 'it(BaseViewManagerInterface & com.facebook.react.viewmanagers.RNSScreenManagerInterface<com.swmansion.rnscreens.Screen!>)!'.
#31 325.5 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/ScreenViewManager.kt:34:72 Argument type mismatch: actual type is 'com.swmansion.rnscreens.ScreenViewManager', but 'U!' was expected.
#31 325.5 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/SearchBarManager.kt:27:63 Type argument is not within its bounds: should be subtype of 'it(BaseViewManagerInterface & com.facebook.react.viewmanagers.RNSSearchBarManagerInterface<com.swmansion.rnscreens.SearchBarView!>)!'.
#31 325.5 e: file:///app/node_modules/react-native-screens/android/src/main/java/com/swmansion/rnscreens/SearchBarManager.kt:27:81 Argument type mismatch: actual type is 'com.swmansion.rnscreens.SearchBarManager', but 'U!' was expected.
#31 325.6 
#31 325.6 > Task :react-native-screens:compileReleaseKotlin FAILED
#31 325.6 
#31 325.6 FAILURE: Build completed with 2 failures.
#31 325.6 
#31 325.6 1: Task failed with an exception.
#31 325.6 -----------
#31 325.6 * What went wrong:
#31 325.6 Execution failed for task ':react-native-safe-area-context:compileReleaseKotlin'.
#31 325.6 > A failure occurred while executing org.jetbrains.kotlin.compilerRunner.GradleCompilerRunnerWithWorkers$GradleKotlinCompilerWorkAction
#31 325.6    > Compilation error. See log for more details
#31 325.6 
#31 325.6 * Try:
#31 325.7 153 actionable tasks: 153 executed
#31 325.7 > Run with --stacktrace option to get the stack trace.
#31 325.7 > Run with --info or --debug option to get more log output.
#31 325.7 > Run with --scan to get full insights.
#31 325.7 > Get more help at https://help.gradle.org.
#31 325.7 ==============================================================================
#31 325.7 
#31 325.7 2: Task failed with an exception.
#31 325.7 -----------
#31 325.7 * What went wrong:
#31 325.7 Execution failed for task ':react-native-screens:compileReleaseKotlin'.
#31 325.7 > A failure occurred while executing org.jetbrains.kotlin.compilerRunner.GradleCompilerRunnerWithWorkers$GradleKotlinCompilerWorkAction
#31 325.7    > Compilation error. See log for more details
#31 325.7 
#31 325.7 * Try:
#31 325.7 > Run with --stacktrace option to get the stack trace.
#31 325.7 > Run with --info or --debug option to get more log output.
#31 325.7 > Run with --scan to get full insights.
#31 325.7 > Get more help at https://help.gradle.org.
#31 325.7 ==============================================================================
#31 325.7 
#31 325.7 BUILD FAILED in 5m 25s
```

Fix those warnings:

```shell
#31 296.7 > Task :react-native-camera-kit:compileReleaseKotlin
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/CKCamera.kt:32:8 'interface RCTEventEmitter : Any, JavaScriptModule' is deprecated. Deprecated in Java.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/CKCamera.kt:222:25 Condition is always 'false'.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/CKCamera.kt:307:18 'fun setTargetAspectRatio(p0: Int): Preview.Builder' is deprecated. Deprecated in Java.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/CKCamera.kt:317:14 'fun setTargetAspectRatio(p0: Int): ImageCapture.Builder' is deprecated. Deprecated in Java.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/CKCamera.kt:326:14 'fun setTargetAspectRatio(p0: Int): ImageAnalysis.Builder' is deprecated. Deprecated in Java.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/CKCameraManager.kt:10:8 'class MapBuilder : Any' is deprecated. Deprecated in Java.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/CKCameraManager.kt:55:16 'class MapBuilder : Any' is deprecated. Deprecated in Java.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/CKCameraManager.kt:56:48 'class MapBuilder : Any' is deprecated. Deprecated in Java.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/CKCameraManager.kt:57:39 'class MapBuilder : Any' is deprecated. Deprecated in Java.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/CKCameraManager.kt:58:43 'class MapBuilder : Any' is deprecated. Deprecated in Java.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/CKCameraManager.kt:59:35 'class MapBuilder : Any' is deprecated. Deprecated in Java.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/CKCameraManager.kt:60:36 'class MapBuilder : Any' is deprecated. Deprecated in Java.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/CKCameraManager.kt:61:51 'class MapBuilder : Any' is deprecated. Deprecated in Java.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/CKCameraManager.kt:62:52 'class MapBuilder : Any' is deprecated. Deprecated in Java.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/RNCameraKitPackage.kt:3:8 'class TurboReactPackage : BaseReactPackage' is deprecated. Use BaseReactPackage instead.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/RNCameraKitPackage.kt:13:28 'class TurboReactPackage : BaseReactPackage' is deprecated. Use BaseReactPackage instead.
#31 296.7 w: file:///app/node_modules/react-native-camera-kit/android/src/main/java/com/rncamerakit/barcode/BarcodeFrame.kt:61:14 'fun invalidate(p0: Rect!): Unit' is deprecated. Deprecated in Java.
```

And those:

```shell
#31 312.9 > Task :react-native-keychain:compileReleaseKotlin
#31 312.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageBase.kt:320:42 'val isInsideSecureHardware: Boolean' is deprecated. Deprecated in Java.
#31 312.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageKeystoreAesGcm.kt:191:44 'fun setUserAuthenticationValidityDurationSeconds(p0: Int): KeyGenParameterSpec.Builder' is deprecated. Deprecated in Java.
#31 312.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageKeystoreRsaEcb.kt:211:34 'fun setUserAuthenticationValidityDurationSeconds(p0: Int): KeyGenParameterSpec.Builder' is deprecated. Deprecated in Java.
```