import 'dart:async';
import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../i18n/app_localizations.dart';
import '../../models/ai.dart';
import '../../providers/ai_provider.dart';
import '../../providers/locale_provider.dart';
import '../../providers/session_provider.dart';
import '../../theme.dart';
import '../../widgets/app_button.dart';
import '../../widgets/app_top_bar.dart';
import '../../widgets/form_field.dart';
import '../../widgets/select_field.dart';

/// Manage the user's AI subscriptions (ai-instruct-132): add a connection to a
/// provider with your own key, share it with other users, pick the default, and
/// delete one. Mirrors cwclock-ui's AiSubscriptions settings card.
class AiSubscriptionsScreen extends ConsumerStatefulWidget {
  const AiSubscriptionsScreen({super.key});

  @override
  ConsumerState<AiSubscriptionsScreen> createState() => _AiSubscriptionsScreenState();
}

class _AiSubscriptionsScreenState extends ConsumerState<AiSubscriptionsScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() async {
      try {
        await ref.read(aiProvider.notifier).listProviders();
        await ref.read(aiProvider.notifier).listConnections();
      } catch (_) {
        /* shown via snackbar on explicit actions */
      }
    });
  }

  String Function(String, [Map<String, String>?]) get _t => translateWith(ref.read(localeProvider));

  void _error(Object e) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(apiErrorMessage(e, ref.read(localeProvider)))),
    );
  }

  Future<void> _showForm({AiConnection? editing}) async {
    final t = _t;
    final providers = ref.read(aiProvider).providers;
    var provider = editing?.provider ?? '';
    var name = editing?.name ?? '';
    var token = '';
    var model = editing?.model ?? '';
    var baseUrl = editing?.baseUrl ?? '';
    var maxTokens = editing != null && editing.maxTokens > 0 ? editing.maxTokens.toString() : '';
    var busy = false;
    String? formError;
    final suffix = _randomSuffix();

    await showDialog<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setLocal) {
          final selected = providers.where((p) => p.name == provider).firstOrNull;
          final effectiveModel =
              model.trim().isNotEmpty ? model.trim() : (selected?.defaultModel ?? '');
          final suggestedName = effectiveModel.isNotEmpty ? '$effectiveModel-$suffix' : '';
          return AlertDialog(
            title: Text(editing != null ? t('ai.editSubscription') : t('ai.addSubscription')),
            content: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  AppSelectField(
                    label: t('ai.provider'),
                    value: provider,
                    placeholder: t('ai.pickProvider'),
                    items: [for (final p in providers) SelectItem(p.name, p.label)],
                    onChanged: (v) => setLocal(() => provider = v),
                  ),
                  AppFormField(
                    label: t('ai.name'),
                    value: name,
                    placeholder: suggestedName.isNotEmpty ? suggestedName : t('ai.namePlaceholder'),
                    onChanged: (v) => name = v,
                  ),
                  AppFormField(
                    label: t('ai.apiKey'),
                    value: token,
                    obscureText: true,
                    placeholder: editing != null ? t('ai.keyLeaveBlank') : (selected?.keyHint ?? ''),
                    onChanged: (v) => token = v,
                  ),
                  _AutocompleteField(
                    label: t('ai.model'),
                    initialValue: model,
                    placeholder: selected?.defaultModel ?? t('ai.modelPlaceholder'),
                    optionsBuilder: (text) {
                      final models = selected?.models ?? const <String>[];
                      final q = text.trim().toLowerCase();
                      if (q.isEmpty) return models;
                      return models.where((m) => m.toLowerCase().contains(q));
                    },
                    onChanged: (v) => setLocal(() => model = v),
                  ),
                  AppFormField(
                    label: t('ai.baseUrl'),
                    value: baseUrl,
                    placeholder: selected?.defaultUrl ?? 'https://...',
                    onChanged: (v) => baseUrl = v,
                  ),
                  AppFormField(
                    label: t('ai.maxTokens'),
                    value: maxTokens,
                    keyboardType: TextInputType.number,
                    placeholder: t('ai.maxTokensPlaceholder'),
                    onChanged: (v) => maxTokens = v,
                  ),
                  if (formError != null)
                    Padding(
                      padding: const EdgeInsets.only(top: 4),
                      child: Text(formError!, style: TextStyle(color: AppColors.of(context).danger)),
                    ),
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: busy ? null : () => Navigator.pop(context),
                child: Text(t('common.cancel')),
              ),
              TextButton(
                onPressed: busy
                    ? null
                    : () async {
                        if (provider.isEmpty) {
                          setLocal(() => formError = t('ai.pickProviderError'));
                          return;
                        }
                        if (editing == null && token.trim().isEmpty) {
                          setLocal(() => formError = t('ai.keyRequired'));
                          return;
                        }
                        setLocal(() {
                          busy = true;
                          formError = null;
                        });
                        final payload = {
                          'provider': provider,
                          'name': name.trim().isNotEmpty ? name.trim() : suggestedName,
                          'token': token.trim(),
                          'model': model.trim(),
                          'baseUrl': baseUrl.trim(),
                          'maxTokens': int.tryParse(maxTokens.trim()) ?? 0,
                        };
                        try {
                          if (editing != null) {
                            await ref.read(aiProvider.notifier).updateConnection(editing.id, payload);
                          } else {
                            await ref.read(aiProvider.notifier).createConnection(payload);
                          }
                          if (context.mounted) Navigator.pop(context);
                        } catch (e) {
                          setLocal(() {
                            busy = false;
                            formError = apiErrorMessage(e, ref.read(localeProvider));
                          });
                        }
                      },
                child: Text(editing != null ? t('common.save') : t('ai.addSubscription')),
              ),
            ],
          );
        },
      ),
    );
  }

  Future<void> _confirmDelete(AiConnection c) async {
    final t = _t;
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(t('ai.deleteTitle')),
        content: Text(t('ai.deleteBody', {'name': c.name.isNotEmpty ? c.name : c.provider})),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: Text(t('common.cancel'))),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(t('common.delete'), style: TextStyle(color: AppColors.of(context).danger)),
          ),
        ],
      ),
    );
    if (ok == true) {
      try {
        await ref.read(aiProvider.notifier).deleteConnection(c.id);
      } catch (e) {
        _error(e);
      }
    }
  }

  Future<void> _setDefault(AiConnection c) async {
    try {
      await ref.read(aiProvider.notifier).setDefault(c.id);
    } catch (e) {
      _error(e);
    }
  }

  Future<void> _showShares(AiConnection c) async {
    final t = _t;
    List<AiConnectionShare> shares = [];
    try {
      shares = await ref.read(aiProvider.notifier).listShares(c.id);
    } catch (e) {
      _error(e);
    }
    if (!mounted) return;
    var email = '';
    var busy = false;

    await showDialog<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setLocal) => AlertDialog(
          title: Text(t('ai.shareTitle')),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(t('ai.shareIntro'), style: TextStyle(color: AppColors.of(context).textMuted, fontSize: 13)),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Expanded(
                      child: _AutocompleteField(
                        initialValue: email,
                        keyboardType: TextInputType.emailAddress,
                        placeholder: t('ai.shareEmailPlaceholder'),
                        optionsBuilder: (text) async {
                          try {
                            return await ref.read(aiProvider.notifier).searchShareableEmails(text.trim());
                          } catch (_) {
                            return const <String>[];
                          }
                        },
                        onChanged: (v) => setLocal(() => email = v),
                      ),
                    ),
                    const SizedBox(width: 8),
                    IconButton(
                      icon: const Icon(Icons.add),
                      onPressed: busy
                          ? null
                          : () async {
                              if (email.trim().isEmpty) return;
                              setLocal(() => busy = true);
                              try {
                                shares = await ref.read(aiProvider.notifier).shareConnection(c.id, email.trim());
                                setLocal(() {
                                  email = '';
                                  busy = false;
                                });
                              } catch (e) {
                                setLocal(() => busy = false);
                                _error(e);
                              }
                            },
                    ),
                  ],
                ),
                if (shares.isEmpty)
                  Padding(
                    padding: const EdgeInsets.only(top: 4),
                    child: Text(t('ai.shareEmpty'),
                        style: TextStyle(color: AppColors.of(context).textMuted, fontStyle: FontStyle.italic)),
                  )
                else
                  for (final s in shares)
                    ListTile(
                      dense: true,
                      contentPadding: EdgeInsets.zero,
                      title: Text(s.label),
                      trailing: IconButton(
                        icon: Icon(Icons.delete_outline, color: AppColors.of(context).danger),
                        onPressed: () async {
                          try {
                            await ref.read(aiProvider.notifier).unshareConnection(c.id, s.userId);
                            setLocal(() => shares = shares.where((x) => x.userId != s.userId).toList());
                          } catch (e) {
                            _error(e);
                          }
                        },
                      ),
                    ),
              ],
            ),
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(context), child: Text(t('common.close'))),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final t = translateWith(ref.watch(localeProvider));
    final ai = ref.watch(aiProvider);
    final userId = ref.watch(sessionProvider).user?.id;
    final colors = AppColors.of(context);

    return Scaffold(
      appBar: AppTopBar(title: t('ai.subscriptionsTitle')),
      backgroundColor: colors.background,
      body: SafeArea(
        child: SingleChildScrollView(
          padding: EdgeInsets.all(AppSpacing.of(2)),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(t('ai.intro'), style: TextStyle(color: colors.textMuted, fontSize: 14)),
              AppButton(
                title: t('ai.addSubscription'),
                icon: Icons.add,
                onPressed: () => _showForm(),
                margin: EdgeInsets.only(top: AppSpacing.of(2)),
              ),
              if (ai.connections.isEmpty)
                Padding(
                  padding: EdgeInsets.only(top: AppSpacing.of(3)),
                  child: Center(
                    child: Text(t('ai.noSubscriptions'),
                        style: TextStyle(color: colors.textMuted, fontStyle: FontStyle.italic)),
                  ),
                ),
              for (var index = 0; index < ai.connections.length; index++)
                _connectionRow(context, t, ai.connections[index], index == 0, userId),
            ],
          ),
        ),
      ),
    );
  }

  Widget _connectionRow(
    BuildContext context,
    String Function(String, [Map<String, String>?]) t,
    AiConnection c,
    bool isDefault,
    String? userId,
  ) {
    final colors = AppColors.of(context);
    final owned = c.ownerId == userId || c.isOwner;
    return Container(
      margin: EdgeInsets.only(top: AppSpacing.of(1.5)),
      padding: EdgeInsets.all(AppSpacing.of(1.5)),
      decoration: BoxDecoration(
        color: colors.backgroundMuted,
        borderRadius: BorderRadius.circular(AppRadius.value),
      ),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Flexible(
                      child: Text(
                        c.name.isNotEmpty ? c.name : c.provider,
                        style: TextStyle(fontWeight: FontWeight.w600, color: colors.text),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    if (isDefault) _badge(context, t('ai.default'), colors.primary),
                    if (!owned) _badge(context, t('ai.shared'), colors.textMuted),
                  ],
                ),
                const SizedBox(height: 2),
                Text(
                  c.model.isNotEmpty ? '${c.provider} · ${c.model}' : c.provider,
                  style: TextStyle(fontSize: 12, color: colors.textMuted),
                ),
              ],
            ),
          ),
          if (!isDefault)
            IconButton(
              tooltip: t('ai.setDefault'),
              icon: Icon(Icons.star_border, color: colors.textMuted),
              onPressed: () => _setDefault(c),
            ),
          if (owned) ...[
            IconButton(
              tooltip: t('ai.share'),
              icon: Icon(Icons.share_outlined, color: colors.textMuted),
              onPressed: () => _showShares(c),
            ),
            IconButton(
              tooltip: t('common.edit'),
              icon: Icon(Icons.edit_outlined, color: colors.textMuted),
              onPressed: () => _showForm(editing: c),
            ),
            IconButton(
              tooltip: t('common.delete'),
              icon: Icon(Icons.delete_outline, color: colors.danger),
              onPressed: () => _confirmDelete(c),
            ),
          ],
        ],
      ),
    );
  }

  Widget _badge(BuildContext context, String text, Color color) {
    return Container(
      margin: const EdgeInsets.only(left: 6),
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(text, style: TextStyle(fontSize: 10, color: color, fontWeight: FontWeight.w600)),
    );
  }
}

