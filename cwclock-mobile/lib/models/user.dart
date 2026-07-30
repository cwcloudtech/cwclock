class User {
  final String id;
  final String email;
  final String role;
  final String? name;
  final String? surname;

  const User({
    required this.id,
    required this.email,
    required this.role,
    this.name,
    this.surname,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] as String,
      email: json['email'] as String? ?? '',
      role: json['role'] as String? ?? '',
      name: json['name'] as String?,
      surname: json['surname'] as String?,
    );
  }
}
