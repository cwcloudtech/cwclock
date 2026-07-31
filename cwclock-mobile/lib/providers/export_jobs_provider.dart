import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/export_job.dart';
import 'api_providers.dart';

class ExportJobsState {
  final List<ExportJob> items;

  const ExportJobsState({this.items = const []});

  ExportJobsState copyWith({List<ExportJob>? items}) => ExportJobsState(items: items ?? this.items);
}

/// Mobile only exposes "run now" for jobs already configured on the web app
/// (see ExportJobsScreen) - creating/editing/deleting a job's schedule and
/// targets needs the full form the web ExportJobs page has, which is out of
/// scope here (ai-instruct-124).
class ExportJobsNotifier extends Notifier<ExportJobsState> {
  @override
  ExportJobsState build() => const ExportJobsState();

  Future<List<ExportJob>> listExportJobs(String orgId) async {
    final response = await ref.read(apiClientProvider).dio.get('/organizations/$orgId/export-jobs/');
    final items = (response.data as List).map((e) => ExportJob.fromJson(e as Map<String, dynamic>)).toList();
    state = state.copyWith(items: items);
    return items;
  }

  Future<void> runExportJob(String orgId, String jobId) {
    return ref.read(apiClientProvider).dio.post('/organizations/$orgId/export-jobs/$jobId/run');
  }
}

final exportJobsProvider = NotifierProvider<ExportJobsNotifier, ExportJobsState>(ExportJobsNotifier.new);
