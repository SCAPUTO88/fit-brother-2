import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:fit_brother_app/core/router/app_router.dart';
import 'package:fit_brother_app/features/auth/auth_provider.dart';
import 'package:fit_brother_app/features/auth/invite_accept_screen.dart';
import 'package:fit_brother_app/features/auth/login_screen.dart';
import 'package:fit_brother_app/features/auth/password_reset_screen.dart';
import 'package:fit_brother_app/features/catalog/catalog_provider.dart';
import 'package:fit_brother_app/features/catalog/exercise_catalog_screen.dart';
import 'package:fit_brother_app/features/students/create_student_screen.dart';
import 'package:fit_brother_app/features/workouts/plan_detail_screen.dart';
import 'package:fit_brother_app/features/workouts/workout_templates_screen.dart';
import 'package:fit_brother_app/features/workouts/workouts_provider.dart';
import 'package:fit_brother_app/main.dart';

class _TestAuthNotifier extends AuthNotifier {
  final AuthState _initial;

  _TestAuthNotifier(this._initial);

  @override
  AuthState build() => _initial;
}

class _FakeWorkoutBuilderService implements WorkoutBuilderService {
  String? lastLegacyExerciseId;
  bool clonedCalled = false;

  final List<WorkoutTemplateSummary> _templates = [
    const WorkoutTemplateSummary(
      id: 'tpl-1',
      organizationId: 'org-1',
      title: 'Hipertrofia Base AB',
      description: 'Treino base para hipertrofia',
      goal: 'Hipertrofia',
      latestVersionNumber: 1,
      sessionCount: 2,
      exerciseCount: 4,
    ),
  ];

  final List<WorkoutAssignmentSummary> _assignments = const [
    WorkoutAssignmentSummary(
      id: 'asgn-1',
      templateId: 'tpl-1',
      workoutVersionId: 'ver-1',
      studentId: 'stu-1',
      studentName: 'João Silva',
      templateTitle: 'Hipertrofia Base AB',
      goal: 'Hipertrofia',
      versionNumber: 1,
      validFrom: '2026-04-01',
      status: 'ACTIVE',
      sessionCount: 2,
    ),
  ];

  @override
  Future<List<WorkoutTemplateSummary>> listTemplates({bool includeArchived = false}) async {
    return List<WorkoutTemplateSummary>.from(_templates);
  }

  @override
  Future<WorkoutTemplateDetail> getTemplate(String id) async {
    return WorkoutTemplateDetail(
      summary: _templates.first,
      sessions: const [],
      versions: const [],
    );
  }

  @override
  Future<WorkoutTemplateDetail> createTemplate({
    required String title,
    required String description,
    required String goal,
    required List<BuilderSessionItem> sessions,
  }) {
    throw UnimplementedError();
  }

  @override
  Future<WorkoutTemplateDetail> updateTemplate(
    String id, {
    required String title,
    required String description,
    required String goal,
    bool? isArchived,
    required List<BuilderSessionItem> sessions,
  }) {
    throw UnimplementedError();
  }

  @override
  Future<WorkoutTemplateDetail> cloneTemplate(String id, {String title = ''}) async {
    clonedCalled = true;
    final copy = WorkoutTemplateSummary(
      id: 'tpl-2',
      organizationId: 'org-1',
      title: title.isEmpty ? 'Cópia de Hipertrofia Base AB' : title,
      goal: 'Hipertrofia',
      latestVersionNumber: 0,
      sessionCount: 2,
      exerciseCount: 4,
    );
    _templates.add(copy);
    return WorkoutTemplateDetail(summary: copy, sessions: const [], versions: const []);
  }

  @override
  Future<WorkoutVersionSummary> publishVersion(String id) async {
    return const WorkoutVersionSummary(
      id: 'ver-2',
      templateId: 'tpl-1',
      versionNumber: 2,
      title: 'Hipertrofia Base AB',
    );
  }

  @override
  Future<WorkoutAssignmentDetail> assignTemplate(
    String templateId, {
    required String studentId,
    String? workoutVersionId,
    String notes = '',
  }) {
    throw UnimplementedError();
  }

  @override
  Future<List<WorkoutAssignmentSummary>> listAssignments({String? studentId}) async {
    return _assignments;
  }

  @override
  Future<WorkoutAssignmentDetail> getAssignment(String id) {
    throw UnimplementedError();
  }

  @override
  Future<void> updateAssignmentStatus(String id, String status) async {}

  @override
  Future<ResolvedExerciseMedia> resolveExerciseMedia({
    String? legacyExerciseId,
    String? catalogExerciseId,
    String? mediaId,
    String? fallbackVideoUrl,
  }) async {
    lastLegacyExerciseId = legacyExerciseId;
    return const ResolvedExerciseMedia(
      signedUrl: 'https://storage.fitbrother.test/videos/signed-demo.mp4',
      mediaType: 'VIDEO',
      expiresIn: 300,
    );
  }
}

