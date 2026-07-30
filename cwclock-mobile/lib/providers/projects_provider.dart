import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/project.dart';
import 'api_providers.dart';

class ProjectsState {
  final List<Project> items;

  const ProjectsState({this.items = const []});

  ProjectsState copyWith({List<Project>? items}) => ProjectsState(items: items ?? this.items);
}

/// Ported from src/redux/projects/projects.actions.js + projects.reducer.js.
class ProjectsNotifier extends Notifier<ProjectsState> {
  @override
  ProjectsState build() => const ProjectsState();

  Future<List<Project>> listProjects(String orgId) async {
    final response = await ref.read(apiClientProvider).dio.get('/organizations/$orgId/projects/');
    final items = (response.data as List).map((e) => Project.fromJson(e as Map<String, dynamic>)).toList();
    state = state.copyWith(items: items);
    return items;
  }

  /// Creates a project under a specific client - the API nests project
  /// creation under its client, unlike update/delete which address the
  /// project directly.
  Future<Project> createProject(String orgId, String clientId, Map<String, dynamic> fields) async {
    final response = await ref
        .read(apiClientProvider)
        .dio
        .post('/organizations/$orgId/clients/$clientId/projects/', data: fields);
    final project = Project.fromJson(response.data as Map<String, dynamic>);
    state = state.copyWith(items: [...state.items, project]);
    return project;
  }

  /// fields may include clientId - the API allows moving a project to a
  /// different client of the same organization this way.
  Future<Project> updateProject(String orgId, String projectId, Map<String, dynamic> fields) async {
    final response =
        await ref.read(apiClientProvider).dio.put('/organizations/$orgId/projects/$projectId', data: fields);
    final updated = Project.fromJson(response.data as Map<String, dynamic>);
    state = state.copyWith(items: [for (final p in state.items) p.id == updated.id ? updated : p]);
    return updated;
  }

  Future<void> deleteProject(String orgId, String projectId) async {
    await ref.read(apiClientProvider).dio.delete('/organizations/$orgId/projects/$projectId');
    state = state.copyWith(items: state.items.where((p) => p.id != projectId).toList());
  }
}

final projectsProvider = NotifierProvider<ProjectsNotifier, ProjectsState>(ProjectsNotifier.new);
