class Member {
  final String userId;
  final String role;
  final String email;
  final String? name;
  final String? surname;

  const Member({
    required this.userId,
    required this.role,
    required this.email,
    this.name,
    this.surname,
  });

  factory Member.fromJson(Map<String, dynamic> json) {
    return Member(
      userId: json['userId'] as String,
      role: json['role'] as String? ?? '',
      email: json['email'] as String? ?? '',
      name: json['name'] as String?,
      surname: json['surname'] as String?,
    );
  }
}
