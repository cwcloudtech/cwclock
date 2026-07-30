import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../models/project.dart';
import '../../providers/clients_provider.dart';
import '../../providers/locale_provider.dart';
import '../../providers/projects_provider.dart';
import '../../providers/session_provider.dart';
import '../../theme.dart';
import '../../widgets/app_button.dart';
import '../../widgets/app_screen.dart';
import '../../widgets/color_swatch_picker.dart';
import '../../widgets/error_banner.dart';
import '../../widgets/form_field.dart';
import '../../widgets/select_field.dart';
import '../../widgets/app_top_bar.dart';

/// Create (project is null) or edit (present). Color uses a fixed swatch row
/// instead of a native color picker. Ported from
/// src/screens/projects/ProjectFormScreen.js.
class ProjectFormScreen extends ConsumerStatefulWidget {
  final Project? project;

  const ProjectFormScreen({super.key, this.project});

  @override
  ConsumerState<ProjectFormScreen> createState() => _ProjectFormScreenState();
}

class _ProjectFormScreenState extends ConsumerState<ProjectFormScreen> {
  late String _name = widget.project?.name ?? '';
  late String _clientId = widget.project?.clientId ?? '';
  late String _color = widget.project?.color ?? defaultProjectColor;
  late String _dailyRate = widget.project?.dailyRate != null ? widget.project!.dailyRate.toString() : '';
  late String _subdivisionsText = subdivisionsToText(widget.project?.subdivisions ?? const []);
  String? _error;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId != null) {
      Future.microtask(() => ref.read(clientsProvider.notifier).listClients(orgId));
    }
  }

  Future<void> _handleSave() async {
    if (_name.trim().isEmpty || _clientId.isEmpty) return;
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;
    final locale = ref.read(localeProvider);

    setState(() {
      _error = null;
      _saving = true;
    });
    final fields = {
      'name': _name.trim(),
      'color': _color,
      'dailyRate': _dailyRate.isEmpty ? null : num.tryParse(_dailyRate),
      'subdivisions': textToSubdivisions(_subdivisionsText),
    };
    try {
      if (widget.project != null) {
        await ref
            .read(projectsProvider.notifier)
            .updateProject(orgId, widget.project!.id, {...fields, 'clientId': _clientId});
      } else {
        await ref.read(projectsProvider.notifier).createProject(orgId, _clientId, fields);
      }
      if (mounted) Navigator.of(context).pop();
    } catch (e) {
      setState(() => _error = apiErrorMessage(asApiException(e), locale));
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  void _handleDelete() {
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null || widget.project == null) return;

    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: Text(t('projects.deleteProjectTitle')),
        content: Text(t('projects.deleteProjectBody', {'name': widget.project!.name})),
        actions: [
          TextButton(onPressed: () => Navigator.pop(dialogContext), child: Text(t('common.cancel'))),
          TextButton(
            onPressed: () async {
              Navigator.pop(dialogContext);
              await ref.read(projectsProvider.notifier).deleteProject(orgId, widget.project!.id);
              if (!mounted) return;
              Navigator.of(context).pop();
            },
            child: Text(t('common.delete'), style: TextStyle(color: AppColors.of(context).danger)),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);
    final clients = ref.watch(clientsProvider).items;
    final clientItems = [for (final c in clients) SelectItem(c.id, c.name)];

    return Scaffold(
      appBar: AppTopBar(title: widget.project != null ? t('projects.editProject') : t('projects.addProject')),
      body: AppScreen(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            AppFormField(label: t('projects.name'), value: _name, onChanged: (v) => setState(() => _name = v)),
            AppSelectField(
              label: t('projects.client'),
              value: _clientId,
              onChanged: (v) => setState(() => _clientId = v),
              items: clientItems,
              placeholder: t('projects.client'),
            ),
            ColorSwatchPicker(
              label: t('projects.color'),
              value: _color,
              onChanged: (v) => setState(() => _color = v),
            ),
            AppFormField(
              label: t('projects.dailyRate'),
              value: _dailyRate,
              onChanged: (v) => setState(() => _dailyRate = v),
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
            ),
            AppFormField(
              label: t('projects.subdivisions'),
              value: _subdivisionsText,
              onChanged: (v) => setState(() => _subdivisionsText = v),
              placeholder: t('projects.subdivisionsHint'),
            ),
            ErrorBanner(message: _error),
            AppButton(
              title: t('common.save'),
              onPressed: _name.trim().isEmpty || _clientId.isEmpty ? null : _handleSave,
              loading: _saving,
              margin: EdgeInsets.only(bottom: AppSpacing.of(1.5)),
            ),
            if (widget.project != null)
              AppButton(title: t('common.delete'), variant: AppButtonVariant.danger, onPressed: _handleDelete),
          ],
        ),
      ),
    );
  }
}
