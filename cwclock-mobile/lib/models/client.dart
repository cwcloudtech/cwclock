class Client {
  final String id;
  final String name;
  final String? email;
  final String? contactName;
  final String? address;
  final String? postalCode;
  final String? city;
  final String? country;
  final String? vatNumber;
  final num? vatRate;
  final String? identificationNumber;
  final String? purchaseOrder;
  final num? hoursPerDay;
  final num? dailyRate;
  final bool sendReportsWithInvoice;

  const Client({
    required this.id,
    required this.name,
    this.email,
    this.contactName,
    this.address,
    this.postalCode,
    this.city,
    this.country,
    this.vatNumber,
    this.vatRate,
    this.identificationNumber,
    this.purchaseOrder,
    this.hoursPerDay,
    this.dailyRate,
    this.sendReportsWithInvoice = false,
  });

  factory Client.fromJson(Map<String, dynamic> json) {
    return Client(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
      email: json['email'] as String?,
      contactName: json['contactName'] as String?,
      address: json['address'] as String?,
      postalCode: json['postalCode'] as String?,
      city: json['city'] as String?,
      country: json['country'] as String?,
      vatNumber: json['vatNumber'] as String?,
      vatRate: json['vatRate'] as num?,
      identificationNumber: json['identificationNumber'] as String?,
      purchaseOrder: json['purchaseOrder'] as String?,
      hoursPerDay: json['hoursPerDay'] as num?,
      dailyRate: json['dailyRate'] as num?,
      sendReportsWithInvoice: json['sendReportsWithInvoice'] as bool? ?? false,
    );
  }
}

class ClientFields {
  final String name;
  final String email;
  final String contactName;
  final String address;
  final String postalCode;
  final String city;
  final String country;
  final String vatNumber;
  final String vatRate;
  final String identificationNumber;
  final String purchaseOrder;
  final String hoursPerDay;
  final String dailyRate;
  final bool sendReportsWithInvoice;

  const ClientFields({
    this.name = '',
    this.email = '',
    this.contactName = '',
    this.address = '',
    this.postalCode = '',
    this.city = '',
    this.country = '',
    this.vatNumber = '',
    this.vatRate = '',
    this.identificationNumber = '',
    this.purchaseOrder = '',
    this.hoursPerDay = '',
    this.dailyRate = '',
    this.sendReportsWithInvoice = false,
  });

  factory ClientFields.fromClient(Client c) {
    return ClientFields(
      name: c.name,
      email: c.email ?? '',
      contactName: c.contactName ?? '',
      address: c.address ?? '',
      postalCode: c.postalCode ?? '',
      city: c.city ?? '',
      country: c.country ?? '',
      vatNumber: c.vatNumber ?? '',
      vatRate: c.vatRate != null ? c.vatRate.toString() : '',
      identificationNumber: c.identificationNumber ?? '',
      purchaseOrder: c.purchaseOrder ?? '',
      hoursPerDay: c.hoursPerDay != null ? c.hoursPerDay.toString() : '',
      dailyRate: c.dailyRate != null ? c.dailyRate.toString() : '',
      sendReportsWithInvoice: c.sendReportsWithInvoice,
    );
  }

  ClientFields copyWith({
    String? name,
    String? email,
    String? contactName,
    String? address,
    String? postalCode,
    String? city,
    String? country,
    String? vatNumber,
    String? vatRate,
    String? identificationNumber,
    String? purchaseOrder,
    String? hoursPerDay,
    String? dailyRate,
    bool? sendReportsWithInvoice,
  }) {
    return ClientFields(
      name: name ?? this.name,
      email: email ?? this.email,
      contactName: contactName ?? this.contactName,
      address: address ?? this.address,
      postalCode: postalCode ?? this.postalCode,
      city: city ?? this.city,
      country: country ?? this.country,
      vatNumber: vatNumber ?? this.vatNumber,
      vatRate: vatRate ?? this.vatRate,
      identificationNumber: identificationNumber ?? this.identificationNumber,
      purchaseOrder: purchaseOrder ?? this.purchaseOrder,
      hoursPerDay: hoursPerDay ?? this.hoursPerDay,
      dailyRate: dailyRate ?? this.dailyRate,
      sendReportsWithInvoice: sendReportsWithInvoice ?? this.sendReportsWithInvoice,
    );
  }

  Map<String, dynamic> toJson() {
    num? toNumber(String v) => v.isEmpty ? null : num.tryParse(v);
    return {
      'name': name,
      'email': email,
      'contactName': contactName,
      'address': address,
      'postalCode': postalCode,
      'city': city,
      'country': country,
      'vatNumber': vatNumber,
      'vatRate': toNumber(vatRate),
      'identificationNumber': identificationNumber,
      'purchaseOrder': purchaseOrder,
      'hoursPerDay': toNumber(hoursPerDay),
      'dailyRate': toNumber(dailyRate),
      'sendReportsWithInvoice': sendReportsWithInvoice,
    };
  }
}
