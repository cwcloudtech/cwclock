import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../i18n/app_localizations.dart';
import '../../providers/ai_provider.dart';
import '../../providers/locale_provider.dart';
import '../../providers/session_provider.dart';
import '../../theme.dart';
import '../../widgets/app_top_bar.dart';
import '../../widgets/markdown_text.dart';

class _Msg {
  final String role; // user | assistant | error
  final String content;
  const _Msg(this.role, this.content);
}

/// The assistant chat (ai-instruct-132) — the mobile counterpart of cwclock-ui's
/// AiAssistant page. Answers render as Markdown (headings, lists, tables, …).
class AiChatScreen extends ConsumerStatefulWidget {
  const AiChatScreen({super.key});

  @override
  ConsumerState<AiChatScreen> createState() => _AiChatScreenState();
}

class _AiChatScreenState extends ConsumerState<AiChatScreen> {
  final List<_Msg> _messages = [];
  final TextEditingController _input = TextEditingController();
  final ScrollController _scroll = ScrollController();
  String? _conversationId;
  String _connectionId = '';

  @override
  void initState() {
    super.initState();
    Future.microtask(() async {
      try {
        final connections = await ref.read(aiProvider.notifier).listConnections();
        if (connections.isNotEmpty && _connectionId.isEmpty) {
          setState(() => _connectionId = connections.first.id);
        }
        await ref.read(aiProvider.notifier).listConversations();
      } catch (_) {
        /* best-effort */
      }
    });
  }

  @override
  void dispose() {
    _input.dispose();
    _scroll.dispose();
    super.dispose();
  }

