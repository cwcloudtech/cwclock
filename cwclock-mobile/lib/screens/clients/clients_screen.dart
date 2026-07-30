import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../providers/clients_provider.dart';
import '../../providers/locale_provider.dart';
import '../../providers/session_provider.dart';
import '../../theme.dart';

/// Ported from src/screens/clients/ClientsScreen.js.
class ClientsScreen extends ConsumerStatefulWidget {
  const ClientsScreen({super.key});

  @override
  ConsumerState<ClientsScreen> createState() => _ClientsScreenState();
}

class _ClientsScreenState extends ConsumerState<ClientsScreen> {
  @override
  void initState() {
    super.initState();
    final orgId = ref.read(sessionProvider).orgId;
    if (orgId != null) {
      Future.microtask(() => ref.read(clientsProvider.notifier).listClients(orgId));
    }
  }

  @override
  Widget build(BuildContext context) {
    final locale = ref.watch(localeProvider);
    final t = translateWith(locale);
    final clients = ref.watch(clientsProvider).items;

    return Scaffold(
      appBar: AppBar(title: Text(t('clients.title'))),
      body: SafeArea(
        child: clients.isEmpty
            ? Padding(
                padding: EdgeInsets.all(AppSpacing.of(2)),
                child: Text(t('clients.noClients'), style: TextStyle(color: AppColors.of(context).textMuted)),
              )
            : ListView.separated(
                padding: EdgeInsets.symmetric(horizontal: AppSpacing.of(2)),
                itemCount: clients.length,
                separatorBuilder: (_, _) => Divider(height: 1, color: AppColors.of(context).border),
                itemBuilder: (context, index) {
                  final client = clients[index];
                  return ListTile(
                    contentPadding: EdgeInsets.zero,
                    title: Text(client.name),
                    subtitle: (client.email?.isNotEmpty ?? false) ? Text(client.email!) : null,
                    onTap: () => context.push('/clients/form', extra: client),
                  );
                },
              ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => context.push('/clients/form'),
        backgroundColor: AppColors.of(context).primary,
        child: const Icon(Icons.add, color: kWhite),
      ),
    );
  }
}
