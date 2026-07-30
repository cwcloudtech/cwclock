class Country {
  final String iso;
  final String name;
  final String? currency;

  const Country({required this.iso, required this.name, this.currency});

  factory Country.fromJson(Map<String, dynamic> json) {
    return Country(
      iso: json['iso'] as String? ?? '',
      name: json['name'] as String? ?? '',
      currency: json['currency'] as String?,
    );
  }
}
