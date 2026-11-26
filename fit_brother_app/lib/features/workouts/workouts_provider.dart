import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/api/dio_client.dart';
import '../auth/auth_provider.dart';
import '../catalog/catalog_provider.dart';

// ---------- Legacy Models ----------

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
  final String? catalogExerciseId;
  final String name;
  final int sets;
  final String reps;
  final int restSeconds;
  final String notes;
  final String videoUrl;
  final bool hasVideo;

  const ExerciseSummary({
    required this.id,
    this.catalogExerciseId,
    required this.name,
    required this.sets,
    required this.reps,
    required this.restSeconds,
    required this.notes,
    required this.videoUrl,
    this.hasVideo = false,
  });

  factory ExerciseSummary.fromJson(Map<String, dynamic> json) => ExerciseSummary(
        id: json['id'] as String,
        catalogExerciseId: json['catalog_exercise_id'] as String?,
        name: json['name'] as String,
        sets: json['sets'] as int,
        reps: json['reps'] as String? ?? '',
        restSeconds: json['rest_seconds'] as int,
        notes: json['notes'] as String? ?? '',
        videoUrl: json['video_url'] as String? ?? '',
        hasVideo: json['has_video'] as bool? ?? false,
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

// ---------- Phase 5 Builder Models ----------

class BuilderExerciseItem {
  final String id;
  final String? catalogExerciseId;
  final String? legacyExerciseId;
  final String? mediaId;
  final String name;
  final String primaryMuscle;
  final String equipment;
  final int sets;
  final String reps;
  final int restSeconds;
  final String notes;
  final int position;
  final bool hasMedia;

  const BuilderExerciseItem({
    this.id = '',
    this.catalogExerciseId,
    this.legacyExerciseId,
    this.mediaId,
    required this.name,
    this.primaryMuscle = '',
    this.equipment = '',
    required this.sets,
    required this.reps,
    required this.restSeconds,
    this.notes = '',
    this.position = 0,
    this.hasMedia = false,
  });

  BuilderExerciseItem copyWith({
    String? id,
    String? catalogExerciseId,
    String? legacyExerciseId,
    String? mediaId,
    String? name,
    String? primaryMuscle,
    String? equipment,
    int? sets,
    String? reps,
    int? restSeconds,
    String? notes,
    int? position,
    bool? hasMedia,
  }) =>
      BuilderExerciseItem(
        id: id ?? this.id,
        catalogExerciseId: catalogExerciseId ?? this.catalogExerciseId,
        legacyExerciseId: legacyExerciseId ?? this.legacyExerciseId,
        mediaId: mediaId ?? this.mediaId,
        name: name ?? this.name,
        primaryMuscle: primaryMuscle ?? this.primaryMuscle,
        equipment: equipment ?? this.equipment,
        sets: sets ?? this.sets,
        reps: reps ?? this.reps,
        restSeconds: restSeconds ?? this.restSeconds,
        notes: notes ?? this.notes,
        position: position ?? this.position,
        hasMedia: hasMedia ?? this.hasMedia,
      );

  factory BuilderExerciseItem.fromJson(Map<String, dynamic> json) => BuilderExerciseItem(
        id: json['id'] as String? ?? '',
        catalogExerciseId: json['catalog_exercise_id'] as String?,
        legacyExerciseId: json['legacy_exercise_id'] as String?,
        mediaId: json['media_id'] as String?,
        name: json['name'] as String? ?? '',
        primaryMuscle: json['primary_muscle'] as String? ?? '',
        equipment: json['equipment'] as String? ?? '',
        sets: (json['sets'] as num?)?.toInt() ?? 0,
        reps: json['reps'] as String? ?? '',
        restSeconds: (json['rest_seconds'] as num?)?.toInt() ?? 0,
        notes: json['notes'] as String? ?? '',
        position: (json['position'] as num?)?.toInt() ?? 0,
        hasMedia: json['has_media'] as bool? ?? false,
      );

  Map<String, dynamic> toInputJson(int idx) => {
        if (catalogExerciseId != null && catalogExerciseId!.isNotEmpty)
          'catalog_exercise_id': catalogExerciseId,
        if (mediaId != null && mediaId!.isNotEmpty) 'media_id': mediaId,
        'name': name,
        'primary_muscle': primaryMuscle,
        'equipment': equipment,
        'sets': sets,
        'reps': reps,
        'rest_seconds': restSeconds,
        'notes': notes,
        'position': idx + 1,
      };
}

class BuilderSessionItem {
  final String id;
  final String title;
  final int dayOfWeek;
  final String notes;
  final int position;
  final List<BuilderExerciseItem> exercises;

  const BuilderSessionItem({
    this.id = '',
    required this.title,
    this.dayOfWeek = 1,
    this.notes = '',
    this.position = 0,
    this.exercises = const [],
  });

  BuilderSessionItem copyWith({
    String? id,
    String? title,
    int? dayOfWeek,
    String? notes,
    int? position,
    List<BuilderExerciseItem>? exercises,
  }) =>
      BuilderSessionItem(
        id: id ?? this.id,
        title: title ?? this.title,
        dayOfWeek: dayOfWeek ?? this.dayOfWeek,
        notes: notes ?? this.notes,
        position: position ?? this.position,
        exercises: exercises ?? this.exercises,
      );

  factory BuilderSessionItem.fromJson(Map<String, dynamic> json) => BuilderSessionItem(
        id: json['id'] as String? ?? '',
        title: json['title'] as String? ?? '',
        dayOfWeek: (json['day_of_week'] as num?)?.toInt() ?? 0,
        notes: json['notes'] as String? ?? '',
        position: (json['position'] as num?)?.toInt() ?? 0,
        exercises: (json['exercises'] as List? ?? const [])
            .map((e) => BuilderExerciseItem.fromJson(Map<String, dynamic>.from(e as Map)))
            .toList(),
      );

  Map<String, dynamic> toInputJson(int idx) => {
        'title': title,
        'day_of_week': dayOfWeek,
        'notes': notes,
        'position': idx + 1,
        'exercises': exercises
            .asMap()
            .entries
            .map((entry) => entry.value.toInputJson(entry.key))
            .toList(),
      };
}

class WorkoutTemplateSummary {
  final String id;
  final String organizationId;
  final String title;
  final String description;
  final String goal;
  final String? clonedFromTemplateId;
  final bool isArchived;
  final int latestVersionNumber;
  final String? latestVersionId;
  final int sessionCount;
  final int exerciseCount;

  const WorkoutTemplateSummary({
    required this.id,
    required this.organizationId,
    required this.title,
    this.description = '',
    this.goal = '',
    this.clonedFromTemplateId,
    this.isArchived = false,
    this.latestVersionNumber = 0,
    this.latestVersionId,
    this.sessionCount = 0,
    this.exerciseCount = 0,
  });

  factory WorkoutTemplateSummary.fromJson(Map<String, dynamic> json) =>
      WorkoutTemplateSummary(
        id: json['id'] as String? ?? '',
        organizationId: json['organization_id'] as String? ?? '',
        title: json['title'] as String? ?? '',
        description: json['description'] as String? ?? '',
        goal: json['goal'] as String? ?? '',
        clonedFromTemplateId: json['cloned_from_template_id'] as String?,
        isArchived: json['is_archived'] as bool? ?? false,
        latestVersionNumber: (json['latest_version_number'] as num?)?.toInt() ?? 0,
        latestVersionId: json['latest_version_id'] as String?,
        sessionCount: (json['session_count'] as num?)?.toInt() ?? 0,
        exerciseCount: (json['exercise_count'] as num?)?.toInt() ?? 0,
      );
}

class WorkoutVersionSummary {
  final String id;
  final String templateId;
  final int versionNumber;
  final String title;
  final String goal;
  final int sessionCount;
  final String publishedAt;

  const WorkoutVersionSummary({
    required this.id,
    required this.templateId,
    required this.versionNumber,
    required this.title,
    this.goal = '',
    this.sessionCount = 0,
    this.publishedAt = '',
  });

  factory WorkoutVersionSummary.fromJson(Map<String, dynamic> json) =>
      WorkoutVersionSummary(
        id: json['id'] as String? ?? '',
        templateId: json['template_id'] as String? ?? '',
        versionNumber: (json['version_number'] as num?)?.toInt() ?? 0,
        title: json['title'] as String? ?? '',
        goal: json['goal'] as String? ?? '',
        sessionCount: (json['session_count'] as num?)?.toInt() ?? 0,
        publishedAt: json['published_at'] as String? ?? '',
      );
}

class WorkoutTemplateDetail {
  final WorkoutTemplateSummary summary;
  final List<BuilderSessionItem> sessions;
  final List<WorkoutVersionSummary> versions;

  const WorkoutTemplateDetail({
    required this.summary,
    required this.sessions,
    required this.versions,
  });

  factory WorkoutTemplateDetail.fromJson(Map<String, dynamic> json) {
    final draft = Map<String, dynamic>.from(json['draft_structure'] as Map? ?? const {});
    final sessionsList = draft['sessions'] as List? ?? const [];
    final versionsList = json['versions'] as List? ?? const [];

    return WorkoutTemplateDetail(
      summary: WorkoutTemplateSummary.fromJson(json),
      sessions: sessionsList
          .map((e) => BuilderSessionItem.fromJson(Map<String, dynamic>.from(e as Map)))
          .toList(),
      versions: versionsList
          .map((e) => WorkoutVersionSummary.fromJson(Map<String, dynamic>.from(e as Map)))
          .toList(),
    );
  }
}

class WorkoutAssignmentSummary {
  final String id;
  final String templateId;
  final String workoutVersionId;
  final String studentId;
  final String studentName;
  final String templateTitle;
  final String goal;
  final int versionNumber;
  final String validFrom;
  final String? validUntil;
  final String status;
  final String notes;
  final int sessionCount;

  const WorkoutAssignmentSummary({
    required this.id,
    required this.templateId,
    required this.workoutVersionId,
    required this.studentId,
    required this.studentName,
    required this.templateTitle,
    this.goal = '',
    required this.versionNumber,
    required this.validFrom,
    this.validUntil,
    required this.status,
    this.notes = '',
    this.sessionCount = 0,
  });

  factory WorkoutAssignmentSummary.fromJson(Map<String, dynamic> json) =>
      WorkoutAssignmentSummary(
        id: json['id'] as String? ?? '',
        templateId: json['template_id'] as String? ?? '',
        workoutVersionId: json['workout_version_id'] as String? ?? '',
        studentId: json['student_id'] as String? ?? '',
        studentName: json['student_name'] as String? ?? '',
        templateTitle: json['template_title'] as String? ?? '',
        goal: json['goal'] as String? ?? '',
        versionNumber: (json['version_number'] as num?)?.toInt() ?? 1,
        validFrom: json['valid_from'] as String? ?? '',
        validUntil: json['valid_until'] as String?,
        status: json['status'] as String? ?? 'ACTIVE',
        notes: json['notes'] as String? ?? '',
        sessionCount: (json['session_count'] as num?)?.toInt() ?? 0,
      );
}

class WorkoutAssignmentDetail {
  final WorkoutAssignmentSummary summary;
  final List<BuilderSessionItem> sessions;

  const WorkoutAssignmentDetail({
    required this.summary,
    required this.sessions,
  });

  factory WorkoutAssignmentDetail.fromJson(Map<String, dynamic> json) {
    final snapshot = Map<String, dynamic>.from(json['snapshot_structure'] as Map? ?? const {});
    final sessionsList = snapshot['sessions'] as List? ?? const [];

    return WorkoutAssignmentDetail(
      summary: WorkoutAssignmentSummary.fromJson(json),
      sessions: sessionsList
          .map((e) => BuilderSessionItem.fromJson(Map<String, dynamic>.from(e as Map)))
          .toList(),
    );
  }
}

class ResolvedExerciseMedia {
  final String signedUrl;
  final String mediaType;
  final int expiresIn;
  final List<String> instructionsPT;

  const ResolvedExerciseMedia({
    required this.signedUrl,
    required this.mediaType,
    this.expiresIn = 300,
    this.instructionsPT = const [],
  });
}

// ---------- Builder Service ----------

abstract class WorkoutBuilderService {
  Future<List<WorkoutTemplateSummary>> listTemplates({bool includeArchived = false});
  Future<WorkoutTemplateDetail> getTemplate(String id);
  Future<WorkoutTemplateDetail> createTemplate({
    required String title,
    required String description,
    required String goal,
    required List<BuilderSessionItem> sessions,
  });
  Future<WorkoutTemplateDetail> updateTemplate(
    String id, {
    required String title,
    required String description,
    required String goal,
    bool? isArchived,
    required List<BuilderSessionItem> sessions,
  });
  Future<WorkoutTemplateDetail> cloneTemplate(String id, {String title = ''});
  Future<WorkoutVersionSummary> publishVersion(String id);
  Future<WorkoutAssignmentDetail> assignTemplate(
    String templateId, {
    required String studentId,
    String? workoutVersionId,
    String notes = '',
  });
  Future<List<WorkoutAssignmentSummary>> listAssignments({String? studentId});
  Future<WorkoutAssignmentDetail> getAssignment(String id);
  Future<void> updateAssignmentStatus(String id, String status);
  Future<ResolvedExerciseMedia> resolveExerciseMedia({
    String? legacyExerciseId,
    String? catalogExerciseId,
    String? mediaId,
    String? fallbackVideoUrl,
  });
}

class DioWorkoutBuilderService implements WorkoutBuilderService {
  final DioClient _client;
  final CatalogService _catalogService;

  const DioWorkoutBuilderService(this._client, this._catalogService);

  @override
  Future<List<WorkoutTemplateSummary>> listTemplates({bool includeArchived = false}) async {
    final res = await _client.dio.get<dynamic>(
      '/workout-templates',
      queryParameters: {'include_archived': includeArchived},
    );
    final list = res.data as List? ?? const [];
    return list
        .map((e) => WorkoutTemplateSummary.fromJson(Map<String, dynamic>.from(e as Map)))
        .toList();
  }

  @override
  Future<WorkoutTemplateDetail> getTemplate(String id) async {
    final res = await _client.dio.get<dynamic>('/workout-templates/$id');
    return WorkoutTemplateDetail.fromJson(Map<String, dynamic>.from(res.data as Map));
  }

  @override
  Future<WorkoutTemplateDetail> createTemplate({
    required String title,
    required String description,
    required String goal,
    required List<BuilderSessionItem> sessions,
  }) async {
    final res = await _client.dio.post<dynamic>(
      '/workout-templates',
      data: {
        'title': title,
        'description': description,
        'goal': goal,
        'sessions': sessions
            .asMap()
            .entries
            .map((entry) => entry.value.toInputJson(entry.key))
            .toList(),
      },
    );
    return WorkoutTemplateDetail.fromJson(Map<String, dynamic>.from(res.data as Map));
  }

  @override
  Future<WorkoutTemplateDetail> updateTemplate(
    String id, {
    required String title,
    required String description,
    required String goal,
    bool? isArchived,
    required List<BuilderSessionItem> sessions,
  }) async {
    final res = await _client.dio.patch<dynamic>(
      '/workout-templates/$id',
      data: {
        'title': title,
        'description': description,
        'goal': goal,
        if (isArchived != null) 'is_archived': isArchived,
        'sessions': sessions
            .asMap()
            .entries
            .map((entry) => entry.value.toInputJson(entry.key))
            .toList(),
      },
    );
    return WorkoutTemplateDetail.fromJson(Map<String, dynamic>.from(res.data as Map));
  }

  @override
  Future<WorkoutTemplateDetail> cloneTemplate(String id, {String title = ''}) async {
    final res = await _client.dio.post<dynamic>(
      '/workout-templates/$id/clone',
      data: {if (title.trim().isNotEmpty) 'title': title.trim()},
    );
    return WorkoutTemplateDetail.fromJson(Map<String, dynamic>.from(res.data as Map));
  }

  @override
  Future<WorkoutVersionSummary> publishVersion(String id) async {
    final res = await _client.dio.post<dynamic>('/workout-templates/$id/publish');
    return WorkoutVersionSummary.fromJson(Map<String, dynamic>.from(res.data as Map));
  }

  @override
  Future<WorkoutAssignmentDetail> assignTemplate(
    String templateId, {
    required String studentId,
    String? workoutVersionId,
    String notes = '',
  }) async {
    final res = await _client.dio.post<dynamic>(
      '/workout-templates/$templateId/assign',
      data: {
        'student_id': studentId,
        if (workoutVersionId != null && workoutVersionId.isNotEmpty)
          'workout_version_id': workoutVersionId,
        if (notes.trim().isNotEmpty) 'notes': notes.trim(),
      },
    );
    return WorkoutAssignmentDetail.fromJson(Map<String, dynamic>.from(res.data as Map));
  }

  @override
  Future<List<WorkoutAssignmentSummary>> listAssignments({String? studentId}) async {
    final res = await _client.dio.get<dynamic>(
      '/workout-assignments',
      queryParameters: {
        if (studentId != null && studentId.isNotEmpty) 'student_id': studentId,
      },
    );
    final list = res.data as List? ?? const [];
    return list
        .map((e) => WorkoutAssignmentSummary.fromJson(Map<String, dynamic>.from(e as Map)))
        .toList();
  }

  @override
  Future<WorkoutAssignmentDetail> getAssignment(String id) async {
    final res = await _client.dio.get<dynamic>('/workout-assignments/$id');
    return WorkoutAssignmentDetail.fromJson(Map<String, dynamic>.from(res.data as Map));
  }

  @override
  Future<void> updateAssignmentStatus(String id, String status) async {
    await _client.dio.patch<dynamic>(
      '/workout-assignments/$id/status',
      data: {'status': status},
    );
  }

  @override
  Future<ResolvedExerciseMedia> resolveExerciseMedia({
    String? legacyExerciseId,
    String? catalogExerciseId,
    String? mediaId,
    String? fallbackVideoUrl,
  }) async {
    if (catalogExerciseId != null && catalogExerciseId.isNotEmpty) {
      final detail = await _catalogService.getExercise(catalogExerciseId);
      String? targetMediaId = mediaId;
      if ((targetMediaId == null || targetMediaId.isEmpty) && detail.media.isNotEmpty) {
        targetMediaId = detail.media.first.id;
      }
      if (targetMediaId != null && targetMediaId.isNotEmpty) {
        final signed = await _catalogService.getSignedMediaUrl(catalogExerciseId, targetMediaId);
        return ResolvedExerciseMedia(
          signedUrl: signed.signedUrl,
          mediaType: signed.mediaType,
          expiresIn: signed.expiresIn,
          instructionsPT: detail.instructionsPT,
        );
      }
      return ResolvedExerciseMedia(
        signedUrl: '',
        mediaType: 'INSTRUCTIONS',
        instructionsPT: detail.instructionsPT,
      );
    }

    if (legacyExerciseId != null && legacyExerciseId.isNotEmpty) {
      final res = await _client.dio.get<dynamic>(
        '/videos/exercises/$legacyExerciseId/signed-url',
      );
      final map = Map<String, dynamic>.from(res.data as Map);
      return ResolvedExerciseMedia(
        signedUrl: map['signed_url'] as String? ?? '',
        mediaType: 'VIDEO',
        expiresIn: (map['expires_in'] as num?)?.toInt() ?? 300,
      );
    }

    if (fallbackVideoUrl != null && fallbackVideoUrl.trim().isNotEmpty) {
      return ResolvedExerciseMedia(
        signedUrl: fallbackVideoUrl.trim(),
        mediaType: 'VIDEO',
      );
    }

    throw StateError('Nenhuma mídia configurada para este exercício.');
  }
}

// ---------- Providers ----------

final workoutBuilderServiceProvider = Provider<WorkoutBuilderService>((ref) {
  return DioWorkoutBuilderService(
    ref.watch(dioClientProvider),
    ref.watch(catalogServiceProvider),
  );
});

final workoutPlansProvider = FutureProvider<List<WorkoutPlanSummary>>((ref) async {
  final isLoggedIn = ref.watch(authProvider.select((s) => s.isLoggedIn));
  if (!isLoggedIn) {
    return const [];
  }
  final dioClient = ref.watch(dioClientProvider);
  final response = await dioClient.dio.get('/workouts');
  final List data = response.data as List;
  return data.map((e) => WorkoutPlanSummary.fromJson(e as Map<String, dynamic>)).toList();
});

final planDetailsProvider = FutureProvider.family<PlanDetails, String>((ref, id) async {
  final isLoggedIn = ref.watch(authProvider.select((s) => s.isLoggedIn));
  if (!isLoggedIn) {
    throw StateError('Usuário não autenticado');
  }
  final dioClient = ref.watch(dioClientProvider);
  final response = await dioClient.dio.get('/workouts/$id');
  return PlanDetails.fromJson(response.data as Map<String, dynamic>);
});

final workoutTemplatesProvider = FutureProvider<List<WorkoutTemplateSummary>>((ref) async {
  final isLoggedIn = ref.watch(authProvider.select((s) => s.isLoggedIn));
  if (!isLoggedIn) {
    return const [];
  }
  return ref.watch(workoutBuilderServiceProvider).listTemplates();
});

final workoutTemplateDetailProvider =
    FutureProvider.family<WorkoutTemplateDetail, String>((ref, id) async {
  final isLoggedIn = ref.watch(authProvider.select((s) => s.isLoggedIn));
  if (!isLoggedIn) {
    throw StateError('Usuário não autenticado');
  }
  return ref.watch(workoutBuilderServiceProvider).getTemplate(id);
});

final workoutAssignmentsProvider = FutureProvider<List<WorkoutAssignmentSummary>>((ref) async {
  final isLoggedIn = ref.watch(authProvider.select((s) => s.isLoggedIn));
  if (!isLoggedIn) {
    return const [];
  }
  return ref.watch(workoutBuilderServiceProvider).listAssignments();
});

final workoutAssignmentDetailProvider =
    FutureProvider.family<WorkoutAssignmentDetail, String>((ref, id) async {
  final isLoggedIn = ref.watch(authProvider.select((s) => s.isLoggedIn));
  if (!isLoggedIn) {
    throw StateError('Usuário não autenticado');
  }
  return ref.watch(workoutBuilderServiceProvider).getAssignment(id);
});
