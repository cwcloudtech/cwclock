class Currency {
  final String iso;
  final String name;

  const Currency({required this.iso, required this.name});

  factory Currency.fromJson(Map<String, dynamic> json) {
    return Currency(
      iso: json['iso'] as String? ?? '',
      name: json['name'] as String? ?? '',
    );
  }
}
