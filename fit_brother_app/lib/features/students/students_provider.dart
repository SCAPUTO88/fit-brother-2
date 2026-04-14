import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../auth/auth_provider.dart';

// ---------- Model ----------

class StudentSummary {
  final String id;
  final String fullName;
  final String email;

  const StudentSummary({
    required this.id,
    required this.fullName,
    required this.email,
  });

  factory StudentSummary.fromJson(Map<String, dynamic> json) => StudentSummary(
        id: json['id'] as String,
        fullName: json['full_name'] as String,
        email: json['email'] as String,
      );
}

// ---------- Provider ----------

final studentsProvider = FutureProvider<List<StudentSummary>>((ref) async {
  final dioClient = ref.watch(dioClientProvider);
  final response = await dioClient.dio.get('/students');
  final List data = response.data as List;
  return data.map((e) => StudentSummary.fromJson(e as Map<String, dynamic>)).toList();
});
