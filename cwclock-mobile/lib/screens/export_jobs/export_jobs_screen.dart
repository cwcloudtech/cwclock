import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../models/export_job.dart';
import '../../providers/export_jobs_provider.dart';
import '../../providers/locale_provider.dart';
import '../../providers/session_provider.dart';
import '../../theme.dart';
import '../../widgets/app_top_bar.dart';

/// Mobile only runs jobs that were already configured on the web app - no
/// create/edit/delete/enable-toggle here, just a play button per job
/// (ai-instruct-124). Ported (in reduced form) from
/// src/screens/ExportJobs/ExportJobs.jsx.
class ExportJobsScreen extends ConsumerStatefulWidget {
  const ExportJobsScreen({super.key});

  @override
  ConsumerState<ExportJobsScreen> createState() => _ExportJobsScreenState();
}

class _ExportJobsScreenState extends ConsumerState<ExportJobsScreen> {
  final Set<String> _runningIds = {};

  @override
  void initState() {
    super.initState();
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId != null) {
      Future.microtask(() => ref.read(exportJobsProvider.notifier).listExportJobs(orgId));
    }
  }

  Future<void> _handleRun(ExportJob job) async {
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);

    setState(() => _runningIds.add(job.id));
    try {
      await ref.read(exportJobsProvider.notifier).runExportJob(orgId, job.id);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(t('exportJobs.runSuccess'))));
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(apiErrorMessage(asApiException(e), locale))),
        );
      }
    } finally {
      if (mounted) setState(() => _runningIds.remove(job.id));
    }
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);
    final jobs = ref.watch(exportJobsProvider).items;

    return Scaffold(
      appBar: AppTopBar(title: t('exportJobs.title')),
      body: SafeArea(
        child: jobs.isEmpty
            ? Padding(
                padding: EdgeInsets.all(AppSpacing.of(2)),
                child: Text(t('exportJobs.noJobs'), style: TextStyle(color: AppColors.of(context).textMuted)),
              )
            : ListView.separated(
                padding: EdgeInsets.symmetric(horizontal: AppSpacing.of(2)),
                itemCount: jobs.length,
                separatorBuilder: (_, _) => Divider(height: 1, color: AppColors.of(context).border),
                itemBuilder: (context, index) {
                  final job = jobs[index];
                  final running = _runningIds.contains(job.id);
                  return ListTile(
                    contentPadding: EdgeInsets.zero,
                    title: Text(job.name, maxLines: 1, overflow: TextOverflow.ellipsis),
                    subtitle: Text(
                      '${job.cronExpression} · ${job.reportTypes.join(', ')}',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    trailing: running
                        ? SizedBox(
                            width: 24,
                            height: 24,
                            child: CircularProgressIndicator(strokeWidth: 2, color: AppColors.of(context).primary),
                          )
                        : IconButton(
                            onPressed: () => _handleRun(job),
                            icon: Icon(Icons.play_circle_outline, color: AppColors.of(context).primary),
                            tooltip: t('exportJobs.runNow'),
                          ),
                  );
                },
              ),
      ),
    );
  }
}
