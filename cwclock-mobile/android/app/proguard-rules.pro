# Add project specific ProGuard rules here.
# By default, the flags in this file are appended to flags specified
# in /usr/local/Cellar/android-sdk/24.3.3/tools/proguard/proguard-android.txt
# You can edit the include path and order by changing the proguardFiles
# directive in build.gradle.
#
# For more details, see
#   http://developer.android.com/guide/developing/tools/proguard.html

# react-native-reanimated / react-native-screens / camera-kit and friends
# ship their own consumer-rules.pro that gets merged in automatically -
# nothing project-specific is needed here yet (minification is off for now,
# see build.gradle's enableProguardInReleaseBuilds).