extension _FirstOrNull<T> on Iterable<T> {
  T? get firstOrNull {
    final it = iterator;
    return it.moveNext() ? it.current : null;
  }
}

final _rand = Random();

/// Eight random hex chars, the tail of a suggested subscription name (same idea
/// as uprodit's AiConnections), so two connections on the same model read apart.
String _randomSuffix() =>
    List.generate(4, (_) => _rand.nextInt(256).toRadixString(16).padLeft(2, '0')).join();

/// A text field with a dropdown of suggestions (a datalist-style autocomplete),
/// used for the model (suggestions from the provider's catalogue) and the share
/// email (suggestions fetched from the server). Free text is always accepted.
class _AutocompleteField extends StatelessWidget {
  final String? label;
  final String initialValue;
  final FutureOr<Iterable<String>> Function(String text) optionsBuilder;
  final ValueChanged<String> onChanged;
  final String? placeholder;
  final TextInputType? keyboardType;

  const _AutocompleteField({
    this.label,
    required this.initialValue,
    required this.optionsBuilder,
    required this.onChanged,
    this.placeholder,
    this.keyboardType,
  });

  @override
  Widget build(BuildContext context) {
    final colors = AppColors.of(context);
    return Padding(
      padding: EdgeInsets.only(bottom: AppSpacing.of(2)),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (label != null)
            Padding(
              padding: EdgeInsets.only(bottom: AppSpacing.of(0.5)),
              child: Text(label!, style: TextStyle(fontSize: 13, color: colors.textMuted)),
            ),
          Autocomplete<String>(
            initialValue: TextEditingValue(text: initialValue),
            optionsBuilder: (value) => optionsBuilder(value.text),
            onSelected: onChanged,
            fieldViewBuilder: (context, controller, focusNode, onFieldSubmitted) {
              return TextField(
                controller: controller,
                focusNode: focusNode,
                keyboardType: keyboardType,
                textCapitalization: TextCapitalization.none,
                style: TextStyle(fontSize: 16, color: colors.text),
                decoration: InputDecoration(
                  hintText: placeholder,
                  hintStyle: TextStyle(color: colors.textMuted),
                  contentPadding: EdgeInsets.symmetric(
                    horizontal: AppSpacing.of(1.5),
                    vertical: AppSpacing.of(1.25),
                  ),
                ),
                onChanged: onChanged,
                onSubmitted: (_) => onFieldSubmitted(),
              );
            },
            optionsViewBuilder: (context, onSelected, options) {
              return Align(
                alignment: Alignment.topLeft,
                child: Material(
                  elevation: 4,
                  borderRadius: BorderRadius.circular(AppRadius.value),
                  child: ConstrainedBox(
                    constraints: const BoxConstraints(maxHeight: 220, maxWidth: 320),
                    child: ListView(
                      shrinkWrap: true,
                      padding: EdgeInsets.zero,
                      children: [
                        for (final option in options)
                          ListTile(
                            dense: true,
                            title: Text(option),
                            onTap: () => onSelected(option),
                          ),
                      ],
                    ),
                  ),
                ),
              );
            },
          ),
        ],
      ),
    );
  }
}
