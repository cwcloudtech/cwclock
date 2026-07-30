import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../providers/countries_provider.dart';
import '../../providers/currencies_provider.dart';
import '../../providers/locale_provider.dart';
import '../../providers/organizations_provider.dart';
import '../../providers/permissions.dart' as perm;
import '../../providers/session_provider.dart';
import '../../widgets/app_button.dart';
import '../../widgets/app_screen.dart';
import '../../widgets/error_banner.dart';
import '../../widgets/form_field.dart';
import '../../widgets/select_field.dart';

/// View/edit the current organization's profile. Editing is owner-only,
/// matching OrganizationHandler.Update's route gate - other roles see a
/// read-only view. Ported from
/// src/screens/organizations/OrganizationScreen.js.
class OrganizationScreen extends ConsumerStatefulWidget {
  const OrganizationScreen({super.key});

  @override
  ConsumerState<OrganizationScreen> createState() => _OrganizationScreenState();
}

class _OrganizationScreenState extends ConsumerState<OrganizationScreen> {
  Map<String, String>? _fields;
  String? _error;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      ref.read(countriesProvider.notifier).listCountries();
      ref.read(currenciesProvider.notifier).listCurrencies();
    });
  }

  void _setField(String key, String value) => setState(() => _fields = {...?_fields, key: value});

  Future<void> _handleSave() async {
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null || _fields == null) return;
    final locale = ref.read(localeProvider);
    setState(() {
      _error = null;
      _saving = true;
    });
    try {
      await ref.read(organizationsProvider.notifier).updateOrganization(orgId, _fields!);
    } catch (e) {
      setState(() => _error = apiErrorMessage(asApiException(e), locale));
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);
    final session = ref.watch(sessionProvider);
    final orgsState = ref.watch(organizationsProvider);
    final countries = ref.watch(countriesProvider);
    final currencies = ref.watch(currenciesProvider);

    final org = orgsState.items.where((o) => o.id == session.orgId).firstOrNull;
    if (org == null) return const SizedBox.shrink();

    _fields ??= {
      'name': org.name,
      'accountingEmail': org.accountingEmail ?? '',
      'address': org.address ?? '',
      'postalCode': org.postalCode ?? '',
      'city': org.city ?? '',
      'country': org.country ?? '',
      'currency': org.currency ?? '',
      'vatNumber': org.vatNumber ?? '',
      'identificationNumber': org.identificationNumber ?? '',
      'iban': org.iban ?? '',
      'bic': org.bic ?? '',
    };
    final fields = _fields!;
    final canEdit = perm.isOrgOwner(session.user, org);

    final countryItems = [for (final c in countries) SelectItem(c.iso, c.name)];
    final currencyItems = [for (final c in currencies) SelectItem(c.iso, c.name)];

    return Scaffold(
      appBar: AppBar(title: Text(t('organizations.profileTitle'))),
      body: AppScreen(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (!canEdit) ErrorBanner(message: t('organizations.ownerOnlyHint')),
            AppFormField(
              label: t('organizations.name'),
              value: fields['name']!,
              onChanged: (v) => _setField('name', v),
              enabled: canEdit,
            ),
            AppFormField(
              label: t('organizations.accountingEmail'),
              value: fields['accountingEmail']!,
              onChanged: (v) => _setField('accountingEmail', v),
              enabled: canEdit,
              keyboardType: TextInputType.emailAddress,
              textCapitalization: TextCapitalization.none,
            ),
            AppFormField(
              label: t('organizations.address'),
              value: fields['address']!,
              onChanged: (v) => _setField('address', v),
              enabled: canEdit,
            ),
            AppFormField(
              label: t('organizations.postalCode'),
              value: fields['postalCode']!,
              onChanged: (v) => _setField('postalCode', v),
              enabled: canEdit,
            ),
            AppFormField(
              label: t('organizations.city'),
              value: fields['city']!,
              onChanged: (v) => _setField('city', v),
              enabled: canEdit,
            ),
            AppSelectField(
              label: t('organizations.country'),
              value: fields['country']!,
              onChanged: (v) => _setField('country', v),
              items: countryItems,
              placeholder: t('organizations.country'),
              enabled: canEdit,
            ),
            AppSelectField(
              label: t('organizations.currency'),
              value: fields['currency']!,
              onChanged: (v) => _setField('currency', v),
              items: currencyItems,
              placeholder: t('organizations.currency'),
              enabled: canEdit,
            ),
            AppFormField(
              label: t('organizations.vatNumber'),
              value: fields['vatNumber']!,
              onChanged: (v) => _setField('vatNumber', v),
              enabled: canEdit,
            ),
            AppFormField(
              label: t('organizations.identificationNumber'),
              value: fields['identificationNumber']!,
              onChanged: (v) => _setField('identificationNumber', v),
              enabled: canEdit,
            ),
            AppFormField(
              label: t('organizations.iban'),
              value: fields['iban']!,
              onChanged: (v) => _setField('iban', v),
              enabled: canEdit,
              textCapitalization: TextCapitalization.characters,
            ),
            AppFormField(
              label: t('organizations.bic'),
              value: fields['bic']!,
              onChanged: (v) => _setField('bic', v),
              enabled: canEdit,
              textCapitalization: TextCapitalization.characters,
            ),
            ErrorBanner(message: _error),
            if (canEdit) AppButton(title: t('common.save'), onPressed: _handleSave, loading: _saving),
          ],
        ),
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