class _FakeCatalogService implements CatalogService {
  final List<CatalogExerciseSummary> _items = const [
    CatalogExerciseSummary(
      id: 'ex-1',
      externalSource: 'free-exercise-db',
      namePT: 'Supino Reto com Barra',
      nameEN: 'Barbell Bench Press - Medium Grip',
      primaryMuscle: 'chest',
      primaryMusclePT: 'Peitoral',
      equipment: 'barbell',
      equipmentPT: 'Barra',
      category: 'strength',
      categoryPT: 'Força / Hipertrofia',
      level: 'beginner',
      licenseInfo: 'The Unlicense (Public Domain)',
      isCustom: false,
    ),
    CatalogExerciseSummary(
      id: 'ex-2',
      organizationId: 'org-1',
      externalSource: 'custom',
      namePT: 'Crucifixo Customizado do Studio',
      primaryMuscle: 'chest',
      primaryMusclePT: 'Peitoral',
      equipment: 'cable',
      equipmentPT: 'Polia / Cabo',
      category: 'strength',
      categoryPT: 'Força / Hipertrofia',
      level: 'intermediate',
      licenseInfo: 'Organização / Conteúdo Próprio',
      isCustom: true,
    ),
  ];

  @override
  Future<CatalogFiltersData> fetchFilters() async {
    return const CatalogFiltersData(
      muscleGroups: [CatalogOption(code: 'chest', namePT: 'Peitoral')],
      equipments: [CatalogOption(code: 'barbell', namePT: 'Barra')],
      categories: [CatalogOption(code: 'strength', namePT: 'Força / Hipertrofia')],
    );
  }

  @override
  Future<CatalogSearchResult> searchExercises(CatalogSearchQuery q) async {
    final filtered = _items.where((item) {
      if (q.scope == 'custom' && !item.isCustom) return false;
      if (q.scope == 'global' && item.isCustom) return false;
      if (q.query.isNotEmpty &&
          !item.namePT.toLowerCase().contains(q.query.toLowerCase())) {
        return false;
      }
      return true;
    }).toList();
    return CatalogSearchResult(items: filtered, total: filtered.length);
  }

  @override
  Future<CatalogExerciseDetail> getExercise(String id) async {
    final found = _items.firstWhere((e) => e.id == id);
    return CatalogExerciseDetail(
      summary: found,
      instructionsPT: const ['Mantenha as escápulas estabilizadas.'],
      media: const [],
    );
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
  }) {
    throw UnimplementedError();
  }

  @override
  Future<void> deleteCustomExercise(String id) async {}

  @override
  Future<CatalogMediaItem> uploadExerciseMedia({
    required String exerciseId,
    required String filename,
    required String mimeType,
    required String mediaType,
    required Uint8List bytes,
  }) {
    throw UnimplementedError();
  }

  @override
  Future<SignedMediaResponse> getSignedMediaUrl(String exerciseId, String mediaId) {
    throw UnimplementedError();
  }
}

