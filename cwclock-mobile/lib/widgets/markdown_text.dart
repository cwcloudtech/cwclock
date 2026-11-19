import 'package:flutter/material.dart';

import '../theme.dart';

/// A small, self-contained Markdown renderer for assistant answers
/// (ai-instruct-132) — headings, bold/italic, inline code, code fences, bullet
/// and numbered lists, blockquotes, horizontal rules, links and GitHub-style
/// pipe tables. Hand-rolled (like strong-fish's markdown_text) rather than
/// pulling a package in, so the chat renders the same shapes the web app does
/// (react-markdown + remark-gfm) with no extra dependency. Links are shown
/// styled but are not tappable (there's no url_launcher in the app).
class MarkdownText extends StatelessWidget {
  final String data;
  final Color? color;

  const MarkdownText(this.data, {super.key, this.color});

  static final RegExp _inlineRe = RegExp(
    r'(\*\*(.+?)\*\*)' // 1,2 bold
    r'|(__(.+?)__)' // 3,4 bold
    r'|(`([^`]+)`)' // 5,6 inline code
    r'|(\[([^\]]+)\]\(([^)\s]+)\))' // 7 outer, 8 text, 9 url
    r'|(\*(.+?)\*)' // 10,11 italic
    r'|(_(.+?)_)', // 12,13 italic
  );

  bool _isTableSeparator(String line) {
    return RegExp(r'^\s*\|?\s*:?-{1,}:?\s*(\|\s*:?-{1,}:?\s*)+\|?\s*$').hasMatch(line);
  }

  List<String> _cells(String line) {
    var l = line.trim();
    if (l.startsWith('|')) l = l.substring(1);
    if (l.endsWith('|')) l = l.substring(0, l.length - 1);
    return l.split('|').map((c) => c.trim()).toList();
  }

  List<InlineSpan> _inline(BuildContext context, String text) {
    final colors = AppColors.of(context);
    final spans = <InlineSpan>[];
    var last = 0;
    for (final m in _inlineRe.allMatches(text)) {
      if (m.start > last) spans.add(TextSpan(text: text.substring(last, m.start)));
      if (m.group(1) != null) {
        spans.add(TextSpan(text: m.group(2), style: const TextStyle(fontWeight: FontWeight.bold)));
      } else if (m.group(3) != null) {
        spans.add(TextSpan(text: m.group(4), style: const TextStyle(fontWeight: FontWeight.bold)));
      } else if (m.group(5) != null) {
        spans.add(TextSpan(
          text: m.group(6),
          style: TextStyle(fontFamily: 'monospace', backgroundColor: colors.backgroundMuted),
        ));
      } else if (m.group(7) != null) {
        spans.add(TextSpan(
          text: m.group(8),
          style: TextStyle(color: colors.primary, decoration: TextDecoration.underline),
        ));
      } else if (m.group(10) != null) {
        spans.add(TextSpan(text: m.group(11), style: const TextStyle(fontStyle: FontStyle.italic)));
      } else if (m.group(12) != null) {
        spans.add(TextSpan(text: m.group(13), style: const TextStyle(fontStyle: FontStyle.italic)));
      }
      last = m.end;
    }
    if (last < text.length) spans.add(TextSpan(text: text.substring(last)));
    return spans;
  }

