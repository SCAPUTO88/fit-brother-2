import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../auth/auth_provider.dart';

// ---------- Models ----------

class WorkoutPlanSummary {
  final String id;
  final String title;
  final String studentId;
  final String description;
  final bool isActive;

  const WorkoutPlanSummary({
    required this.id,
    required this.title,
    required this.studentId,
    required this.description,
    required this.isActive,
  });

  factory WorkoutPlanSummary.fromJson(Map<String, dynamic> json) => WorkoutPlanSummary(
        id: json['id'] as String,
        title: json['title'] as String,
        studentId: json['student_id'] as String,
        description: json['description'] as String? ?? '',
        isActive: json['is_active'] as bool,
      );
}

// ---------- Provider ----------

final workoutPlansProvider = FutureProvider<List<WorkoutPlanSummary>>((ref) async {
  final dioClient = ref.watch(dioClientProvider);
  final response = await dioClient.dio.get('/workouts');
  final List data = response.data as List;
  return data.map((e) => WorkoutPlanSummary.fromJson(e as Map<String, dynamic>)).toList();
});
