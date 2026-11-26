import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../catalog/catalog_provider.dart';
import '../students/students_provider.dart';
import 'workouts_provider.dart';

class TemplateEditorScreen extends ConsumerStatefulWidget {
  final String? templateId;

  const TemplateEditorScreen({super.key, this.templateId});

  @override
  ConsumerState<TemplateEditorScreen> createState() =>
      _TemplateEditorScreenState();
}

class _TemplateEditorScreenState extends ConsumerState<TemplateEditorScreen> {
  final _formKey = GlobalKey<FormState>();
  final _titleController = TextEditingController();
  final _goalController = TextEditingController();
  final _descController = TextEditingController();

  List<BuilderSessionItem> _sessions = [];
  List<WorkoutVersionSummary> _versions = [];
  bool _loaded = false;
  bool _saving = false;
  String? _savedTemplateId;

  @override
  void initState() {
    super.initState();
    _savedTemplateId = widget.templateId;
    if (widget.templateId == null) {
      _sessions = [
        const BuilderSessionItem(
          title: 'Treino A',
          dayOfWeek: 1,
          exercises: [],
        ),
      ];
      _loaded = true;
    }
  }

  @override
  void dispose() {
    _titleController.dispose();
    _goalController.dispose();
    _descController.dispose();
    super.dispose();
  }

  void _populateFromDetail(WorkoutTemplateDetail detail) {
    if (_loaded) return;
    _titleController.text = detail.summary.title;
    _goalController.text = detail.summary.goal;
    _descController.text = detail.summary.description;
    _sessions = List<BuilderSessionItem>.from(detail.sessions);
    _versions = List<WorkoutVersionSummary>.from(detail.versions);
    _loaded = true;
  }

