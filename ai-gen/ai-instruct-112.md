# AI instruction 112

## Mobile CICD

I want you to fix all the following errors (coming from the CICD Docker build):

```shell
#31 313.9 > Task :react-native-keychain:compileReleaseKotlin
#31 313.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:171:11 Variable 'instance' is never used
#31 313.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:290:40 Variable 'cipher' initializer is redundant
#31 313.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:308:53 Unnecessary safe call on a non-null receiver of type CipherStorage?
#31 313.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:574:33 Elvis operator (?:) always returns the left operand of non-nullable type String
#31 313.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/KeychainModule.kt:576:33 Elvis operator (?:) always returns the left operand of non-nullable type String
#31 313.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageBase.kt:320:42 'getter for isInsideSecureHardware: Boolean' is deprecated. Deprecated in Java
#31 313.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageBase.kt:449:59 Unnecessary non-null assertion (!!) on a non-null receiver of type Key
#31 313.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageBase.kt:544:55 Parameter 'output' is never used, could be renamed to _
#31 313.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageBase.kt:547:54 Parameter 'input' is never used, could be renamed to _
#31 313.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageKeystoreAesGcm.kt:191:44 'setUserAuthenticationValidityDurationSeconds(Int): KeyGenParameterSpec.Builder' is deprecated. Deprecated in Java
#31 313.9 w: file:///app/node_modules/react-native-keychain/android/src/main/java/com/oblador/keychain/cipherStorage/CipherStorageKeystoreRsaEcb.kt:211:34 'setUserAuthenticationValidityDurationSeconds(Int): KeyGenParameterSpec.Builder' is deprecated. Deprecated in Java
#31 314.0 
#31 314.0 > Task :app:processReleaseResources FAILED
#31 314.2 > Task :app:stripReleaseDebugSymbols
#31 314.3 > Task :react-native-keychain:verifyReleaseResources
#31 314.4 
#31 314.4 FAILURE: Build failed with an exception.
#31 314.4 
#31 314.4 * What went wrong:
#31 314.4 Execution failed for task ':app:processReleaseResources'.
#31 314.4 > A failure occurred while executing com.android.build.gradle.internal.res.LinkApplicationAndroidResourcesTask$TaskAction
#31 314.4    > Android resource linking failed
#31 314.4      com.cwclock.mobile.app-mergeReleaseResources-3:/values/values.xml:393: error: resource drawable/rn_edit_text_material (aka com.cwclock.mobile:drawable/rn_edit_text_material) not found.
#31 314.4      com.cwclock.mobile.app-mergeReleaseResources-3:/values/values.xml:393: error: resource drawable/rn_edit_text_material (aka com.cwclock.mobile:drawable/rn_edit_text_material) not found.
#31 314.4      error: failed linking references.
#31 314.4 
#31 314.4 
#31 314.4 * Try:
#31 314.4 > Run with --stacktrace option to get the stack trace.
#31 314.4 > Run with --info or --debug option to get more log output.
#31 314.4 > Run with --scan to get full insights.
#31 314.4 > Get more help at https://help.gradle.org.
#31 314.4 
#31 314.4 BUILD FAILED in 5m 13s
#31 314.4 203 actionable tasks: 203 executed
#31 ERROR: process "/bin/sh -c VERSION=\"$(cat VERSION)\" &&     sed -i \"s/versionName \\\"[^\\\"]*\\\"/versionName \\\"${VERSION}\\\"/\" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk \"/out/cwclock-v${VERSION}.apk\"" did not complete successfully: exit code: 1
------
 > [mobile-build 8/8] RUN VERSION="$(cat VERSION)" &&     sed -i "s/versionName "[^"]*"/versionName "${VERSION}"/" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk "/out/cwclock-v${VERSION}.apk":
314.4 
314.4 
314.4 * Try:
314.4 > Run with --stacktrace option to get the stack trace.
314.4 > Run with --info or --debug option to get more log output.
314.4 > Run with --scan to get full insights.
314.4 > Get more help at https://help.gradle.org.
314.4 
314.4 BUILD FAILED in 5m 13s
314.4 203 actionable tasks: 203 executed
```
