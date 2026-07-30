import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../models/client.dart';
import '../../providers/clients_provider.dart';
import '../../providers/countries_provider.dart';
import '../../providers/locale_provider.dart';
import '../../providers/session_provider.dart';
import '../../theme.dart';
import '../../widgets/app_button.dart';
import '../../widgets/app_screen.dart';
import '../../widgets/error_banner.dart';
import '../../widgets/form_field.dart';
import '../../widgets/select_field.dart';
import '../../widgets/toggle_row.dart';
import '../../widgets/app_top_bar.dart';

/// Create (client is null) or edit (present). Ported from
/// src/screens/clients/ClientFormScreen.js.
class ClientFormScreen extends ConsumerStatefulWidget {
  final Client? client;

  const ClientFormScreen({super.key, this.client});

  @override
  ConsumerState<ClientFormScreen> createState() => _ClientFormScreenState();
}

class _ClientFormScreenState extends ConsumerState<ClientFormScreen> {
  late ClientFields _fields =
      widget.client != null ? ClientFields.fromClient(widget.client!) : const ClientFields();
  String? _error;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    Future.microtask(() => ref.read(countriesProvider.notifier).listCountries());
  }

  Future<void> _handleSave() async {
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;
    final locale = ref.read(localeProvider);

    setState(() {
      _error = null;
      _saving = true;
    });
    try {
      if (widget.client != null) {
        await ref.read(clientsProvider.notifier).updateClient(orgId, widget.client!.id, _fields.toJson());
      } else {
        await ref.read(clientsProvider.notifier).createClient(orgId, _fields.toJson());
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
    if (orgId == null || widget.client == null) return;

    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: Text(t('clients.deleteClientTitle')),
        content: Text(t('clients.deleteClientBody', {'name': widget.client!.name})),
        actions: [
          TextButton(onPressed: () => Navigator.pop(dialogContext), child: Text(t('common.cancel'))),
          TextButton(
            onPressed: () async {
              Navigator.pop(dialogContext);
              await ref.read(clientsProvider.notifier).deleteClient(orgId, widget.client!.id);
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
    final countries = ref.watch(countriesProvider);
    final countryItems = [for (final c in countries) SelectItem(c.iso, c.name)];
    final f = _fields;

    return Scaffold(
      appBar: AppTopBar(title: widget.client != null ? t('clients.editClient') : t('clients.addClient')),
      body: AppScreen(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            AppFormField(label: t('clients.name'), value: f.name, onChanged: (v) => setState(() => _fields = f.copyWith(name: v))),
            AppFormField(
              label: t('clients.email'),
              value: f.email,
              onChanged: (v) => setState(() => _fields = f.copyWith(email: v)),
              keyboardType: TextInputType.emailAddress,
              textCapitalization: TextCapitalization.none,
            ),
            AppFormField(
              label: t('clients.contactName'),
              value: f.contactName,
              onChanged: (v) => setState(() => _fields = f.copyWith(contactName: v)),
            ),
            AppFormField(
              label: t('clients.address'),
              value: f.address,
              onChanged: (v) => setState(() => _fields = f.copyWith(address: v)),
            ),
            AppFormField(
              label: t('clients.postalCode'),
              value: f.postalCode,
              onChanged: (v) => setState(() => _fields = f.copyWith(postalCode: v)),
            ),
            AppFormField(
              label: t('clients.city'),
              value: f.city,
              onChanged: (v) => setState(() => _fields = f.copyWith(city: v)),
            ),
            AppSelectField(
              label: t('clients.country'),
              value: f.country,
              onChanged: (v) => setState(() => _fields = f.copyWith(country: v)),
              items: countryItems,
              placeholder: t('clients.country'),
            ),
            AppFormField(
              label: t('clients.vatNumber'),
              value: f.vatNumber,
              onChanged: (v) => setState(() => _fields = f.copyWith(vatNumber: v)),
            ),
            AppFormField(
              label: t('clients.vatRate'),
              value: f.vatRate,
              onChanged: (v) => setState(() => _fields = f.copyWith(vatRate: v)),
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
            ),
            AppFormField(
              label: t('clients.identificationNumber'),
              value: f.identificationNumber,
              onChanged: (v) => setState(() => _fields = f.copyWith(identificationNumber: v)),
            ),
            AppFormField(
              label: t('clients.purchaseOrder'),
              value: f.purchaseOrder,
              onChanged: (v) => setState(() => _fields = f.copyWith(purchaseOrder: v)),
            ),
            AppFormField(
              label: t('clients.hoursPerDay'),
              value: f.hoursPerDay,
              onChanged: (v) => setState(() => _fields = f.copyWith(hoursPerDay: v)),
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
            ),
            AppFormField(
              label: t('clients.dailyRate'),
              value: f.dailyRate,
              onChanged: (v) => setState(() => _fields = f.copyWith(dailyRate: v)),
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
            ),
            AppToggleRow(
              label: t('clients.sendReportsWithInvoice'),
              value: f.sendReportsWithInvoice,
              onChanged: (v) => setState(() => _fields = f.copyWith(sendReportsWithInvoice: v)),
            ),
            ErrorBanner(message: _error),
            AppButton(
              title: t('common.save'),
              onPressed: f.name.trim().isEmpty || f.country.isEmpty ? null : _handleSave,
              loading: _saving,
              margin: EdgeInsets.only(bottom: AppSpacing.of(1.5)),
            ),
            if (widget.client != null)
              AppButton(title: t('common.delete'), variant: AppButtonVariant.danger, onPressed: _handleDelete),
          ],
        ),
      ),
    );
  }
}
