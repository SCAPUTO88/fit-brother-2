import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/api/dio_client.dart';
import '../auth/auth_provider.dart';

class CatalogOption {
  final String code;
  final String namePT;
  final String nameEN;

  const CatalogOption({
    required this.code,
    required this.namePT,
    this.nameEN = '',
  });

  factory CatalogOption.fromJson(Map<String, dynamic> json) => CatalogOption(
        code: json['code'] as String? ?? '',
        namePT: json['name_pt'] as String? ?? '',
        nameEN: json['name_en'] as String? ?? '',
      );
}

class CatalogFiltersData {
  final List<CatalogOption> muscleGroups;
  final List<CatalogOption> equipments;
  final List<CatalogOption> categories;

  const CatalogFiltersData({
    required this.muscleGroups,
    required this.equipments,
    required this.categories,
  });

  factory CatalogFiltersData.fromJson(Map<String, dynamic> json) {
    final muscles = (json['muscle_groups'] as List? ?? const [])
        .map((e) => CatalogOption.fromJson(Map<String, dynamic>.from(e as Map)))
        .toList();
    final equips = (json['equipments'] as List? ?? const [])
        .map((e) => CatalogOption.fromJson(Map<String, dynamic>.from(e as Map)))
        .toList();
    final cats = (json['categories'] as List? ?? const [])
        .map((e) => CatalogOption.fromJson(Map<String, dynamic>.from(e as Map)))
        .toList();
    return CatalogFiltersData(
      muscleGroups: muscles,
      equipments: equips,
      categories: cats,
    );
  }
}

class CatalogMediaItem {
  final String id;
  final String mediaType;
  final String mimeType;
  final String filename;
  final int sizeBytes;
  final String status;

  const CatalogMediaItem({
    required this.id,
    required this.mediaType,
    required this.mimeType,
    required this.filename,
    required this.sizeBytes,
    required this.status,
  });

  factory CatalogMediaItem.fromJson(Map<String, dynamic> json) => CatalogMediaItem(
        id: json['id'] as String? ?? '',
        mediaType: json['media_type'] as String? ?? '',
        mimeType: json['mime_type'] as String? ?? '',
        filename: json['filename'] as String? ?? '',
        sizeBytes: (json['size_bytes'] as num?)?.toInt() ?? 0,
        status: json['status'] as String? ?? '',
      );
}

class CatalogExerciseSummary {
  final String id;
  final String? organizationId;
  final String externalSource;
  final String namePT;
  final String nameEN;
  final String primaryMuscle;
  final String primaryMusclePT;
  final List<String> secondaryMuscles;
  final String equipment;
  final String equipmentPT;
  final String category;
  final String categoryPT;
  final String level;
  final String licenseInfo;
  final bool isCustom;

  const CatalogExerciseSummary({
    required this.id,
    this.organizationId,
    required this.externalSource,
    required this.namePT,
    this.nameEN = '',
    required this.primaryMuscle,
    required this.primaryMusclePT,
    this.secondaryMuscles = const [],
    required this.equipment,
    required this.equipmentPT,
    required this.category,
    required this.categoryPT,
    required this.level,
    required this.licenseInfo,
    required this.isCustom,
  });

  factory CatalogExerciseSummary.fromJson(Map<String, dynamic> json) =>
      CatalogExerciseSummary(
        id: json['id'] as String? ?? '',
        organizationId: json['organization_id'] as String?,
        externalSource: json['external_source'] as String? ?? '',
        namePT: json['name_pt'] as String? ?? '',
        nameEN: json['name_en'] as String? ?? '',
        primaryMuscle: json['primary_muscle'] as String? ?? '',
        primaryMusclePT: json['primary_muscle_pt'] as String? ?? '',
        secondaryMuscles: (json['secondary_muscles'] as List? ?? const [])
            .map((e) => e.toString())
            .toList(),
        equipment: json['equipment'] as String? ?? '',
        equipmentPT: json['equipment_pt'] as String? ?? '',
        category: json['category'] as String? ?? '',
        categoryPT: json['category_pt'] as String? ?? '',
        level: json['level'] as String? ?? '',
        licenseInfo: json['license_info'] as String? ?? '',
        isCustom: json['is_custom'] as bool? ?? false,
      );
}

