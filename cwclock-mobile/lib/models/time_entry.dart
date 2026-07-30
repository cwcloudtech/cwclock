class TimeEntry {
  final String id;
  final String text;
  final String day;
  final String? start;
  final String? end;
  final bool allDay;
  final bool half;
  final String clientId;
  final String projectId;

  const TimeEntry({
    required this.id,
    required this.text,
    required this.day,
    this.start,
    this.end,
    this.allDay = false,
    this.half = false,
    required this.clientId,
    required this.projectId,
  });

  factory TimeEntry.fromJson(Map<String, dynamic> json) {
    return TimeEntry(
      id: json['id'] as String,
      text: json['text'] as String? ?? '',
      day: json['day'] as String? ?? '',
      start: json['start'] as String?,
      end: json['end'] as String?,
      allDay: json['allDay'] as bool? ?? false,
      half: json['half'] as bool? ?? false,
      clientId: json['clientId'] as String? ?? '',
      projectId: json['projectId'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'text': text,
      'day': day,
      'start': start,
      'end': end,
      'allDay': allDay,
      'half': half,
      'clientId': clientId,
      'projectId': projectId,
    };
  }

  TimeEntry copyWith({
    String? text,
    String? day,
    String? start,
    String? end,
    bool? allDay,
    bool? half,
    String? clientId,
    String? projectId,
  }) {
    return TimeEntry(
      id: id,
      text: text ?? this.text,
      day: day ?? this.day,
      start: start ?? this.start,
      end: end ?? this.end,
      allDay: allDay ?? this.allDay,
      half: half ?? this.half,
      clientId: clientId ?? this.clientId,
      projectId: projectId ?? this.projectId,
    );
  }
}
