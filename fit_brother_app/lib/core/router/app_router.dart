import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../features/auth/auth_provider.dart';
import '../../features/auth/invite_accept_screen.dart';
import '../../features/auth/login_screen.dart';
import '../../features/auth/mfa_settings_screen.dart';
import '../../features/auth/password_reset_screen.dart';
import '../../features/dashboard/dashboard_screen.dart';
import '../../features/privacy/privacy_center_screen.dart';
import '../../features/students/create_student_screen.dart';
import '../../features/workouts/create_exercise_screen.dart';
import '../../features/workouts/create_plan_screen.dart';
import '../../features/workouts/create_session_screen.dart';
import '../../features/workouts/plan_detail_screen.dart';
import '../../features/workouts/workout_plans_screen.dart';

class _RouterRefreshNotifier extends ChangeNotifier {
  void refresh() => notifyListeners();
}

const _publicRoutes = {'/login', '/forgot-password', '/activate'};

final appRouterProvider = Provider<GoRouter>((ref) {
  final refreshListenable = _RouterRefreshNotifier();
  ref.listen<AuthState>(authProvider, (_, _) => refreshListenable.refresh());
  ref.onDispose(refreshListenable.dispose);

  return GoRouter(
    initialLocation: '/login',
    refreshListenable: refreshListenable,
    redirect: (context, state) {
      final isLoggedIn = ref.read(authProvider).isLoggedIn;
      final isPublicRoute = _publicRoutes.contains(state.matchedLocation);

      if (!isLoggedIn && !isPublicRoute) {
        return '/login';
      }
      if (isLoggedIn && isPublicRoute) {
        return '/dashboard';
      }
      return null;
    },
    routes: [
      GoRoute(
        path: '/login',
        builder: (context, state) => const LoginScreen(),
      ),
      GoRoute(
        path: '/forgot-password',
        builder: (context, state) => PasswordResetScreen(
          initialToken: state.uri.queryParameters['token'],
        ),
      ),
      GoRoute(
        path: '/activate',
        builder: (context, state) => InviteAcceptScreen(
          initialToken: state.uri.queryParameters['token'],
        ),
      ),
      GoRoute(
        path: '/dashboard',
        builder: (context, state) => const DashboardScreen(),
      ),
      GoRoute(
        path: '/privacy',
        builder: (context, state) => const PrivacyCenterScreen(),
      ),
      GoRoute(
        path: '/settings/security',
        builder: (context, state) => const MFASettingsScreen(),
      ),
      GoRoute(
        path: '/students/create',
        builder: (context, state) => const CreateStudentScreen(),
      ),
      GoRoute(
        path: '/workouts',
        builder: (context, state) => const WorkoutPlansScreen(),
      ),
      GoRoute(
        path: '/workouts/create',
        builder: (context, state) => const CreatePlanScreen(),
      ),
      GoRoute(
        path: '/workouts/:id',
        builder: (context, state) {
          final id = state.pathParameters['id']!;
          return PlanDetailScreen(planId: id);
        },
      ),
      GoRoute(
        path: '/workouts/:id/sessions/create',
        builder: (context, state) {
          final planId = state.pathParameters['id']!;
          return CreateSessionScreen(planId: planId);
        },
      ),
      GoRoute(
        path: '/workouts/:plan_id/sessions/:session_id/exercises/create',
        builder: (context, state) {
          final planId = state.pathParameters['plan_id']!;
          final sessionId = state.pathParameters['session_id']!;
          return CreateExerciseScreen(planId: planId, sessionId: sessionId);
        },
      ),
    ],
  );
});
