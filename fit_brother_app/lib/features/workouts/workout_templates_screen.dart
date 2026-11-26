import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../students/students_provider.dart';
import 'workouts_provider.dart';

class WorkoutTemplatesScreen extends ConsumerStatefulWidget {
  const WorkoutTemplatesScreen({super.key});

  @override
  ConsumerState<WorkoutTemplatesScreen> createState() =>
      _WorkoutTemplatesScreenState();
}

class _WorkoutTemplatesScreenState
    extends ConsumerState<WorkoutTemplatesScreen> {
  int _selectedTab = 0; // 0 = Templates, 1 = Atribuições

  Future<void> _cloneTemplate(WorkoutTemplateSummary tpl) async {
    try {
      final cloned = await ref
          .read(workoutBuilderServiceProvider)
          .cloneTemplate(tpl.id, title: 'Cópia de ${tpl.title}');
      ref.invalidate(workoutTemplatesProvider);
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Template "${cloned.summary.title}" clonado!'),
          backgroundColor: Colors.green,
        ),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Erro ao clonar template: $e'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  Future<void> _publishTemplate(WorkoutTemplateSummary tpl) async {
    try {
      final ver = await ref
          .read(workoutBuilderServiceProvider)
          .publishVersion(tpl.id);
      ref.invalidate(workoutTemplatesProvider);
      ref.invalidate(workoutTemplateDetailProvider(tpl.id));
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Versão v${ver.versionNumber} publicada (snapshot imutável).'),
          backgroundColor: Colors.green,
        ),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Erro ao publicar versão: $e'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  Future<void> _showAssignDialog(WorkoutTemplateSummary tpl) async {
    String? selectedStudentId;
    final notesController = TextEditingController();

    final students = await ref.read(studentsProvider.future).catchError((_) => <StudentSummary>[]);
    if (!mounted) return;

    if (students.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Cadastre ou ative ao menos um aluno antes de atribuir.'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setLocalState) => AlertDialog(
          title: Text('Atribuir "${tpl.title}"'),
          content: SizedBox(
            width: 420,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                DropdownButtonFormField<String>(
                  decoration: const InputDecoration(
                    labelText: 'Aluno',
                    border: OutlineInputBorder(),
                  ),
                  initialValue: selectedStudentId,
                  items: students
                      .map((s) => DropdownMenuItem(value: s.id, child: Text(s.fullName)))
                      .toList(),
                  onChanged: (val) => setLocalState(() => selectedStudentId = val),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: notesController,
                  decoration: const InputDecoration(
                    labelText: 'Observações para o aluno (opcional)',
                    border: OutlineInputBorder(),
                  ),
                  maxLines: 2,
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: const Text('Cancelar'),
            ),
            ElevatedButton(
              onPressed: selectedStudentId == null
                  ? null
                  : () => Navigator.pop(ctx, true),
              child: const Text('Publicar & Atribuir'),
            ),
          ],
        ),
      ),
    );

    if (ok != true || selectedStudentId == null) {
      notesController.dispose();
      return;
    }

    try {
      final assigned = await ref.read(workoutBuilderServiceProvider).assignTemplate(
            tpl.id,
            studentId: selectedStudentId!,
            notes: notesController.text.trim(),
          );
      ref.invalidate(workoutTemplatesProvider);
      ref.invalidate(workoutAssignmentsProvider);
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Treino v${assigned.summary.versionNumber} atribuído para ${assigned.summary.studentName}!',
          ),
          backgroundColor: Colors.green,
        ),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Erro ao atribuir treino: $e'),
          backgroundColor: Colors.red,
        ),
      );
    } finally {
      notesController.dispose();
    }
  }

  @override
  Widget build(BuildContext context) {
    final templatesAsync = ref.watch(workoutTemplatesProvider);
    final assignmentsAsync = ref.watch(workoutAssignmentsProvider);

    return Scaffold(
      backgroundColor: const Color(0xFFF5F7FA),
      appBar: AppBar(
        title: const Text(
          'Workout Builder',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
        backgroundColor: Colors.white,
        foregroundColor: Colors.black87,
        elevation: 1,
      ),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SegmentedButton<int>(
              segments: const [
                ButtonSegment(
                  value: 0,
                  icon: Icon(Icons.layers_outlined),
                  label: Text('Templates'),
                ),
                ButtonSegment(
                  value: 1,
                  icon: Icon(Icons.assignment_turned_in_outlined),
                  label: Text('Atribuições'),
                ),
              ],
              selected: {_selectedTab},
              onSelectionChanged: (sel) => setState(() => _selectedTab = sel.first),
            ),
            const SizedBox(height: 16),
            Expanded(
              child: _selectedTab == 0
                  ? _buildTemplatesList(templatesAsync)
                  : _buildAssignmentsList(assignmentsAsync),
            ),
          ],
        ),
      ),
      floatingActionButton: _selectedTab == 0
          ? FloatingActionButton.extended(
              onPressed: () => context.push('/workout-builder/new'),
              icon: const Icon(Icons.add),
              label: const Text('Novo Template'),
            )
          : null,
    );
  }

  Widget _buildTemplatesList(AsyncValue<List<WorkoutTemplateSummary>> templatesAsync) {
    return templatesAsync.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (err, _) => Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text('Erro ao carregar templates: $err', style: const TextStyle(color: Colors.red)),
            const SizedBox(height: 12),
            ElevatedButton(
              onPressed: () => ref.invalidate(workoutTemplatesProvider),
              child: const Text('Tentar novamente'),
            ),
          ],
        ),
      ),
      data: (items) {
        if (items.isEmpty) {
          return const Center(
            child: Text(
              'Nenhum template criado ainda. Toque em "Novo Template" para começar.',
              style: TextStyle(color: Colors.grey),
            ),
          );
        }
        return ListView.separated(
          itemCount: items.length,
          separatorBuilder: (_, _) => const SizedBox(height: 12),
          itemBuilder: (context, index) {
            final tpl = items[index];
            return Card(
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
                      alignment: WrapAlignment.spaceBetween,
                      crossAxisAlignment: WrapCrossAlignment.center,
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        Text(
                          tpl.title,
                          style: const TextStyle(fontSize: 17, fontWeight: FontWeight.bold),
                        ),
                        Chip(
                          label: Text(
                            tpl.latestVersionNumber > 0
                                ? 'Versão v${tpl.latestVersionNumber}'
                                : 'Rascunho',
                          ),
                          backgroundColor: tpl.latestVersionNumber > 0
                              ? Colors.green.withValues(alpha: 0.12)
                              : Colors.orange.withValues(alpha: 0.12),
                        ),
                      ],
                    ),
                    if (tpl.goal.isNotEmpty || tpl.description.isNotEmpty) ...[
                      const SizedBox(height: 6),
                      Text(
                        [tpl.goal, tpl.description].where((s) => s.isNotEmpty).join(' • '),
                        style: const TextStyle(color: Colors.black54),
                      ),
                    ],
                    const SizedBox(height: 8),
                    Text(
                      '${tpl.sessionCount} sessão(ões) • ${tpl.exerciseCount} exercício(s)',
                      style: const TextStyle(fontSize: 13, color: Colors.black54),
                    ),
                    const SizedBox(height: 12),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        OutlinedButton.icon(
                          onPressed: () => context.push('/workout-builder/${tpl.id}'),
                          icon: const Icon(Icons.edit_outlined, size: 18),
                          label: const Text('Editar'),
                        ),
                        OutlinedButton.icon(
                          onPressed: () => _cloneTemplate(tpl),
                          icon: const Icon(Icons.copy_all_outlined, size: 18),
                          label: const Text('Clonar'),
                        ),
                        OutlinedButton.icon(
                          onPressed: () => _publishTemplate(tpl),
                          icon: const Icon(Icons.publish_outlined, size: 18),
                          label: const Text('Publicar Versão'),
                        ),
                        ElevatedButton.icon(
                          onPressed: () => _showAssignDialog(tpl),
                          icon: const Icon(Icons.person_add_alt_1, size: 18),
                          label: const Text('Atribuir'),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            );
          },
        );
      },
    );
  }

  Widget _buildAssignmentsList(AsyncValue<List<WorkoutAssignmentSummary>> assignmentsAsync) {
    return assignmentsAsync.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (err, _) => Center(
        child: Text('Erro ao carregar atribuições: $err', style: const TextStyle(color: Colors.red)),
      ),
      data: (items) {
        if (items.isEmpty) {
          return const Center(
            child: Text(
              'Nenhuma atribuição publicada ainda.',
              style: TextStyle(color: Colors.grey),
            ),
          );
        }
        return ListView.separated(
          itemCount: items.length,
          separatorBuilder: (_, _) => const SizedBox(height: 10),
          itemBuilder: (context, index) {
            final a = items[index];
            return Card(
              elevation: 0,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
                side: const BorderSide(color: Color(0xFFE4E9EF)),
              ),
              child: ListTile(
                leading: CircleAvatar(
                  backgroundColor: a.status == 'ACTIVE'
                      ? Colors.green.withValues(alpha: 0.14)
                      : Colors.grey.withValues(alpha: 0.2),
                  child: Text(
                    'v${a.versionNumber}',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      color: a.status == 'ACTIVE' ? Colors.green.shade800 : Colors.grey.shade700,
                    ),
                  ),
                ),
                title: Text(a.templateTitle, style: const TextStyle(fontWeight: FontWeight.w600)),
                subtitle: Text(
                  'Aluno: ${a.studentName.isNotEmpty ? a.studentName : a.studentId} • ${a.sessionCount} sessões • ${a.status}',
                ),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => context.push('/workout-assignments/${a.id}'),
              ),
            );
          },
        );
      },
    );
  }
}
