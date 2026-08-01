import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:open_filex/open_filex.dart';
import 'package:package_info_plus/package_info_plus.dart';
import 'package:path_provider/path_provider.dart';

/// Always checks the official production manifest, independent of the
/// connected session's (possibly self-hosted) API URL - self-updates track
/// the cwclock.me release, not whatever backend the session happens to
/// point at (ai-instruct-127).
const _manifestUrl = 'https://api.cwclock.me/v1/manifest';

String _apkUrl(String version) => 'https://www.cwclock.me/cwclock-v$version.apk';

/// Compares dotted version strings segment by segment, numerically, with
/// missing segments treated as 0 - the manifest's version ("0.1") and the
/// app's own pubspec version ("1.0.0") don't necessarily share a segment
/// count. Returns >0 if [a] is newer than [b].
int compareVersions(String a, String b) {
  final partsA = a.split('.');
  final partsB = b.split('.');
  final length = partsA.length > partsB.length ? partsA.length : partsB.length;
  for (var i = 0; i < length; i++) {
    final na = i < partsA.length ? int.tryParse(partsA[i]) ?? 0 : 0;
    final nb = i < partsB.length ? int.tryParse(partsB[i]) ?? 0 : 0;
    if (na != nb) return na.compareTo(nb);
  }
  return 0;
}

class AppUpdateState {
  final String? availableVersion;
  final bool downloading;
  final double progress;

  const AppUpdateState({this.availableVersion, this.downloading = false, this.progress = 0});

  AppUpdateState copyWith({String? availableVersion, bool? downloading, double? progress}) {
    return AppUpdateState(
      availableVersion: availableVersion ?? this.availableVersion,
      downloading: downloading ?? this.downloading,
      progress: progress ?? this.progress,
    );
  }
}

/// Checks the production manifest for a newer app version and, on request,
/// downloads and opens its APK so Android's package installer can update
/// the app in place. Config (api url/key, org id) lives in secure/local
/// storage, which an in-place update leaves untouched (ai-instruct-127).
class AppUpdateNotifier extends Notifier<AppUpdateState> {
  @override
  AppUpdateState build() => const AppUpdateState();

  Future<void> checkForUpdate() async {
    try {
      final response = await Dio().get<Map<String, dynamic>>(_manifestUrl);
      final remoteVersion = response.data?['version'] as String?;
      if (remoteVersion == null) return;

      final packageInfo = await PackageInfo.fromPlatform();
      if (compareVersions(remoteVersion, packageInfo.version) > 0) {
        state = state.copyWith(availableVersion: remoteVersion);
      }
    } catch (_) {
      // Silent - a failed/offline check just leaves no upgrade button
      // shown, not worth an error banner on every Settings visit.
    }
  }

  Future<void> downloadAndInstall() async {
    final version = state.availableVersion;
    if (version == null || state.downloading) return;

    state = state.copyWith(downloading: true, progress: 0);
    try {
      final dir = await getTemporaryDirectory();
      final path = '${dir.path}/cwclock-v$version.apk';
      await Dio().download(
        _apkUrl(version),
        path,
        onReceiveProgress: (received, total) {
          if (total > 0) state = state.copyWith(progress: received / total);
        },
      );

      final result = await OpenFilex.open(path);
      if (result.type != ResultType.done) {
        throw Exception(result.message);
      }
    } finally {
      state = state.copyWith(downloading: false);
    }
  }
}

final appUpdateProvider = NotifierProvider<AppUpdateNotifier, AppUpdateState>(AppUpdateNotifier.new);