class CatalogExerciseDetail {
  final CatalogExerciseSummary summary;
  final List<String> instructionsPT;
  final List<String> instructionsEN;
  final List<CatalogMediaItem> media;

  const CatalogExerciseDetail({
    required this.summary,
    required this.instructionsPT,
    this.instructionsEN = const [],
    required this.media,
  });

  factory CatalogExerciseDetail.fromJson(Map<String, dynamic> json) =>
      CatalogExerciseDetail(
        summary: CatalogExerciseSummary.fromJson(json),
        instructionsPT: (json['instructions_pt'] as List? ?? const [])
            .map((e) => e.toString())
            .toList(),
        instructionsEN: (json['instructions_en'] as List? ?? const [])
            .map((e) => e.toString())
            .toList(),
        media: (json['media'] as List? ?? const [])
            .map((e) => CatalogMediaItem.fromJson(Map<String, dynamic>.from(e as Map)))
            .toList(),
      );
}

class CatalogSearchQuery {
  final String query;
  final String muscle;
  final String equipment;
  final String category;
  final String scope;
  final int limit;
  final int offset;

  const CatalogSearchQuery({
    this.query = '',
    this.muscle = '',
    this.equipment = '',
    this.category = '',
    this.scope = 'all',
    this.limit = 50,
    this.offset = 0,
  });

  CatalogSearchQuery copyWith({
    String? query,
    String? muscle,
    String? equipment,
    String? category,
    String? scope,
    int? limit,
    int? offset,
  }) =>
      CatalogSearchQuery(
        query: query ?? this.query,
        muscle: muscle ?? this.muscle,
        equipment: equipment ?? this.equipment,
        category: category ?? this.category,
        scope: scope ?? this.scope,
        limit: limit ?? this.limit,
        offset: offset ?? this.offset,
      );

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is CatalogSearchQuery &&
          runtimeType == other.runtimeType &&
          query == other.query &&
          muscle == other.muscle &&
          equipment == other.equipment &&
          category == other.category &&
          scope == other.scope &&
          limit == other.limit &&
          offset == other.offset;

  @override
  int get hashCode =>
      Object.hash(query, muscle, equipment, category, scope, limit, offset);
}

class CatalogSearchResult {
  final List<CatalogExerciseSummary> items;
  final int total;

  const CatalogSearchResult({required this.items, required this.total});

  factory CatalogSearchResult.fromJson(Map<String, dynamic> json) =>
      CatalogSearchResult(
        items: (json['items'] as List? ?? const [])
            .map((e) => CatalogExerciseSummary.fromJson(Map<String, dynamic>.from(e as Map)))
            .toList(),
        total: (json['total'] as num?)?.toInt() ?? 0,
      );
}

class SignedMediaResponse {
  final String signedUrl;
  final String mediaType;
  final int expiresIn;

  const SignedMediaResponse({
    required this.signedUrl,
    required this.mediaType,
    required this.expiresIn,
  });

  factory SignedMediaResponse.fromJson(Map<String, dynamic> json) =>
      SignedMediaResponse(
        signedUrl: json['signed_url'] as String? ?? '',
        mediaType: json['media_type'] as String? ?? '',
        expiresIn: (json['expires_in'] as num?)?.toInt() ?? 300,
      );
}

abstract class CatalogService {
  Future<CatalogFiltersData> fetchFilters();
  Future<CatalogSearchResult> searchExercises(CatalogSearchQuery q);
  Future<CatalogExerciseDetail> getExercise(String id);
  Future<CatalogExerciseDetail> createCustomExercise({
    required String namePT,
    required String primaryMuscle,
    required List<String> secondaryMuscles,
    required String equipment,
    required String category,
    required String level,
    required List<String> instructionsPT,
  });
  Future<void> deleteCustomExercise(String id);
  Future<CatalogMediaItem> uploadExerciseMedia({
    required String exerciseId,
    required String filename,
    required String mimeType,
    required String mediaType,
    required Uint8List bytes,
  });
  Future<SignedMediaResponse> getSignedMediaUrl(String exerciseId, String mediaId);
}

class DioCatalogService implements CatalogService {
  final DioClient _client;

  const DioCatalogService(this._client);

  @override
  Future<CatalogFiltersData> fetchFilters() async {
    final res = await _client.dio.get<dynamic>('/catalog/filters');
    return CatalogFiltersData.fromJson(Map<String, dynamic>.from(res.data as Map));
  }

