import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../auth/auth_provider.dart';
import 'workouts_provider.dart';

class WorkoutPlansScreen extends ConsumerWidget {
  const WorkoutPlansScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);
    final isStudent = authState.role == 'STUDENT';
    final plansAsync = ref.watch(workoutPlansProvider);
    final assignmentsAsync = ref.watch(workoutAssignmentsProvider);

    return Scaffold(
      backgroundColor: const Color(0xFFF5F7FA),
      appBar: AppBar(
        title: const Text('Planos de Treino', style: TextStyle(fontWeight: FontWeight.bold)),
        backgroundColor: Colors.white,
        foregroundColor: Colors.black87,
        elevation: 1,
        actions: [
          if (!isStudent)
            TextButton.icon(
              onPressed: () => context.push('/workout-builder'),
              icon: const Icon(Icons.layers_outlined),
              label: const Text('Workout Builder'),
            ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(workoutAssignmentsProvider);
          ref.invalidate(workoutPlansProvider);
        },
        child: ListView(
          padding: const EdgeInsets.all(24),
          children: [
            const Text(
              'Treinos Publicados (Versões)',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.black87),
            ),
            const SizedBox(height: 12),
            assignmentsAsync.when(
              loading: () => const Padding(
                padding: EdgeInsets.symmetric(vertical: 24),
                child: Center(child: CircularProgressIndicator()),
              ),
              error: (err, _) => Card(
                elevation: 0,
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    children: [
                      Text('Erro ao carregar atribuições: $err', style: const TextStyle(color: Colors.red)),
                      const SizedBox(height: 8),
                      TextButton(
                        onPressed: () => ref.invalidate(workoutAssignmentsProvider),
                        child: const Text('Tentar novamente'),
                      ),
                    ],
                  ),
                ),
              ),
              data: (assignments) {
                if (assignments.isEmpty) {
                  return const Card(
                    elevation: 0,
                    child: Padding(
                      padding: EdgeInsets.all(16),
                      child: Text(
                        'Nenhum treino versionado atribuído ainda.',
                        style: TextStyle(color: Colors.grey),
                      ),
                    ),
                  );
                }
                return Column(
                  children: assignments.map((a) {
                    return Card(
                      elevation: 0,
                      margin: const EdgeInsets.only(bottom: 8),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                        side: const BorderSide(color: Color(0xFFE8ECF0)),
                      ),
                      child: ListTile(
                        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                        leading: CircleAvatar(
                          backgroundColor: a.status == 'ACTIVE'
                              ? Colors.blue.shade100
                              : Colors.grey.shade200,
                          child: Text(
                            'v${a.versionNumber}',
                            style: TextStyle(
                              fontWeight: FontWeight.bold,
                              color: a.status == 'ACTIVE' ? Colors.blue.shade800 : Colors.grey.shade700,
                            ),
                          ),
                        ),
                        title: Text(a.templateTitle, style: const TextStyle(fontWeight: FontWeight.w600)),
                        subtitle: Text(
                          [
                            if (a.goal.isNotEmpty) a.goal,
                            '${a.sessionCount} sessões',
                            a.status,
                          ].join(' • '),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        trailing: const Icon(Icons.chevron_right),
                        onTap: () => context.push('/workout-assignments/${a.id}'),
                      ),
                    );
                  }).toList(),
                );
              },
            ),
            const SizedBox(height: 24),
            const Text(
              'Planos Legados',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.black87),
            ),
            const SizedBox(height: 12),
            plansAsync.when(
              loading: () => const Padding(
                padding: EdgeInsets.symmetric(vertical: 24),
                child: Center(child: CircularProgressIndicator()),
              ),
              error: (err, _) => Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(Icons.error_outline, color: Colors.red, size: 48),
                    const SizedBox(height: 12),
                    Text(
                      'Erro ao carregar planos:\n$err',
                      textAlign: TextAlign.center,
                      style: const TextStyle(color: Colors.red),
                    ),
                    const SizedBox(height: 16),
                    ElevatedButton.icon(
                      icon: const Icon(Icons.refresh),
                      label: const Text('Tentar novamente'),
                      onPressed: () => ref.invalidate(workoutPlansProvider),
                    ),
                  ],
                ),
              ),
              data: (plans) => plans.isEmpty
                  ? const Center(
                      child: Padding(
                        padding: EdgeInsets.symmetric(vertical: 24),
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.assignment_outlined, size: 56, color: Colors.grey),
                            SizedBox(height: 12),
                            Text('Nenhum plano legado cadastrado.', style: TextStyle(color: Colors.grey)),
                          ],
                        ),
                      ),
                    )
                  : Column(
                      children: plans.map((p) {
                        return Card(
                          elevation: 0,
                          margin: const EdgeInsets.only(bottom: 8),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12),
                            side: const BorderSide(color: Color(0xFFE8ECF0)),
                          ),
                          child: ListTile(
                            contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                            leading: CircleAvatar(
                              backgroundColor: p.isActive ? Colors.green.shade100 : Colors.grey.shade200,
                              child: Icon(Icons.assignment, color: p.isActive ? Colors.green : Colors.grey),
                            ),
                            title: Text(p.title, style: const TextStyle(fontWeight: FontWeight.w600)),
                            subtitle: p.description.isNotEmpty
                                ? Text(p.description, maxLines: 1, overflow: TextOverflow.ellipsis)
                                : null,
                            trailing: const Icon(Icons.chevron_right),
                            onTap: () {
                              context.push('/workouts/${p.id}');
                            },
                          ),
                        );
                      }).toList(),
                    ),
            ),
          ],
        ),
      ),
      floatingActionButton: isStudent
          ? null
          : FloatingActionButton.extended(
              onPressed: () => context.push('/workouts/create'),
              icon: const Icon(Icons.add),
              label: const Text('Novo Plano'),
            ),
    );
  }
}
