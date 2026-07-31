import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../api/api_client.dart';
import '../../i18n/app_localizations.dart';
import '../../common/format.dart';
import '../../models/invoice.dart';
import '../../providers/clients_provider.dart';
import '../../providers/invoices_provider.dart';
import '../../providers/locale_provider.dart';
import '../../providers/session_provider.dart';
import '../../theme.dart';
import '../../widgets/app_button.dart';
import '../../widgets/date_field.dart';
import '../../widgets/error_banner.dart';
import '../../widgets/select_field.dart';
import '../../widgets/app_top_bar.dart';

DateTime _firstOfMonth() {
  final now = DateTime.now();
  return DateTime(now.year, now.month, 1);
}

/// Mirrors cwclock-ui's STATUSES/STATUS_LABEL_KEY (Invoices.jsx).
const _invoiceStatuses = ['unpaid', 'paid', 'canceled', 'refunded'];

String _invoiceStatusLabelKey(String status) {
  switch (status) {
    case 'paid':
      return 'invoices.statusPaid';
    case 'canceled':
      return 'invoices.statusCanceled';
    case 'refunded':
      return 'invoices.statusRefunded';
    default:
      return 'invoices.statusUnpaid';
  }
}

/// This tab is only shown when the connected user is admin/owner of the
/// current org (see main_tabs_screen.dart, mirroring cwclock-ui's
/// showInvoices gate). Preview never saves anything; Generate allocates a
/// real invoice number, so it's confirmed first. Ported from
/// src/screens/invoices/InvoicesScreen.js.
class InvoicesScreen extends ConsumerStatefulWidget {
  const InvoicesScreen({super.key});

  @override
  ConsumerState<InvoicesScreen> createState() => _InvoicesScreenState();
}

class _InvoicesScreenState extends ConsumerState<InvoicesScreen> {
  String _clientId = '';
  DateTime _start = _firstOfMonth();
  DateTime _end = DateTime.now();
  String? _error;
  bool _busy = false;
  final Set<String> _busyInvoiceIds = {};