void main() {
  testWidgets('App smoke test', (WidgetTester tester) async {
    await tester.pumpWidget(const ProviderScope(child: FitBrotherApp()));
    expect(find.byType(MaterialApp), findsOneWidget);
    expect(find.byType(LoginScreen), findsOneWidget);
    expect(find.text('Esqueci minha senha'), findsOneWidget);
    expect(find.text('Ativar convite'), findsOneWidget);
  });

  testWidgets('Redirects unauthenticated access to /login', (WidgetTester tester) async {
    final container = ProviderContainer();
    addTearDown(container.dispose);

    await tester.pumpWidget(
      UncontrolledProviderScope(
        container: container,
        child: const FitBrotherApp(),
      ),
    );

    container.read(appRouterProvider).go('/dashboard');
    await tester.pumpAndSettle();

    expect(find.byType(LoginScreen), findsOneWidget);
  });

  testWidgets('Allows unauthenticated access to /forgot-password and /activate', (WidgetTester tester) async {
    final container = ProviderContainer();
    addTearDown(container.dispose);

    await tester.pumpWidget(
      UncontrolledProviderScope(
        container: container,
        child: const FitBrotherApp(),
      ),
    );

    container.read(appRouterProvider).go('/forgot-password');
    await tester.pumpAndSettle();
    expect(find.byType(PasswordResetScreen), findsOneWidget);

    container.read(appRouterProvider).go('/activate?token=abc-123');
    await tester.pumpAndSettle();
    expect(find.byType(InviteAcceptScreen), findsOneWidget);
    expect(find.text('abc-123'), findsOneWidget);
  });

  testWidgets('CreateStudentScreen does not require CPF (ADR-006)', (WidgetTester tester) async {
    await tester.pumpWidget(
      const ProviderScope(
        child: MaterialApp(home: CreateStudentScreen()),
      ),
    );

    expect(find.text('Enviar Convite'), findsOneWidget);
    expect(find.text('Cadastro Direto'), findsOneWidget);
    expect(find.textContaining('CPF (somente números)'), findsNothing);
  });

  testWidgets('ExerciseCatalogScreen displays PT-BR catalog, filters, and Trainer FAB', (WidgetTester tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authProvider.overrideWith(
            () => _TestAuthNotifier(const AuthState(isLoggedIn: true, role: 'TRAINER')),
          ),
          catalogServiceProvider.overrideWithValue(_FakeCatalogService()),
        ],
        child: const MaterialApp(home: ExerciseCatalogScreen()),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Biblioteca de Exercícios'), findsOneWidget);
    expect(find.text('Supino Reto com Barra'), findsOneWidget);
    expect(find.text('Crucifixo Customizado do Studio'), findsOneWidget);
    expect(find.text('Novo Exercício'), findsOneWidget);

    await tester.tap(find.text('Meus Customizados'));
    await tester.pumpAndSettle();

    expect(find.text('Supino Reto com Barra'), findsNothing);
    expect(find.text('Crucifixo Customizado do Studio'), findsOneWidget);
  });

  testWidgets('ExerciseCatalogScreen hides creation FAB for Student role', (WidgetTester tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authProvider.overrideWith(
            () => _TestAuthNotifier(const AuthState(isLoggedIn: true, role: 'STUDENT')),
          ),
          catalogServiceProvider.overrideWithValue(_FakeCatalogService()),
        ],
        child: const MaterialApp(home: ExerciseCatalogScreen()),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Supino Reto com Barra'), findsOneWidget);
    expect(find.text('Novo Exercício'), findsNothing);
  });

  testWidgets('WorkoutTemplatesScreen displays templates, clone action, and assignments tab', (WidgetTester tester) async {
    final fakeBuilder = _FakeWorkoutBuilderService();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authProvider.overrideWith(
            () => _TestAuthNotifier(const AuthState(isLoggedIn: true, role: 'TRAINER')),
          ),
          workoutBuilderServiceProvider.overrideWithValue(fakeBuilder),
        ],
        child: const MaterialApp(home: WorkoutTemplatesScreen()),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Workout Builder'), findsOneWidget);
    expect(find.text('Hipertrofia Base AB'), findsOneWidget);
    expect(find.text('Versão v1'), findsOneWidget);

    await tester.tap(find.text('Clonar'));
    await tester.pumpAndSettle();
    expect(fakeBuilder.clonedCalled, isTrue);

    await tester.tap(find.text('Atribuições'));
    await tester.pumpAndSettle();
    expect(find.textContaining('João Silva'), findsOneWidget);
  });

  testWidgets('PlanDetailScreen triggers signed media dialog when hasVideo is true even with empty videoUrl', (WidgetTester tester) async {
    final fakeBuilder = _FakeWorkoutBuilderService();
    const fakeDetails = PlanDetails(
      plan: WorkoutPlanSummary(
        id: 'plan-1',
        title: 'Plano Hipertrofia',
        studentId: 'stu-1',
        description: 'Treino completo',
        isActive: true,
      ),
      sessions: [
        SessionDetails(
          id: 'sess-1',
          title: 'Treino A',
          dayOfWeek: 1,
          notes: '',
          exercises: [
            ExerciseSummary(
              id: 'ex-legacy-1',
              name: 'Supino Reto',
              sets: 4,
              reps: '8-10',
              restSeconds: 90,
              notes: '',
              videoUrl: '',
              hasVideo: true,
            ),
          ],
        ),
      ],
    );

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authProvider.overrideWith(
            () => _TestAuthNotifier(const AuthState(isLoggedIn: true, role: 'STUDENT')),
          ),
          workoutBuilderServiceProvider.overrideWithValue(fakeBuilder),
          planDetailsProvider('plan-1').overrideWith((ref) async => fakeDetails),
        ],
        child: const MaterialApp(home: PlanDetailScreen(planId: 'plan-1')),
      ),
    );
    await tester.pumpAndSettle();

    final playButton = find.byIcon(Icons.play_circle_fill);
    expect(playButton, findsOneWidget);

    await tester.tap(playButton);
    await tester.pumpAndSettle();

    expect(fakeBuilder.lastLegacyExerciseId, equals('ex-legacy-1'));
    expect(find.text('URL assinada de vídeo gerada com segurança'), findsOneWidget);
    expect(find.text('https://storage.fitbrother.test/videos/signed-demo.mp4'), findsOneWidget);
  });
}
