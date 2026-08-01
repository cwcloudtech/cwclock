plugins {
    id("com.android.application")
    // The Flutter Gradle Plugin must be applied after the Android and Kotlin Gradle plugins.
    id("dev.flutter.flutter-gradle-plugin")
}

android {
    namespace = "com.cwclock.mobile"
    compileSdk = flutter.compileSdkVersion
    ndkVersion = flutter.ndkVersion

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    defaultConfig {
        // TODO: Specify your own unique Application ID (https://developer.android.com/studio/build/application-id.html).
        applicationId = "com.cwclock.mobile"
        // You can update the following values to match your application needs.
        // For more information, see: https://flutter.dev/to/review-gradle-config.
        minSdk = flutter.minSdkVersion
        targetSdk = flutter.targetSdkVersion
        versionCode = flutter.versionCode
        versionName = flutter.versionName
    }

    // Every release APK must be signed with the same key, or Android refuses
    // to install one build as an "update" over another (ai-instruct-128 -
    // self-updating from ai-instruct-127 was failing with "App not
    // installed" because CI containers are fresh per build, so Gradle's
    // auto-generated debug key differed between the currently-installed APK
    // and the newly downloaded one). MOBILE_KEYSTORE_PATH mirrors the
    // pre-Flutter React Native build.gradle's env-var-driven signing config.
    signingConfigs {
        create("release") {
            val keystorePath = System.getenv("MOBILE_KEYSTORE_PATH")
            if (!keystorePath.isNullOrEmpty()) {
                storeFile = file(keystorePath)
                storePassword = System.getenv("MOBILE_KEYSTORE_PASSWORD")
                keyAlias = System.getenv("MOBILE_KEY_ALIAS")
                keyPassword = System.getenv("MOBILE_KEY_PASSWORD")
            }
        }
    }

    buildTypes {
        release {
            // Falls back to the auto-generated debug keystore whenever
            // MOBILE_KEYSTORE_PATH isn't set, so `flutter run --release`
            // still works without secrets - but that fallback must never be
            // used for a build that gets published, since its key is
            // regenerated on every fresh machine/container.
            signingConfig = if (System.getenv("MOBILE_KEYSTORE_PATH").isNullOrEmpty())
                signingConfigs.getByName("debug")
            else
                signingConfigs.getByName("release")
        }
    }
}

kotlin {
    compilerOptions {
        jvmTarget = org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17
    }
}

flutter {
    source = "../.."
}
