# AI instruction 113

## Mobile CICD

Now I have this in the CICD pipeline:

```shell
 294.7 > Task :app:compileReleaseJavaWithJavac FAILED
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:14: error: package com.reactnativecommunity.asyncstorage does not exist
#31 294.7 import com.reactnativecommunity.asyncstorage.AsyncStoragePackage;
#31 294.7                                             ^
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:16: error: package com.reactcommunity.rndatetimepicker does not exist
#31 294.7 import com.reactcommunity.rndatetimepicker.RNDateTimePickerPackage;
#31 294.7                                           ^
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:18: error: package com.reactnativecommunity.picker does not exist
#31 294.7 import com.reactnativecommunity.picker.RNCPickerPackage;
#31 294.7                                       ^
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:20: error: package com.ReactNativeBlobUtil does not exist
#31 294.7 import com.ReactNativeBlobUtil.ReactNativeBlobUtilPackage;
#31 294.7                               ^
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:22: error: package com.rncamerakit does not exist
#31 294.7 import com.rncamerakit.RNCameraKitPackage;
#31 294.7                       ^
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:24: error: package com.swmansion.gesturehandler does not exist
#31 294.7 import com.swmansion.gesturehandler.RNGestureHandlerPackage;
#31 294.7                                    ^
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:26: error: package com.oblador.keychain does not exist
#31 294.7 import com.oblador.keychain.KeychainPackage;
#31 294.7                            ^
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:28: error: package org.wonday.pdf does not exist
#31 294.7 import org.wonday.pdf.RNPDFPackage;
#31 294.7                      ^
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:30: error: package com.th3rdwave.safeareacontext does not exist
#31 294.7 import com.th3rdwave.safeareacontext.SafeAreaContextPackage;
#31 294.7                                     ^
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:32: error: package com.swmansion.rnscreens does not exist
#31 294.7 import com.swmansion.rnscreens.RNScreensPackage;
#31 294.7                               ^
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:34: error: package cl.json does not exist
#31 294.7 import cl.json.RNSharePackage;
#31 294.7               ^
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:80: error: cannot find symbol
#31 294.7       new AsyncStoragePackage(),
#31 294.7           ^
#31 294.7   symbol:   class AsyncStoragePackage
#31 294.7   location: class PackageList
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:81: error: cannot find symbol
#31 294.7       new RNDateTimePickerPackage(),
#31 294.7           ^
#31 294.7   symbol:   class RNDateTimePickerPackage
#31 294.7   location: class PackageList
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:82: error: cannot find symbol
#31 294.7       new RNCPickerPackage(),
#31 294.7           ^
#31 294.7   symbol:   class RNCPickerPackage
#31 294.7   location: class PackageList
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:83: error: cannot find symbol
#31 294.7       new ReactNativeBlobUtilPackage(),
#31 294.7           ^
#31 294.7   symbol:   class ReactNativeBlobUtilPackage
#31 294.7   location: class PackageList
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:84: error: cannot find symbol
#31 294.7       new RNCameraKitPackage(),
#31 294.7           ^
#31 294.7   symbol:   class RNCameraKitPackage
#31 294.7   location: class PackageList
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:85: error: cannot find symbol
#31 294.7       new RNGestureHandlerPackage(),
#31 294.7           ^
#31 294.7   symbol:   class RNGestureHandlerPackage
#31 294.7   location: class PackageList
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:86: error: cannot find symbol
#31 294.7       new KeychainPackage(),
#31 294.7           ^
#31 294.7   symbol:   class KeychainPackage
#31 294.7   location: class PackageList
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:87: error: cannot find symbol
#31 294.7       new RNPDFPackage(),
#31 294.7           ^
#31 294.7   symbol:   class RNPDFPackage
#31 294.7   location: class PackageList
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:88: error: cannot find symbol
#31 294.7       new SafeAreaContextPackage(),
#31 294.7           ^
#31 294.7   symbol:   class SafeAreaContextPackage
#31 294.7   location: class PackageList
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:89: error: cannot find symbol
#31 294.7       new RNScreensPackage(),
#31 294.7           ^
#31 294.7   symbol:   class RNScreensPackage
#31 294.7   location: class PackageList
#31 294.7 /app/android/app/build/generated/autolinking/src/main/java/com/facebook/react/PackageList.java:90: error: cannot find symbol
#31 294.7       new RNSharePackage()
#31 294.7           ^
#31 294.7   symbol:   class RNSharePackage
#31 294.7   location: class PackageList
#31 294.7 22 errors
#31 295.2 
#31 295.2 > Task :react-native-gesture-handler:compileReleaseKotlin
#31 295.2 w: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/core/FlingGestureHandler.kt:25:26 Parameter 'event' is never used
#31 295.2 w: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:79:62 The corresponding parameter in the supertype 'ViewGroupManager' is named 'borderRadius'. This may cause problems when calling this function with named arguments.
#31 295.2 w: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:84:63 The corresponding parameter in the supertype 'ViewGroupManager' is named 'borderRadius'. This may cause problems when calling this function with named arguments.
#31 295.2 w: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:89:65 The corresponding parameter in the supertype 'ViewGroupManager' is named 'borderRadius'. This may cause problems when calling this function with named arguments.
#31 295.2 w: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:94:66 The corresponding parameter in the supertype 'ViewGroupManager' is named 'borderRadius'. This may cause problems when calling this function with named arguments.
#31 296.9 
#31 296.9 > Task :react-native-community_datetimepicker:mergeReleaseResources
#31 298.3 
#31 298.3 > Task :react-native-keychain:compileReleaseKotlin
#31 298.3 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:171:11 Variable 'instance' is never used
#31 298.3 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:290:40 Variable 'cipher' initializer is redundant
#31 298.3 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:308:53 Unnecessary safe call on a non-null receiver of type CipherStorage?
#31 298.3 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:574:33 Elvis operator (?:) always returns the left operand of non-nullable type String
#31 298.3 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:576:33 Elvis operator (?:) always returns the left operand of non-nullable type String
#31 298.3 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageBase.kt:320:42 'getter for isInsideSecureHardware: Boolean' is deprecated. Deprecated in Java
#31 298.3 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageBase.kt:449:59 Unnecessary non-null assertion (!!) on a non-null receiver of type Key
#31 298.3 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageBase.kt:544:55 Parameter 'output' is never used, could be renamed to _
#31 298.3 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageBase.kt:547:54 Parameter 'input' is never used, could be renamed to _
#31 298.3 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageKeystoreAesGcm.kt:191:44 'setUserAuthenticationValidityDurationSeconds(Int): KeyGenParameterSpec.Builder' is deprecated. Deprecated in Java
#31 298.3 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageKeystoreRsaEcb.kt:211:34 'setUserAuthenticationValidityDurationSeconds(Int): KeyGenParameterSpec.Builder' is deprecated. Deprecated in Java
#31 298.5 176 actionable tasks: 176 executed
#31 298.5 
#31 298.5 FAILURE: Build failed with an exception.
#31 298.5 
#31 298.5 * What went wrong:
#31 298.5 Execution failed for task ':app:compileReleaseJavaWithJavac'.
#31 298.5 > Compilation failed; see the compiler error output for details.
#31 298.5 
#31 298.5 * Try:
#31 298.5 > Run with --info option to get more log output.
#31 298.5 > Run with --scan to get full insights.
#31 298.5 
#31 298.5 BUILD FAILED in 4m 58s
#31 ERROR: process "/bin/sh -c VERSION=\"$(cat VERSION)\" &&     sed -i \"s/versionName \\\"[^\\\"]*\\\"/versionName \\\"${VERSION}\\\"/\" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk \"/out/cwclock-v${VERSION}.apk\"" did not complete successfully: exit code: 1
------
 > [mobile-build 8/8] RUN VERSION="$(cat VERSION)" &&     sed -i "s/versionName "[^"]*"/versionName "${VERSION}"/" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk "/out/cwclock-v${VERSION}.apk":
298.5 
298.5 * What went wrong:
298.5 Execution failed for task ':app:compileReleaseJavaWithJavac'.
298.5 > Compilation failed; see the compiler error output for details.
298.5 
298.5 * Try:
298.5 > Run with --info option to get more log output.
298.5 > Run with --scan to get full insights.
298.5 
298.5 BUILD FAILED in 4m 58s
------
Dockerfile:59
--------------------
  58 |     COPY VERSION ./VERSION
  59 | >>> RUN VERSION="$(cat VERSION)" && \
  60 | >>>     sed -i "s/versionName \"[^\"]*\"/versionName \"${VERSION}\"/" android/app/build.gradle && \
  61 | >>>     cd android && gradle assembleRelease --no-daemon && \
  62 | >>>     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk "/out/cwclock-v${VERSION}.apk"
  63 |     
--------------------
ERROR: failed to solve: process "/bin/sh -c VERSION=\"$(cat VERSION)\" &&     sed -i \"s/versionName \\\"[^\\\"]*\\\"/versionName \\\"${VERSION}\\\"/\" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk \"/out/cwclock-v${VERSION}.apk\"" did not complete successfully: exit code: 1
```

Fix the dockerbuild once and for all.