  @override
  void initState() {
    super.initState();
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId != null) {
      Future.microtask(() => ref.read(clientsProvider.notifier).listClients(orgId));
    }
  }

  void _refreshInvoices() {
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId != null && _clientId.isNotEmpty) {
      ref
          .read(invoicesProvider.notifier)
          .listInvoices(orgId, _clientId, toDayString(_start), toDayString(_end));
    }
  }

  void _openPdf(String path, String title) {
    context.push('/pdf-viewer', extra: {'path': path, 'title': title});
  }

  Future<void> _handlePreview() async {
    if (_clientId.isEmpty) return;
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);

    setState(() {
      _error = null;
      _busy = true;
    });
    try {
      final path = await ref
          .read(invoicesProvider.notifier)
          .previewInvoicePdf(orgId, _clientId, toDayString(_start), toDayString(_end));
      if (mounted) _openPdf(path, t('invoices.preview'));
    } catch (e) {
      setState(() => _error = apiErrorMessage(asApiException(e), locale));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  void _handleGenerate() {
    if (_clientId.isEmpty) return;
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(t('invoices.generateConfirmTitle')),
        content: Text(t('invoices.generateConfirmBody')),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: Text(t('common.cancel'))),
          TextButton(
            onPressed: () async {
              Navigator.pop(context);
              setState(() {
                _error = null;
                _busy = true;
              });
              try {
                final path = await ref
                    .read(invoicesProvider.notifier)
                    .generateInvoicePdf(orgId, _clientId, toDayString(_start), toDayString(_end));
                _refreshInvoices();
                if (mounted) _openPdf(path, t('invoices.title'));
              } catch (e) {
                setState(() => _error = apiErrorMessage(asApiException(e), locale));
              } finally {
                if (mounted) setState(() => _busy = false);
              }
            },
            child: Text(t('invoices.generate')),
          ),
        ],
      ),
    );
  }

  Future<void> _handleOpenInvoice(Invoice invoice) async {
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);
    try {
      final path = await ref.read(invoicesProvider.notifier).downloadInvoicePdf(orgId, invoice.id);
      if (mounted) _openPdf(path, invoice.number);
    } catch (e) {
      if (mounted) {
        showDialog(
          context: context,
          builder: (context) => AlertDialog(
            title: Text(t('common.retry')),
            content: Text(apiErrorMessage(asApiException(e), locale)),
            actions: [TextButton(onPressed: () => Navigator.pop(context), child: Text(t('common.close')))],
          ),
        );
      }
    }
  }

  Future<void> _runInvoiceAction(
    Invoice invoice,
    Future<void> Function(String orgId, String invoiceId) action,
    String successMessage,
  ) async {
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId == null) return;
    final locale = ref.read(localeProvider);

    setState(() => _busyInvoiceIds.add(invoice.id));
    try {
      await action(orgId, invoice.id);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(successMessage)));
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(apiErrorMessage(asApiException(e), locale))),
        );
      }
    } finally {
      if (mounted) setState(() => _busyInvoiceIds.remove(invoice.id));
    }
  }

  void _handleChangeStatus(Invoice invoice, String status) {
    final t = translateWith(ref.read(localeProvider));
    _runInvoiceAction(
      invoice,
      (orgId, invoiceId) async {
        await ref.read(invoicesProvider.notifier).updateInvoiceStatus(orgId, invoiceId, status);
      },
      t('invoices.updateStatusSuccess'),
    );
  }

  void _handleReupload(Invoice invoice) {
    final t = translateWith(ref.read(localeProvider));
    _runInvoiceAction(
      invoice,
      (orgId, invoiceId) => ref.read(invoicesProvider.notifier).reuploadInvoice(orgId, invoiceId),
      t('invoices.reuploadSuccess'),
    );
  }

  void _handleSendInvoice(Invoice invoice) {
    final t = translateWith(ref.read(localeProvider));
    _runInvoiceAction(
      invoice,
      (orgId, invoiceId) => ref.read(invoicesProvider.notifier).sendInvoiceEmail(orgId, invoiceId),
      t('invoices.sendInvoiceSuccess'),
    );
  }

  void _handleDeleteInvoice(Invoice invoice) {
    final locale = ref.read(localeProvider);
    final t = translateWith(locale);

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(t('invoices.deleteInvoiceTitle')),
        content: Text(t('invoices.deleteInvoiceBody', {'number': invoice.number})),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: Text(t('common.cancel'))),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              _runInvoiceAction(
                invoice,
                (orgId, invoiceId) => ref.read(invoicesProvider.notifier).deleteInvoice(orgId, invoiceId),
                t('invoices.deleteSuccess'),
              );
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
    final invoices = ref.watch(invoicesProvider).items;

    final clientItems = [
      for (final c in clients) SelectItem(c.id, c.name),
    ];

    return Scaffold(
      appBar: AppTopBar(title: t('invoices.title')),
      body: SafeArea(
        child: Column(
          children: [
            Padding(
              padding: EdgeInsets.all(AppSpacing.of(2)),
              child: Column(
                children: [
                  AppSelectField(
                    label: t('invoices.client'),
                    value: _clientId,
                    onChanged: (v) {
                      setState(() => _clientId = v);
                      _refreshInvoices();
                    },
                    items: clientItems,
                    placeholder: t('invoices.client'),
                  ),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Expanded(
                        child: AppDateField(
                          label: t('invoices.startDate'),
                          value: _start,
                          onChanged: (v) {
                            setState(() => _start = v);
                            _refreshInvoices();
                          },
                        ),
                      ),
                      SizedBox(width: AppSpacing.of(1.5)),
                      Expanded(
                        child: AppDateField(
                          label: t('invoices.endDate'),
                          value: _end,
                          onChanged: (v) {
                            setState(() => _end = v);
                            _refreshInvoices();
                          },
                        ),
                      ),
                    ],
                  ),
                  ErrorBanner(message: _error),
                  Row(
                    children: [
                      Expanded(
                        child: AppButton(
                          title: t('invoices.preview'),
                          variant: AppButtonVariant.secondary,
                          onPressed: _clientId.isEmpty ? null : _handlePreview,
                          loading: _busy,
                        ),
                      ),
                      SizedBox(width: AppSpacing.of(1.5)),
                      Expanded(
                        child: AppButton(
                          title: t('invoices.generate'),
                          onPressed: _clientId.isEmpty ? null : _handleGenerate,
                          loading: _busy,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            Padding(
              padding: EdgeInsets.symmetric(horizontal: AppSpacing.of(2)),
              child: Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  t('invoices.previousInvoices'),
                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: AppColors.of(context).textMuted),
                ),
              ),
            ),
            Expanded(
              child: invoices.isEmpty
                  ? Padding(
                      padding: EdgeInsets.symmetric(horizontal: AppSpacing.of(2)),
                      child: Text(t('invoices.noInvoices'), style: TextStyle(color: AppColors.of(context).textMuted)),
                    )
                  : ListView.separated(
                      padding: EdgeInsets.symmetric(horizontal: AppSpacing.of(2)),
                      itemCount: invoices.length,
                      separatorBuilder: (_, _) => Divider(height: 1, color: AppColors.of(context).border),
                      itemBuilder: (context, index) {
                        final invoice = invoices[index];
                        final busy = _busyInvoiceIds.contains(invoice.id);
                        return ListTile(
                          contentPadding: EdgeInsets.zero,
                          title: Text(invoice.number),
                          subtitle: Text('${invoice.status} · ${invoice.totalTTC.toStringAsFixed(2)}'),
                          trailing: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              PopupMenuButton<String>(
                                enabled: !busy,
                                tooltip: t('invoices.changeStatus'),
                                icon: Icon(Icons.sync_alt, color: AppColors.of(context).textMuted),
                                onSelected: (status) => _handleChangeStatus(invoice, status),
                                itemBuilder: (context) => [
                                  for (final status in _invoiceStatuses)
                                    PopupMenuItem(value: status, child: Text(t(_invoiceStatusLabelKey(status)))),
                                ],
                              ),
                              IconButton(
                                onPressed: busy ? null : () => _handleReupload(invoice),
                                icon: Icon(Icons.cloud_upload_outlined, color: AppColors.of(context).textMuted),
                                tooltip: t('invoices.reupload'),
                              ),
                              IconButton(
                                onPressed: busy ? null : () => _handleSendInvoice(invoice),
                                icon: Icon(Icons.mail_outline, color: AppColors.of(context).textMuted),
                                tooltip: t('invoices.sendInvoice'),
                              ),
                              IconButton(
                                onPressed: busy ? null : () => _handleDeleteInvoice(invoice),
                                icon: Icon(Icons.delete_outline, color: AppColors.of(context).danger),
                                tooltip: t('common.delete'),
                              ),
                            ],
                          ),
                          onTap: () => _handleOpenInvoice(invoice),
                        );
                      },
                    ),
            ),
          ],
        ),
      ),
    );
  }
}
