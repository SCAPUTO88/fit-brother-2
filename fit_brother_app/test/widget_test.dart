import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:fit_brother_app/core/router/app_router.dart';
import 'package:fit_brother_app/features/auth/invite_accept_screen.dart';
import 'package:fit_brother_app/features/auth/login_screen.dart';
import 'package:fit_brother_app/features/auth/password_reset_screen.dart';
import 'package:fit_brother_app/features/students/create_student_screen.dart';
import 'package:fit_brother_app/main.dart';

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
}
