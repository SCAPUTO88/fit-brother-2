import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'workouts_provider.dart';

class PlanDetailScreen extends ConsumerWidget {
  final String planId;

  const PlanDetailScreen({super.key, required this.planId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final detailAsync = ref.watch(planDetailsProvider(planId));

    return Scaffold(
      appBar: AppBar(title: const Text('Detalhes do Plano')),
      body: detailAsync.when(
        loading: () =>
            const Center(child: CircularProgressIndicator(color: Colors.blue)),
        error: (err, stack) => Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(
                Icons.error_outline,
                size: 48,
                color: Colors.redAccent,
              ),
              const SizedBox(height: 16),
              Text(
                'Erro ao carregar detalhes: $err',
                style: const TextStyle(color: Colors.redAccent),
              ),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () => ref.invalidate(planDetailsProvider(planId)),
                child: const Text('Tentar Novamente'),
              ),
            ],
          ),
        ),
        data: (details) {
          final plan = details.plan;
          final sessions = details.sessions;

          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(planDetailsProvider(planId)),
            child: ListView(
              padding: const EdgeInsets.all(16.0),
              children: [
                // Plan Info Card
                Card(
                  elevation: 2,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Padding(
                    padding: const EdgeInsets.all(16.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          plan.title,
                          style: const TextStyle(
                            fontSize: 24,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          plan.description,
                          style: const TextStyle(
                            fontSize: 16,
                            color: Colors.grey,
                          ),
                        ),
                        const SizedBox(height: 16),
                        Row(
                          children: [
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 12,
                                vertical: 6,
                              ),
                              decoration: BoxDecoration(
                                color: plan.isActive
                                    ? Colors.green.withValues(alpha: 0.1)
                                    : Colors.red.withValues(alpha: 0.1),
                                borderRadius: BorderRadius.circular(16),
                              ),
                              child: Text(
                                plan.isActive ? 'Ativo' : 'Inativo',
                                style: TextStyle(
                                  color: plan.isActive
                                      ? Colors.green
                                      : Colors.red,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 24),

                // Sessions Header
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text(
                      'Sessões de Treino',
                      style: TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    TextButton.icon(
                      onPressed: () {
                        context.push('/workouts/$planId/sessions/create');
                      },
                      icon: const Icon(Icons.add),
                      label: const Text('Nova Sessão'),
                    ),
                  ],
                ),
                const SizedBox(height: 16),

                // Sessions List
                if (sessions.isEmpty)
                  const Center(
                    child: Padding(
                      padding: EdgeInsets.all(32.0),
                      child: Text(
                        'Nenhuma sessão cadastrada neste plano.',
                        style: TextStyle(color: Colors.grey),
                      ),
                    ),
                  )
                else
                  ...sessions.map((s) => _buildSessionCard(context, s)),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildSessionCard(BuildContext context, SessionDetails session) {
    const days = [
      'Domingo',
      'Segunda-feira',
      'Terça-feira',
      'Quarta-feira',
      'Quinta-feira',
      'Sexta-feira',
      'Sábado',
    ];
    final dayStr = (session.dayOfWeek >= 0 && session.dayOfWeek < 7)
        ? days[session.dayOfWeek]
        : 'Dia Livre';

    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Theme(
        data: Theme.of(context).copyWith(dividerColor: Colors.transparent),
        child: ExpansionTile(
          initiallyExpanded: true,
          title: Text(
            '${session.title} - $dayStr',
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          subtitle: session.notes.isNotEmpty
              ? Text(
                  session.notes,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                )
              : null,
          children: [
            const Divider(),
            if (session.exercises.isEmpty)
              const Padding(
                padding: EdgeInsets.all(16.0),
                child: Text(
                  'Nenhum exercício cadastrado.',
                  style: TextStyle(color: Colors.grey),
                ),
              )
            else
              ListView.separated(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                itemCount: session.exercises.length,
                separatorBuilder: (_, index) => const Divider(height: 1),
                itemBuilder: (context, index) {
                  final ex = session.exercises[index];
                  return ListTile(
                    leading: CircleAvatar(
                      backgroundColor: Colors.blue.withValues(alpha: 0.1),
                      child: Text(
                        '${index + 1}',
                        style: const TextStyle(
                          color: Colors.blue,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                    title: Text(
                      ex.name,
                      style: const TextStyle(fontWeight: FontWeight.bold),
                    ),
                    subtitle: Text(
                      '${ex.sets} séries x ${ex.reps} rep. • ${ex.restSeconds}s recup.',
                    ),
                    trailing: ex.videoUrl.isNotEmpty
                        ? const Icon(Icons.play_circle_fill, color: Colors.blue)
                        : null,
                  );
                },
              ),
            Padding(
              padding: const EdgeInsets.all(8.0),
              child: TextButton.icon(
                onPressed: () {
                  context.push(
                    '/workouts/$planId/sessions/${session.id}/exercises/create',
                  );
                },
                icon: const Icon(Icons.add_circle_outline),
                label: const Text('Adicionar Exercício'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
