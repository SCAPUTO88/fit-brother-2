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

class ExerciseSummary {
  final String id;
  final String name;
  final int sets;
  final String reps;
  final int restSeconds;
  final String notes;
  final String videoUrl;

  const ExerciseSummary({
    required this.id,
    required this.name,
    required this.sets,
    required this.reps,
    required this.restSeconds,
    required this.notes,
    required this.videoUrl,
  });

  factory ExerciseSummary.fromJson(Map<String, dynamic> json) => ExerciseSummary(
        id: json['id'] as String,
        name: json['name'] as String,
        sets: json['sets'] as int,
        reps: json['reps'] as String? ?? '',
        restSeconds: json['rest_seconds'] as int,
        notes: json['notes'] as String? ?? '',
        videoUrl: json['video_url'] as String? ?? '',
      );
}

class SessionDetails {
  final String id;
  final String title;
  final int dayOfWeek;
  final String notes;
  final List<ExerciseSummary> exercises;

  const SessionDetails({
    required this.id,
    required this.title,
    required this.dayOfWeek,
    required this.notes,
    required this.exercises,
  });

  factory SessionDetails.fromJson(Map<String, dynamic> json) {
    final sessionMap = json['session'] as Map<String, dynamic>;
    final exercisesList = json['exercises'] as List? ?? [];
    
    return SessionDetails(
      id: sessionMap['id'] as String,
      title: sessionMap['title'] as String,
      dayOfWeek: sessionMap['day_of_week'] as int,
      notes: sessionMap['notes'] as String? ?? '',
      exercises: exercisesList.map((e) => ExerciseSummary.fromJson(e as Map<String, dynamic>)).toList(),
    );
  }
}

class PlanDetails {
  final WorkoutPlanSummary plan;
  final List<SessionDetails> sessions;

  const PlanDetails({required this.plan, required this.sessions});

  factory PlanDetails.fromJson(Map<String, dynamic> json) {
    final planMap = json['plan'] as Map<String, dynamic>;
    final sessionsList = json['sessions'] as List? ?? [];
    
    return PlanDetails(
      plan: WorkoutPlanSummary.fromJson(planMap),
      sessions: sessionsList.map((e) => SessionDetails.fromJson(e as Map<String, dynamic>)).toList(),
    );
  }
}

// ---------- Provider ----------

final workoutPlansProvider = FutureProvider<List<WorkoutPlanSummary>>((ref) async {
  final dioClient = ref.watch(dioClientProvider);
  final response = await dioClient.dio.get('/workouts');
  final List data = response.data as List;
  return data.map((e) => WorkoutPlanSummary.fromJson(e as Map<String, dynamic>)).toList();
});

final planDetailsProvider = FutureProvider.family<PlanDetails, String>((ref, id) async {
  final dioClient = ref.watch(dioClientProvider);
  final response = await dioClient.dio.get('/workouts/$id');
  return PlanDetails.fromJson(response.data as Map<String, dynamic>);
});
