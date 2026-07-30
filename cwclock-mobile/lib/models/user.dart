class User {
  final String id;
  final String email;
  final String role;
  final String? name;
  final String? surname;
  final String? picture;
  final double pictureX;
  final double pictureY;

  const User({
    required this.id,
    required this.email,
    required this.role,
    this.name,
    this.surname,
    this.picture,
    this.pictureX = 50,
    this.pictureY = 50,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] as String,
      email: json['email'] as String? ?? '',
      role: json['role'] as String? ?? '',
      name: json['name'] as String?,
      surname: json['surname'] as String?,
      picture: json['picture'] as String?,
      pictureX: (json['pictureX'] as num?)?.toDouble() ?? 50,
      pictureY: (json['pictureY'] as num?)?.toDouble() ?? 50,
    );
  }
}