  @override
  Future<CatalogSearchResult> searchExercises(CatalogSearchQuery q) async {
    final params = <String, dynamic>{
      'limit': q.limit,
      'offset': q.offset,
      'scope': q.scope,
    };
    if (q.query.trim().isNotEmpty) params['q'] = q.query.trim();
    if (q.muscle.trim().isNotEmpty) params['muscle'] = q.muscle.trim();
    if (q.equipment.trim().isNotEmpty) params['equipment'] = q.equipment.trim();
    if (q.category.trim().isNotEmpty) params['category'] = q.category.trim();

    final res = await _client.dio.get<dynamic>(
      '/catalog/exercises',
      queryParameters: params,
    );
    return CatalogSearchResult.fromJson(Map<String, dynamic>.from(res.data as Map));
  }

  @override
  Future<CatalogExerciseDetail> getExercise(String id) async {
    final res = await _client.dio.get<dynamic>('/catalog/exercises/$id');
    return CatalogExerciseDetail.fromJson(Map<String, dynamic>.from(res.data as Map));
  }

  @override
  Future<CatalogExerciseDetail> createCustomExercise({
    required String namePT,
    required String primaryMuscle,
    required List<String> secondaryMuscles,
    required String equipment,
    required String category,
    required String level,
    required List<String> instructionsPT,
  }) async {
    final res = await _client.dio.post<dynamic>(
      '/catalog/exercises',
      data: {
        'name_pt': namePT,
        'primary_muscle': primaryMuscle,
        'secondary_muscles': secondaryMuscles,
        'equipment': equipment,
        'category': category,
        'level': level,
        'instructions_pt': instructionsPT,
      },
    );
    return CatalogExerciseDetail.fromJson(Map<String, dynamic>.from(res.data as Map));
  }

  @override
  Future<void> deleteCustomExercise(String id) async {
    await _client.dio.delete<dynamic>('/catalog/exercises/$id');
  }

  @override
  Future<CatalogMediaItem> uploadExerciseMedia({
    required String exerciseId,
    required String filename,
    required String mimeType,
    required String mediaType,
    required Uint8List bytes,
  }) async {
    final reqRes = await _client.dio.post<dynamic>(
      '/catalog/exercises/$exerciseId/media/upload-url',
      data: {
        'filename': filename,
        'mime_type': mimeType,
        'media_type': mediaType,
      },
    );
    final reqData = Map<String, dynamic>.from(reqRes.data as Map);
    final presignedUrl = reqData['presigned_url'] as String;
    final storageKey = reqData['storage_key'] as String;

    final uploadDio = Dio();
    await uploadDio.put<dynamic>(
      presignedUrl,
      data: bytes,
      options: Options(
        headers: {
          'Content-Type': mimeType,
          'Content-Length': bytes.length,
        },
      ),
    );

    final confirmRes = await _client.dio.post<dynamic>(
      '/catalog/exercises/$exerciseId/media/confirm',
      data: {
        'storage_key': storageKey,
        'filename': filename,
        'mime_type': mimeType,
        'size_bytes': bytes.length,
      },
    );
    return CatalogMediaItem.fromJson(Map<String, dynamic>.from(confirmRes.data as Map));
  }

  @override
  Future<SignedMediaResponse> getSignedMediaUrl(String exerciseId, String mediaId) async {
    final res = await _client.dio.get<dynamic>(
      '/catalog/exercises/$exerciseId/media/$mediaId/signed-url',
    );
    return SignedMediaResponse.fromJson(Map<String, dynamic>.from(res.data as Map));
  }
}

final catalogServiceProvider = Provider<CatalogService>((ref) {
  return DioCatalogService(ref.watch(dioClientProvider));
});

final catalogFiltersProvider = FutureProvider<CatalogFiltersData>((ref) async {
  return ref.watch(catalogServiceProvider).fetchFilters();
});

final catalogSearchProvider =
    FutureProvider.family<CatalogSearchResult, CatalogSearchQuery>((ref, query) async {
  return ref.watch(catalogServiceProvider).searchExercises(query);
});

final catalogDetailProvider =
    FutureProvider.family<CatalogExerciseDetail, String>((ref, id) async {
  return ref.watch(catalogServiceProvider).getExercise(id);
});
