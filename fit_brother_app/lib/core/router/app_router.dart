import 'package:go_router/go_router.dart';
import '../../features/auth/login_screen.dart';
import '../../features/dashboard/dashboard_screen.dart';
import '../../features/students/create_student_screen.dart';
import '../../features/workouts/workout_plans_screen.dart';
import '../../features/workouts/create_plan_screen.dart';

final appRouter = GoRouter(
  initialLocation: '/login',
  routes: [
    GoRoute(
      path: '/login',
      builder: (context, state) => const LoginScreen(),
    ),
    GoRoute(
      path: '/dashboard',
      builder: (context, state) => const DashboardScreen(),
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
  ],
);
