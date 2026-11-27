import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'workout_media_dialog.dart';
import 'workouts_provider.dart';

class AssignmentDetailScreen extends ConsumerWidget {
  final String assignmentId;

  const AssignmentDetailScreen({super.key, required this.assignmentId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final detailAsync = ref.watch(workoutAssignmentDetailProvider(assignmentId));

    return Scaffold(
      backgroundColor: const Color(0xFFF5F7FA),
      appBar: AppBar(
        title: const Text('Treino Publicado'),
        backgroundColor: Colors.white,
        foregroundColor: Colors.black87,
        elevation: 1,
      ),
      body: detailAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.error_outline, color: Colors.red, size: 48),
              const SizedBox(height: 12),
              Text('Erro ao carregar treino: $err', style: const TextStyle(color: Colors.red)),
              const SizedBox(height: 12),
              ElevatedButton(
                onPressed: () => ref.invalidate(workoutAssignmentDetailProvider(assignmentId)),
                child: const Text('Tentar novamente'),
              ),
            ],
          ),
        ),
        data: (detail) {
          final a = detail.summary;
          const days = [
            'Domingo',
            'Segunda-feira',
            'Terça-feira',
            'Quarta-feira',
            'Quinta-feira',
            'Sexta-feira',
            'Sábado',
          ];

          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Card(
                elevation: 0,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                  side: const BorderSide(color: Color(0xFFE4E9EF)),
                ),
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        crossAxisAlignment: WrapCrossAlignment.center,
                        children: [
                          Text(
                            a.templateTitle,
                            style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
                          ),
                          Chip(
                            label: Text('Versão v${a.versionNumber} (Imutável)'),
                            backgroundColor: Colors.blue.withValues(alpha: 0.1),
                          ),
                        ],
                      ),
                      const SizedBox(height: 8),
                      if (a.goal.isNotEmpty) Text('Objetivo: ${a.goal}'),
                      if (a.studentName.isNotEmpty) Text('Aluno: ${a.studentName}'),
                      if (a.notes.isNotEmpty) ...[
                        const SizedBox(height: 6),
                        Text(a.notes, style: const TextStyle(color: Colors.black54)),
                      ],
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),
              ...detail.sessions.map((sess) {
                final dayLabel = (sess.dayOfWeek >= 0 && sess.dayOfWeek < 7)
                    ? days[sess.dayOfWeek]
                    : 'Dia Livre';
                return Card(
                  elevation: 0,
                  margin: const EdgeInsets.only(bottom: 14),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                    side: const BorderSide(color: Color(0xFFE4E9EF)),
                  ),
                  child: ExpansionTile(
                    initiallyExpanded: true,
                    title: Text(
                      '${sess.title} — $dayLabel',
                      style: const TextStyle(fontWeight: FontWeight.bold),
                    ),
                    children: [
                      if (sess.exercises.isEmpty)
                        const Padding(
                          padding: EdgeInsets.all(16),
                          child: Text('Nenhum exercício nesta sessão.'),
                        )
                      else
                        ...sess.exercises.asMap().entries.map((entry) {
                          final idx = entry.key;
                          final ex = entry.value;
                          final canShowDemo = ex.hasMedia ||
                              (ex.catalogExerciseId != null && ex.catalogExerciseId!.isNotEmpty) ||
                              (ex.legacyExerciseId != null && ex.legacyExerciseId!.isNotEmpty);
                          return ListTile(
                            leading: CircleAvatar(
                              backgroundColor: Colors.blue.withValues(alpha: 0.1),
                              child: Text('${idx + 1}', style: const TextStyle(color: Colors.blue)),
                            ),
                            title: Text(ex.name, style: const TextStyle(fontWeight: FontWeight.w600)),
                            subtitle: Text(
                              '${ex.sets} séries x ${ex.reps} rep. • ${ex.restSeconds}s recup.${ex.notes.isNotEmpty ? '\n${ex.notes}' : ''}',
                            ),
                            trailing: canShowDemo
                                ? IconButton(
                                    icon: const Icon(Icons.play_circle_fill, color: Colors.blue),
                                    tooltip: 'Ver demonstração',
                                    onPressed: () => showExerciseMediaDialog(
                                      context: context,
                                      ref: ref,
                                      exerciseName: ex.name,
                                      legacyExerciseId: ex.legacyExerciseId,
                                      catalogExerciseId: ex.catalogExerciseId,
                                      mediaId: ex.mediaId,
                                    ),
                                  )
                                : null,
                          );
                        }),
                    ],
                  ),
                );
              }),
            ],
          );
        },
      ),
    );
  }
}
