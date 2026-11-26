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
import 'package:fit_brother_app/main.dart';

class _TestAuthNotifier extends AuthNotifier {
  final AuthState _initial;

  _TestAuthNotifier(this._initial);

  @override
  AuthState build() => _initial;
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
}
