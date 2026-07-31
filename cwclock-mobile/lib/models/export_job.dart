class ExportJob {
  final String id;
  final String name;
  final String cronExpression;
  final List<String> reportTypes;
  final bool enabled;

  const ExportJob({
    required this.id,
    required this.name,
    required this.cronExpression,
    required this.reportTypes,
    required this.enabled,
  });

  factory ExportJob.fromJson(Map<String, dynamic> json) {
    return ExportJob(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
      cronExpression: json['cronExpression'] as String? ?? '',
      reportTypes: (json['reportTypes'] as List?)?.map((e) => e as String).toList() ?? const [],
      enabled: json['enabled'] as bool? ?? false,
    );
  }
}
