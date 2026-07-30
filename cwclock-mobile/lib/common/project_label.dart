import '../models/client.dart';
import '../models/project.dart';

/// Renders a project as "{client} - {project}" for pickers.
String projectLabel(Project project, List<Client> clients) {
  Client? client;
  for (final c in clients) {
    if (c.id == project.clientId) {
      client = c;
      break;
    }
  }
  return '${client?.name ?? '?'} - ${project.name}';
}
