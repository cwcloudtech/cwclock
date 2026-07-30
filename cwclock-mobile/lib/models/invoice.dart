class Invoice {
  final String id;
  final String number;
  final String status;
  final num totalTTC;

  const Invoice({
    required this.id,
    required this.number,
    required this.status,
    required this.totalTTC,
  });

  factory Invoice.fromJson(Map<String, dynamic> json) {
    return Invoice(
      id: json['id'] as String,
      number: json['number'] as String? ?? '',
      status: json['status'] as String? ?? '',
      totalTTC: json['totalTTC'] as num? ?? 0,
    );
  }
}