  Future<String?> _saveDraft() async {
    if (!_formKey.currentState!.validate()) return null;
    setState(() => _saving = true);
    try {
      final svc = ref.read(workoutBuilderServiceProvider);
      final WorkoutTemplateDetail res;
      if (_savedTemplateId == null) {
        res = await svc.createTemplate(
          title: _titleController.text.trim(),
          description: _descController.text.trim(),
          goal: _goalController.text.trim(),
          sessions: _sessions,
        );
        _savedTemplateId = res.summary.id;
      } else {
        res = await svc.updateTemplate(
          _savedTemplateId!,
          title: _titleController.text.trim(),
          description: _descController.text.trim(),
          goal: _goalController.text.trim(),
          sessions: _sessions,
        );
      }
      _versions = res.versions;
      ref.invalidate(workoutTemplatesProvider);
      if (_savedTemplateId != null) {
        ref.invalidate(workoutTemplateDetailProvider(_savedTemplateId!));
      }
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Rascunho salvo! Versões publicadas anteriores permanecem intactas.'),
            backgroundColor: Colors.green,
          ),
        );
      }
      return _savedTemplateId;
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Erro ao salvar template: $e'), backgroundColor: Colors.red),
        );
      }
      return null;
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  Future<void> _publishNewVersion() async {
    final id = await _saveDraft();
    if (id == null) return;
    setState(() => _saving = true);
    try {
      final ver = await ref.read(workoutBuilderServiceProvider).publishVersion(id);
      final refreshed = await ref.read(workoutBuilderServiceProvider).getTemplate(id);
      if (mounted) {
        setState(() {
          _versions = refreshed.versions;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Snapshot v${ver.versionNumber} publicado com sucesso!'),
            backgroundColor: Colors.green,
          ),
        );
      }
      ref.invalidate(workoutTemplatesProvider);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Erro ao publicar versão: $e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  Future<void> _assignToStudent() async {
    final id = await _saveDraft();
    if (id == null || !mounted) return;

    final students = await ref.read(studentsProvider.future).catchError((_) => <StudentSummary>[]);
    if (!mounted) return;
    if (students.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Nenhum aluno ativo disponível para atribuição.')),
      );
      return;
    }

    String? selectedStudentId;
    final notesCtrl = TextEditingController();
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setLocal) => AlertDialog(
          title: const Text('Publicar & Atribuir ao Aluno'),
          content: SizedBox(
            width: 400,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                DropdownButtonFormField<String>(
                  decoration: const InputDecoration(
                    labelText: 'Selecionar Aluno',
                    border: OutlineInputBorder(),
                  ),
                  initialValue: selectedStudentId,
                  items: students
                      .map((s) => DropdownMenuItem(value: s.id, child: Text(s.fullName)))
                      .toList(),
                  onChanged: (v) => setLocal(() => selectedStudentId = v),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: notesCtrl,
                  decoration: const InputDecoration(
                    labelText: 'Observações (opcional)',
                    border: OutlineInputBorder(),
                  ),
                ),
              ],
            ),
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Cancelar')),
            ElevatedButton(
              onPressed: selectedStudentId == null ? null : () => Navigator.pop(ctx, true),
              child: const Text('Confirmar Atribuição'),
            ),
          ],
        ),
      ),
    );

    if (ok != true || selectedStudentId == null) {
      notesCtrl.dispose();
      return;
    }

    try {
      final out = await ref.read(workoutBuilderServiceProvider).assignTemplate(
            id,
            studentId: selectedStudentId!,
            notes: notesCtrl.text.trim(),
          );
      ref.invalidate(workoutAssignmentsProvider);
      ref.invalidate(workoutTemplatesProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Atribuído v${out.summary.versionNumber} para ${out.summary.studentName}!'),
            backgroundColor: Colors.green,
          ),
        );
        context.pop();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Erro ao atribuir: $e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      notesCtrl.dispose();
    }
  }

  Future<void> _addExerciseFromCatalog(int sessionIndex) async {
    String query = '';
    CatalogExerciseSummary? picked;

    picked = await showDialog<CatalogExerciseSummary>(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setLocal) => AlertDialog(
          title: const Text('Escolher Exercício da Biblioteca'),
          content: SizedBox(
            width: 480,
            height: 400,
            child: Column(
              children: [
                TextField(
                  decoration: const InputDecoration(
                    hintText: 'Buscar por nome em PT-BR...',
                    prefixIcon: Icon(Icons.search),
                    border: OutlineInputBorder(),
                  ),
                  onChanged: (v) => setLocal(() => query = v),
                ),
                const SizedBox(height: 12),
                Expanded(
                  child: FutureBuilder<CatalogSearchResult>(
                    future: ref
                        .read(catalogServiceProvider)
                        .searchExercises(CatalogSearchQuery(query: query, limit: 25)),
                    builder: (context, snap) {
                      if (snap.connectionState == ConnectionState.waiting) {
                        return const Center(child: CircularProgressIndicator());
                      }
                      if (snap.hasError) {
                        return Center(child: Text('Erro ao buscar: ${snap.error}'));
                      }
                      final items = snap.data?.items ?? const [];
                      if (items.isEmpty) {
                        return const Center(child: Text('Nenhum exercício encontrado.'));
                      }
                      return ListView.builder(
                        itemCount: items.length,
                        itemBuilder: (context, idx) {
                          final ex = items[idx];
                          return ListTile(
                            title: Text(ex.namePT),
                            subtitle: Text('${ex.primaryMusclePT} • ${ex.equipmentPT}'),
                            trailing: ex.isCustom
                                ? const Chip(label: Text('Custom'))
                                : null,
                            onTap: () => Navigator.pop(ctx, ex),
                          );
                        },
                      );
                    },
                  ),
                ),
              ],
            ),
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Fechar')),
          ],
        ),
      ),
    );

    if (picked == null || !mounted) return;
    final edited = await _showExerciseParamsDialog(
      BuilderExerciseItem(
        catalogExerciseId: picked.id,
        name: picked.namePT,
        primaryMuscle: picked.primaryMuscle,
        equipment: picked.equipment,
        sets: 4,
        reps: '8-12',
        restSeconds: 90,
      ),
    );
    if (edited == null || !mounted) return;

    setState(() {
      final sess = _sessions[sessionIndex];
      _sessions[sessionIndex] = sess.copyWith(
        exercises: [...sess.exercises, edited],
      );
    });
  }

  Future<void> _addFreeExercise(int sessionIndex) async {
    final created = await _showExerciseParamsDialog(
      const BuilderExerciseItem(
        name: '',
        sets: 3,
        reps: '10-12',
        restSeconds: 60,
      ),
    );
    if (created == null || !mounted) return;
    setState(() {
      final sess = _sessions[sessionIndex];
      _sessions[sessionIndex] = sess.copyWith(
        exercises: [...sess.exercises, created],
      );
    });
  }

  Future<BuilderExerciseItem?> _showExerciseParamsDialog(BuilderExerciseItem initial) {
    final nameCtrl = TextEditingController(text: initial.name);
    final setsCtrl = TextEditingController(text: '${initial.sets}');
    final repsCtrl = TextEditingController(text: initial.reps);
    final restCtrl = TextEditingController(text: '${initial.restSeconds}');
    final notesCtrl = TextEditingController(text: initial.notes);

    return showDialog<BuilderExerciseItem>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(initial.name.isEmpty ? 'Novo Exercício' : initial.name),
        content: SizedBox(
          width: 400,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextField(
                  controller: nameCtrl,
                  decoration: const InputDecoration(
                    labelText: 'Nome do exercício',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: setsCtrl,
                        keyboardType: TextInputType.number,
                        decoration: const InputDecoration(
                          labelText: 'Séries',
                          border: OutlineInputBorder(),
                        ),
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: TextField(
                        controller: repsCtrl,
                        decoration: const InputDecoration(
                          labelText: 'Repetições',
                          border: OutlineInputBorder(),
                        ),
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: TextField(
                        controller: restCtrl,
                        keyboardType: TextInputType.number,
                        decoration: const InputDecoration(
                          labelText: 'Descanso (s)',
                          border: OutlineInputBorder(),
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: notesCtrl,
                  decoration: const InputDecoration(
                    labelText: 'Observações técnicas (opcional)',
                    border: OutlineInputBorder(),
                  ),
                  maxLines: 2,
                ),
              ],
            ),
          ),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancelar')),
          ElevatedButton(
            onPressed: () {
              final name = nameCtrl.text.trim();
              if (name.length < 2) return;
              Navigator.pop(
                ctx,
                initial.copyWith(
                  name: name,
                  sets: int.tryParse(setsCtrl.text.trim()) ?? 3,
                  reps: repsCtrl.text.trim(),
                  restSeconds: int.tryParse(restCtrl.text.trim()) ?? 60,
                  notes: notesCtrl.text.trim(),
                ),
              );
            },
            child: const Text('Salvar Item'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    if (widget.templateId != null && !_loaded) {
      final detailAsync = ref.watch(workoutTemplateDetailProvider(widget.templateId!));
      return Scaffold(
        appBar: AppBar(title: const Text('Editor de Template')),
        body: detailAsync.when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (err, _) => Center(child: Text('Erro ao carregar template: $err')),
          data: (detail) {
            _populateFromDetail(detail);
            return _buildBody();
          },
        ),
      );
    }

    return Scaffold(
      backgroundColor: const Color(0xFFF5F7FA),
      appBar: AppBar(
        title: Text(widget.templateId == null ? 'Novo Template' : 'Editar Template'),
        backgroundColor: Colors.white,
        foregroundColor: Colors.black87,
        elevation: 1,
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    const dayNames = [
      'Domingo',
      'Segunda-feira',
      'Terça-feira',
      'Quarta-feira',
      'Quinta-feira',
      'Sexta-feira',
      'Sábado',
    ];

    return SingleChildScrollView(
      padding: const EdgeInsets.all(20),
      child: Form(
        key: _formKey,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
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
                  children: [
                    TextFormField(
                      controller: _titleController,
                      decoration: const InputDecoration(
                        labelText: 'Título do Template',
                        border: OutlineInputBorder(),
                      ),
                      validator: (v) =>
                          v == null || v.trim().length < 3 ? 'Mínimo 3 caracteres' : null,
                    ),
                    const SizedBox(height: 12),
                    TextFormField(
                      controller: _goalController,
                      decoration: const InputDecoration(
                        labelText: 'Objetivo (ex: Hipertrofia, Força)',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextFormField(
                      controller: _descController,
                      decoration: const InputDecoration(
                        labelText: 'Descrição geral',
                        border: OutlineInputBorder(),
                      ),
                      maxLines: 2,
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),
            Wrap(
              alignment: WrapAlignment.spaceBetween,
              crossAxisAlignment: WrapCrossAlignment.center,
              spacing: 8,
              runSpacing: 8,
              children: [
                const Text(
                  'Sessões do Treino',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                ),
                OutlinedButton.icon(
                  onPressed: () {
                    setState(() {
                      final letter = String.fromCharCode(65 + (_sessions.length % 26));
                      _sessions.add(
                        BuilderSessionItem(
                          title: 'Treino $letter',
                          dayOfWeek: (_sessions.length + 1) % 7,
                        ),
                      );
                    });
                  },
                  icon: const Icon(Icons.add),
                  label: const Text('Adicionar Sessão'),
                ),
              ],
            ),
            const SizedBox(height: 12),
            ..._sessions.asMap().entries.map((entry) {
              final sIdx = entry.key;
              final sess = entry.value;
              return Card(
                elevation: 0,
                margin: const EdgeInsets.only(bottom: 14),
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
                        spacing: 12,
                        runSpacing: 10,
                        crossAxisAlignment: WrapCrossAlignment.center,
                        children: [
                          SizedBox(
                            width: 240,
                            child: TextFormField(
                              initialValue: sess.title,
                              decoration: const InputDecoration(
                                labelText: 'Nome da Sessão',
                                border: OutlineInputBorder(),
                              ),
                              onChanged: (v) {
                                _sessions[sIdx] = _sessions[sIdx].copyWith(title: v);
                              },
                            ),
                          ),
                          SizedBox(
                            width: 190,
                            child: DropdownButtonFormField<int>(
                              decoration: const InputDecoration(
                                labelText: 'Dia da semana',
                                border: OutlineInputBorder(),
                              ),
                              initialValue: sess.dayOfWeek,
                              items: List.generate(
                                7,
                                (i) => DropdownMenuItem(value: i, child: Text(dayNames[i])),
                              ),
                              onChanged: (val) {
                                if (val != null) {
                                  setState(() {
                                    _sessions[sIdx] = _sessions[sIdx].copyWith(dayOfWeek: val);
                                  });
                                }
                              },
                            ),
                          ),
                          if (_sessions.length > 1)
                            IconButton(
                              tooltip: 'Remover sessão',
                              icon: const Icon(Icons.delete_outline, color: Colors.red),
                              onPressed: () {
                                setState(() {
                                  _sessions.removeAt(sIdx);
                                });
                              },
                            ),
                        ],
                      ),
                      const SizedBox(height: 12),
                      if (sess.exercises.isEmpty)
                        const Padding(
                          padding: EdgeInsets.symmetric(vertical: 8),
                          child: Text(
                            'Nenhum exercício nesta sessão ainda.',
                            style: TextStyle(color: Colors.black54),
                          ),
                        )
                      else
                        ...sess.exercises.asMap().entries.map((exEntry) {
                          final eIdx = exEntry.key;
                          final ex = exEntry.value;
                          return ListTile(
                            dense: true,
                            contentPadding: EdgeInsets.zero,
                            leading: CircleAvatar(
                              radius: 14,
                              backgroundColor: Colors.blue.shade50,
                              child: Text('${eIdx + 1}', style: const TextStyle(color: Colors.blue)),
                            ),
                            title: Text(ex.name, style: const TextStyle(fontWeight: FontWeight.w600)),
                            subtitle: Text('${ex.sets}x ${ex.reps} • ${ex.restSeconds}s descanso'),
                            trailing: Wrap(
                              spacing: 4,
                              children: [
                                if (eIdx > 0)
                                  IconButton(
                                    icon: const Icon(Icons.arrow_upward, size: 18),
                                    tooltip: 'Mover acima',
                                    onPressed: () {
                                      setState(() {
                                        final list = List<BuilderExerciseItem>.from(sess.exercises);
                                        final item = list.removeAt(eIdx);
                                        list.insert(eIdx - 1, item);
                                        _sessions[sIdx] = sess.copyWith(exercises: list);
                                      });
                                    },
                                  ),
                                IconButton(
                                  icon: const Icon(Icons.delete_outline, size: 18, color: Colors.red),
                                  tooltip: 'Remover exercício',
                                  onPressed: () {
                                    setState(() {
                                      final list = List<BuilderExerciseItem>.from(sess.exercises);
                                      list.removeAt(eIdx);
                                      _sessions[sIdx] = sess.copyWith(exercises: list);
                                    });
                                  },
                                ),
                              ],
                            ),
                          );
                        }),
                      const SizedBox(height: 8),
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: [
                          OutlinedButton.icon(
                            onPressed: () => _addExerciseFromCatalog(sIdx),
                            icon: const Icon(Icons.menu_book, size: 18),
                            label: const Text('Adicionar do Catálogo'),
                          ),
                          TextButton.icon(
                            onPressed: () => _addFreeExercise(sIdx),
                            icon: const Icon(Icons.add_circle_outline, size: 18),
                            label: const Text('Exercício Livre'),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              );
            }),
            const SizedBox(height: 12),
            Wrap(
              spacing: 12,
              runSpacing: 12,
              children: [
                ElevatedButton.icon(
                  onPressed: _saving ? null : _saveDraft,
                  icon: const Icon(Icons.save_outlined),
                  label: Text(_saving ? 'Salvando...' : 'Salvar Rascunho'),
                ),
                OutlinedButton.icon(
                  onPressed: _saving ? null : _publishNewVersion,
                  icon: const Icon(Icons.publish),
                  label: const Text('Publicar Nova Versão'),
                ),
                ElevatedButton.icon(
                  onPressed: _saving ? null : _assignToStudent,
                  icon: const Icon(Icons.person_add),
                  label: const Text('Atribuir a Aluno'),
                ),
              ],
            ),
            if (_versions.isNotEmpty) ...[
              const SizedBox(height: 24),
              const Text(
                'Histórico de Versões Publicadas (Imutáveis)',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              ..._versions.map(
                (v) => Card(
                  elevation: 0,
                  child: ListTile(
                    leading: CircleAvatar(
                      backgroundColor: Colors.green.shade50,
                      child: Text('v${v.versionNumber}', style: const TextStyle(color: Colors.green)),
                    ),
                    title: Text(v.title),
                    subtitle: Text('${v.sessionCount} sessões • Publicada em ${v.publishedAt}'),
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
