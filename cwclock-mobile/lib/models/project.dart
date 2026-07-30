class Project {
  final String id;
  final String name;
  final String clientId;
  final String? color;
  final num? dailyRate;
  final List<String> subdivisions;

  const Project({
    required this.id,
    required this.name,
    required this.clientId,
    this.color,
    this.dailyRate,
    this.subdivisions = const [],
  });

  factory Project.fromJson(Map<String, dynamic> json) {
    return Project(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
      clientId: json['clientId'] as String? ?? '',
      color: json['color'] as String?,
      dailyRate: json['dailyRate'] as num?,
      subdivisions: (json['subdivisions'] as List?)?.map((e) => e.toString()).toList() ?? const [],
    );
  }
}

const defaultProjectColor = '#1cb9f7';

const presetProjectColors = [
  '#1cb9f7',
  '#f76c1c',
  '#1cf7a0',
  '#f71c6c',
  '#7c1cf7',
  '#f7d11c',
  '#1c3ff7',
  '#8a8a8a',
];

String subdivisionsToText(List<String> subdivisions) => subdivisions.join(', ');

List<String> textToSubdivisions(String text) => text
    .split(',')
    .map((s) => s.trim())
    .where((s) => s.isNotEmpty)
    .toList();
