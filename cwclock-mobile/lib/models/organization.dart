class Organization {
  final String id;
  final String name;
  final String? ownerId;
  final String? accountingEmail;
  final String? address;
  final String? postalCode;
  final String? city;
  final String? country;
  final String? currency;
  final String? vatNumber;
  final String? identificationNumber;
  final String? iban;
  final String? bic;

  const Organization({
    required this.id,
    required this.name,
    this.ownerId,
    this.accountingEmail,
    this.address,
    this.postalCode,
    this.city,
    this.country,
    this.currency,
    this.vatNumber,
    this.identificationNumber,
    this.iban,
    this.bic,
  });

  factory Organization.fromJson(Map<String, dynamic> json) {
    return Organization(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
      ownerId: json['ownerId'] as String?,
      accountingEmail: json['accountingEmail'] as String?,
      address: json['address'] as String?,
      postalCode: json['postalCode'] as String?,
      city: json['city'] as String?,
      country: json['country'] as String?,
      currency: json['currency'] as String?,
      vatNumber: json['vatNumber'] as String?,
      identificationNumber: json['identificationNumber'] as String?,
      iban: json['iban'] as String?,
      bic: json['bic'] as String?,
    );
  }

  Map<String, dynamic> toUpdateJson({
    required String name,
    required String accountingEmail,
    required String address,
    required String postalCode,
    required String city,
    required String country,
    required String currency,
    required String vatNumber,
    required String identificationNumber,
    required String iban,
    required String bic,
  }) {
    return {
      'name': name,
      'accountingEmail': accountingEmail,
      'address': address,
      'postalCode': postalCode,
      'city': city,
      'country': country,
      'currency': currency,
      'vatNumber': vatNumber,
      'identificationNumber': identificationNumber,
      'iban': iban,
      'bic': bic,
    };
  }
}
