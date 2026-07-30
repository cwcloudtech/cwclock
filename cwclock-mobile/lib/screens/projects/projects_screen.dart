import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../providers/clients_provider.dart';
import '../../providers/locale_provider.dart';
import '../../providers/projects_provider.dart';
import '../../providers/session_provider.dart';
import '../../theme.dart';

/// Ported from src/screens/projects/ProjectsScreen.js.
class ProjectsScreen extends ConsumerStatefulWidget {
  const ProjectsScreen({super.key});

  @override
  ConsumerState<ProjectsScreen> createState() => _ProjectsScreenState();
}

class _ProjectsScreenState extends ConsumerState<ProjectsScreen> {
  @override
  void initState() {
    super.initState();
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId != null) {
      Future.microtask(() {
        ref.read(clientsProvider.notifier).listClients(orgId);
        ref.read(projectsProvider.notifier).listProjects(orgId);
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);
    final projects = ref.watch(projectsProvider).items;
    final clients = ref.watch(clientsProvider).items;

    return Scaffold(
      appBar: AppBar(title: Text(t('projects.title'))),
      body: SafeArea(
        child: projects.isEmpty
            ? Padding(
                padding: EdgeInsets.all(AppSpacing.of(2)),
                child: Text(t('projects.noProjects'), style: const TextStyle(color: AppColors.textMuted)),
              )
            : ListView.separated(
                padding: EdgeInsets.symmetric(horizontal: AppSpacing.of(2)),
                itemCount: projects.length,
                separatorBuilder: (_, _) => const Divider(height: 1, color: AppColors.border),
                itemBuilder: (context, index) {
                  final project = projects[index];
                  final client = clients.where((c) => c.id == project.clientId).firstOrNull;
                  return ListTile(
                    contentPadding: EdgeInsets.zero,
                    leading: Container(
                      width: 12,
                      height: 12,
                      decoration: BoxDecoration(
                        color: project.color != null ? colorFromHex(project.color!) : AppColors.primary,
                        shape: BoxShape.circle,
                      ),
                    ),
                    title: Text(project.name),
                    subtitle: client != null ? Text(client.name) : null,
                    onTap: () => context.push('/projects/form', extra: project),
                  );
                },
              ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => context.push('/projects/form'),
        backgroundColor: AppColors.primary,
        child: const Icon(Icons.add, color: AppColors.white),
      ),
    );
  }
}

extension _FirstOrNull<T> on Iterable<T> {
  T? get firstOrNull {
    final it = iterator;
    return it.moveNext() ? it.current : null;
  }
}