  Widget _paragraph(BuildContext context, String text, TextStyle base) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Text.rich(TextSpan(style: base, children: _inline(context, text))),
    );
  }

  Widget _heading(BuildContext context, int level, String text, TextStyle base) {
    final size = level == 1
        ? 20.0
        : level == 2
            ? 18.0
            : 16.0;
    return Padding(
      padding: const EdgeInsets.only(top: 6, bottom: 6),
      child: Text.rich(TextSpan(
        style: base.copyWith(fontSize: size, fontWeight: FontWeight.w700),
        children: _inline(context, text),
      )),
    );
  }

  Widget _list(BuildContext context, List<_ListEntry> items, TextStyle base) {
    var ordinal = 0;
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          for (final item in items)
            Padding(
              padding: const EdgeInsets.only(bottom: 3),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Padding(
                    padding: const EdgeInsets.only(right: 8, top: 1),
                    child: Text(item.ordered ? '${++ordinal}.' : '•', style: base),
                  ),
                  Expanded(
                    child: Text.rich(TextSpan(style: base, children: _inline(context, item.text))),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }

  Widget _codeBlock(BuildContext context, String code, TextStyle base) {
    final colors = AppColors.of(context);
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: colors.backgroundMuted,
        borderRadius: BorderRadius.circular(AppRadius.value),
      ),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Text(code, style: base.copyWith(fontFamily: 'monospace', fontSize: 13)),
      ),
    );
  }

  Widget _blockquote(BuildContext context, String text, TextStyle base) {
    final colors = AppColors.of(context);
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.only(left: 10),
      decoration: BoxDecoration(
        border: Border(left: BorderSide(color: colors.border, width: 3)),
      ),
      child: Text.rich(TextSpan(
        style: base.copyWith(color: colors.textMuted),
        children: _inline(context, text),
      )),
    );
  }

  Widget _table(BuildContext context, List<String> lines, TextStyle base) {
    final colors = AppColors.of(context);
    final header = _cells(lines[0]);
    final body = lines.skip(2).map(_cells).toList();
    final columns = header.length;

    TableRow row(List<String> cells, {required bool head}) {
      return TableRow(
        decoration: head ? BoxDecoration(color: colors.backgroundMuted) : null,
        children: [
          for (var c = 0; c < columns; c++)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
              child: Text.rich(TextSpan(
                style: head ? base.copyWith(fontWeight: FontWeight.w700) : base,
                children: _inline(context, c < cells.length ? cells[c] : ''),
              )),
            ),
        ],
      );
    }

    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Table(
          defaultColumnWidth: const IntrinsicColumnWidth(),
          border: TableBorder.all(color: colors.border),
          children: [
            row(header, head: true),
            for (final r in body) row(r, head: false),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final colors = AppColors.of(context);
    final base = TextStyle(fontSize: 15, height: 1.4, color: color ?? colors.text);
    final lines = data.replaceAll('\r\n', '\n').split('\n');
    final widgets = <Widget>[];
    var i = 0;

    bool isHeading(String l) => RegExp(r'^#{1,4}\s').hasMatch(l.trim());
    bool isHr(String l) => RegExp(r'^(-{3,}|\*{3,}|_{3,})$').hasMatch(l.trim());
    bool isList(String l) => RegExp(r'^\s*([-*+]|\d+\.)\s+').hasMatch(l);
    bool isTableStart(int idx) =>
        lines[idx].contains('|') && idx + 1 < lines.length && _isTableSeparator(lines[idx + 1]);

    while (i < lines.length) {
      final line = lines[i];
      final trimmed = line.trim();
      if (trimmed.isEmpty) {
        i++;
        continue;
      }

      // fenced code
      if (trimmed.startsWith('```')) {
        final buf = <String>[];
        i++;
        while (i < lines.length && !lines[i].trim().startsWith('```')) {
          buf.add(lines[i]);
          i++;
        }
        if (i < lines.length) i++; // closing fence
        widgets.add(_codeBlock(context, buf.join('\n'), base));
        continue;
      }

      // table
      if (isTableStart(i)) {
        final tableLines = <String>[lines[i], lines[i + 1]];
        i += 2;
        while (i < lines.length && lines[i].contains('|') && lines[i].trim().isNotEmpty) {
          tableLines.add(lines[i]);
          i++;
        }
        widgets.add(_table(context, tableLines, base));
        continue;
      }

      // heading
      final h = RegExp(r'^(#{1,4})\s+(.*)$').firstMatch(trimmed);
      if (h != null) {
        widgets.add(_heading(context, h.group(1)!.length, h.group(2)!, base));
        i++;
        continue;
      }

      // horizontal rule
      if (isHr(line)) {
        widgets.add(Divider(color: colors.border, height: 16));
        i++;
        continue;
      }

      // blockquote
      if (trimmed.startsWith('>')) {
        final buf = <String>[];
        while (i < lines.length && lines[i].trim().startsWith('>')) {
          buf.add(lines[i].trim().replaceFirst(RegExp(r'^>\s?'), ''));
          i++;
        }
        widgets.add(_blockquote(context, buf.join(' '), base));
        continue;
      }

      // list
      if (isList(line)) {
        final items = <_ListEntry>[];
        final itemRe = RegExp(r'^\s*([-*+]|\d+\.)\s+(.*)$');
        while (i < lines.length && isList(lines[i])) {
          final m = itemRe.firstMatch(lines[i])!;
          items.add(_ListEntry(RegExp(r'^\d+\.$').hasMatch(m.group(1)!), m.group(2)!));
          i++;
        }
        widgets.add(_list(context, items, base));
        continue;
      }

      // paragraph
      final buf = <String>[];
      while (i < lines.length &&
          lines[i].trim().isNotEmpty &&
          !lines[i].trim().startsWith('```') &&
          !isTableStart(i) &&
          !isHeading(lines[i]) &&
          !isHr(lines[i]) &&
          !lines[i].trim().startsWith('>') &&
          !isList(lines[i])) {
        buf.add(lines[i].trim());
        i++;
      }
      if (buf.isNotEmpty) widgets.add(_paragraph(context, buf.join(' '), base));
    }

    return Column(crossAxisAlignment: CrossAxisAlignment.start, children: widgets);
  }
}

class _ListEntry {
  final bool ordered;
  final String text;
  const _ListEntry(this.ordered, this.text);
}