  String Function(String, [Map<String, String>?]) get _t => translateWith(ref.read(localeProvider));

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scroll.hasClients) {
        _scroll.animateTo(_scroll.position.maxScrollExtent,
            duration: const Duration(milliseconds: 200), curve: Curves.easeOut);
      }
    });
  }

  void _newChat() {
    setState(() {
      _messages.clear();
      _conversationId = null;
    });
  }

  Future<void> _openConversation(String id) async {
    try {
      final conv = await ref.read(aiProvider.notifier).getConversation(id);
      setState(() {
        _messages
          ..clear()
          ..addAll([
            for (final ex in conv.exchanges) ...[
              _Msg('user', ex.question),
              _Msg('assistant', ex.answer),
            ],
          ]);
        _conversationId = conv.id;
      });
      _scrollToBottom();
    } catch (e) {
      _snack(apiErrorMessage(e, ref.read(localeProvider)));
    }
  }

  void _snack(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(message)));
  }

  Future<void> _send() async {
    final text = _input.text.trim();
    if (text.isEmpty || ref.read(aiProvider).isSending) return;
    _input.clear();
    setState(() => _messages.add(_Msg('user', text)));
    _scrollToBottom();
    final orgId = ref.read(sessionProvider).orgId;
    try {
      final res = await ref.read(aiProvider.notifier).sendChat(
            message: text,
            conversationId: _conversationId,
            connectionId: _connectionId,
            orgId: orgId,
          );
      setState(() {
        _messages.add(_Msg('assistant', res['message'] as String? ?? ''));
        final convId = res['conversationId'] as String?;
        if (convId != null && convId.isNotEmpty) _conversationId = convId;
      });
      _scrollToBottom();
      await ref.read(aiProvider.notifier).listConversations();
    } catch (e) {
      final message = apiErrorMessage(e, ref.read(localeProvider));
      setState(() => _messages.add(_Msg('error', message)));
      _scrollToBottom();
    }
  }

  Future<void> _openHistory() async {
    final t = _t;
    final conversations = ref.read(aiProvider).conversations;
    await showModalBottomSheet<void>(
      context: context,
      showDragHandle: true,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.add),
              title: Text(t('ai.newChat')),
              onTap: () {
                Navigator.pop(context);
                _newChat();
              },
            ),
            const Divider(height: 1),
            if (conversations.isEmpty)
              Padding(
                padding: const EdgeInsets.all(16),
                child: Text(t('ai.noConversations'),
                    style: TextStyle(color: AppColors.of(context).textMuted)),
              )
            else
              Flexible(
                child: ListView(
                  shrinkWrap: true,
                  children: [
                    for (final c in conversations)
                      ListTile(
                        leading: const Icon(Icons.chat_bubble_outline),
                        title: Text(c.title, maxLines: 1, overflow: TextOverflow.ellipsis),
                        selected: c.id == _conversationId,
                        onTap: () {
                          Navigator.pop(context);
                          _openConversation(c.id);
                        },
                      ),
                  ],
                ),
              ),
            const Divider(height: 1),
            if (conversations.isNotEmpty)
              ListTile(
                leading: Icon(Icons.delete_outline, color: AppColors.of(context).danger),
                title: Text(t('ai.clearHistory'), style: TextStyle(color: AppColors.of(context).danger)),
                onTap: () async {
                  Navigator.pop(context);
                  await _clearHistory();
                },
              ),
          ],
        ),
      ),
    );
  }

  Future<void> _clearHistory() async {
    final t = _t;
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(t('ai.clearHistoryTitle')),
        content: Text(t('ai.clearHistoryBody')),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: Text(t('common.cancel'))),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(t('ai.clearHistory'), style: TextStyle(color: AppColors.of(context).danger)),
          ),
        ],
      ),
    );
    if (ok == true) {
      try {
        await ref.read(aiProvider.notifier).clearHistory();
        _newChat();
      } catch (e) {
        _snack(apiErrorMessage(e, ref.read(localeProvider)));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final t = translateWith(ref.watch(localeProvider));
    final ai = ref.watch(aiProvider);
    final colors = AppColors.of(context);
    final connections = ai.connections;

    return Scaffold(
      appBar: AppTopBar(
        title: t('ai.chatTitle'),
        actions: [
          if (connections.length > 1)
            PopupMenuButton<String>(
              icon: const Icon(Icons.psychology_outlined),
              tooltip: t('ai.model'),
              initialValue: _connectionId,
              onSelected: (v) => setState(() => _connectionId = v),
              itemBuilder: (context) => [
                for (final c in connections)
                  PopupMenuItem(value: c.id, child: Text(c.name.isNotEmpty ? c.name : c.provider)),
              ],
            ),
          IconButton(
            tooltip: t('ai.newChat'),
            icon: const Icon(Icons.add_comment_outlined),
            onPressed: _newChat,
          ),
          IconButton(
            tooltip: t('ai.history'),
            icon: const Icon(Icons.history),
            onPressed: _openHistory,
          ),
        ],
      ),
      backgroundColor: colors.background,
      body: SafeArea(
        child: Column(
          children: [
            Expanded(
              child: _messages.isEmpty && !ai.isSending
                  ? _welcome(context, t)
                  : ListView.builder(
                      controller: _scroll,
                      padding: EdgeInsets.all(AppSpacing.of(2)),
                      itemCount: _messages.length + (ai.isSending ? 1 : 0),
                      itemBuilder: (context, i) {
                        if (i >= _messages.length) return _typing(context);
                        return _bubble(context, _messages[i]);
                      },
                    ),
            ),
            _composer(context, t, ai.isSending),
          ],
        ),
      ),
    );
  }

  Widget _welcome(BuildContext context, String Function(String, [Map<String, String>?]) t) {
    final colors = AppColors.of(context);
    return Center(
      child: Padding(
        padding: EdgeInsets.all(AppSpacing.of(3)),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.auto_awesome, size: 40, color: colors.primary),
            SizedBox(height: AppSpacing.of(1.5)),
            Text(t('ai.welcomeTitle'),
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600, color: colors.text)),
            SizedBox(height: AppSpacing.of(1)),
            Text(t('ai.welcomeBody'),
                textAlign: TextAlign.center, style: TextStyle(color: colors.textMuted)),
          ],
        ),
      ),
    );
  }

  Widget _bubble(BuildContext context, _Msg m) {
    final colors = AppColors.of(context);
    if (m.role == 'user') {
      return Align(
        alignment: Alignment.centerRight,
        child: Container(
          margin: const EdgeInsets.only(bottom: 10, left: 40),
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
          decoration: BoxDecoration(
            color: colors.primary,
            borderRadius: BorderRadius.circular(AppRadius.value),
          ),
          child: Text(m.content, style: const TextStyle(color: kWhite)),
        ),
      );
    }
    final isError = m.role == 'error';
    return Align(
      alignment: Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.only(bottom: 10, right: 24),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        decoration: BoxDecoration(
          color: isError ? colors.danger.withValues(alpha: 0.12) : colors.backgroundMuted,
          borderRadius: BorderRadius.circular(AppRadius.value),
          border: isError ? Border.all(color: colors.danger) : null,
        ),
        child: isError
            ? Text(m.content, style: TextStyle(color: colors.danger))
            : MarkdownText(m.content, color: colors.text),
      ),
    );
  }

  Widget _typing(BuildContext context) {
    final colors = AppColors.of(context);
    return Align(
      alignment: Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.only(bottom: 10),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
        decoration: BoxDecoration(
          color: colors.backgroundMuted,
          borderRadius: BorderRadius.circular(AppRadius.value),
        ),
        child: SizedBox(
          width: 20,
          height: 20,
          child: CircularProgressIndicator(strokeWidth: 2, color: colors.textMuted),
        ),
      ),
    );
  }

  Widget _composer(BuildContext context, String Function(String, [Map<String, String>?]) t, bool sending) {
    final colors = AppColors.of(context);
    return Container(
      padding: EdgeInsets.all(AppSpacing.of(1)),
      decoration: BoxDecoration(
        color: colors.background,
        border: Border(top: BorderSide(color: colors.border)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          Expanded(
            child: TextField(
              controller: _input,
              minLines: 1,
              maxLines: 5,
              textCapitalization: TextCapitalization.sentences,
              decoration: InputDecoration(
                hintText: t('ai.askPlaceholder'),
                contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              ),
              onSubmitted: (_) => _send(),
            ),
          ),
          const SizedBox(width: 8),
          Material(
            color: colors.primary,
            shape: const CircleBorder(),
            child: InkWell(
              customBorder: const CircleBorder(),
              onTap: sending ? null : _send,
              child: Padding(
                padding: const EdgeInsets.all(10),
                child: Icon(sending ? Icons.hourglass_empty : Icons.send, color: kWhite, size: 20),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
