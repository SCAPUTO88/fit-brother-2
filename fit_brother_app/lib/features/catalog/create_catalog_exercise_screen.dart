import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'catalog_provider.dart';

class CreateCatalogExerciseScreen extends ConsumerStatefulWidget {
  const CreateCatalogExerciseScreen({super.key});

  @override
  ConsumerState<CreateCatalogExerciseScreen> createState() =>
      _CreateCatalogExerciseScreenState();
}

class _CreateCatalogExerciseScreenState
    extends ConsumerState<CreateCatalogExerciseScreen> {
  final _formKey = GlobalKey<FormState>();

  String _namePT = '';
  String _primaryMuscle = 'chest';
  String _equipment = 'barbell';
  String _category = 'strength';
  String _level = 'intermediate';
  String _secondaryRaw = '';
  String _instructionsRaw = '';

  Uint8List? _mediaBytes;
  String? _mediaFilename;
  bool _saving = false;

  String _resolveMimeAndType(String filename) {
    final lower = filename.toLowerCase();
    if (lower.endsWith('.webm')) return 'video/webm|VIDEO';
    if (lower.endsWith('.mov')) return 'video/quicktime|VIDEO';
    if (lower.endsWith('.mp4')) return 'video/mp4|VIDEO';
    if (lower.endsWith('.gif')) return 'image/gif|GIF';
    if (lower.endsWith('.png')) return 'image/png|IMAGE';
    if (lower.endsWith('.webp')) return 'image/webp|IMAGE';
    return 'image/jpeg|IMAGE';
  }

  Future<void> _pickMedia() async {
    final result = await FilePicker.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['mp4', 'webm', 'mov', 'jpg', 'jpeg', 'png', 'webp', 'gif'],
      withData: true,
    );
    if (result == null || result.files.single.bytes == null) return;

    final file = result.files.single;
    final parts = _resolveMimeAndType(file.name).split('|');
    final maxMb = parts[1] == 'VIDEO' ? 50 : (parts[1] == 'GIF' ? 15 : 10);
    if (file.bytes!.length > maxMb * 1024 * 1024) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Arquivo excede ${maxMb}MB para ${parts[1]}.')),
      );
      return;
    }

    setState(() {
      _mediaBytes = file.bytes;
      _mediaFilename = file.name;
    });
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    _formKey.currentState!.save();

    final secondary = _secondaryRaw
        .split(',')
        .map((s) => s.trim().toLowerCase())
        .where((s) => s.isNotEmpty)
        .toList();
    final instructions = _instructionsRaw
        .split('\n')
        .map((s) => s.trim())
        .where((s) => s.isNotEmpty)
        .toList();

    setState(() => _saving = true);
    try {
      final service = ref.read(catalogServiceProvider);
      final created = await service.createCustomExercise(
        namePT: _namePT.trim(),
        primaryMuscle: _primaryMuscle,
        secondaryMuscles: secondary,
        equipment: _equipment,
        category: _category,
        level: _level,
        instructionsPT: instructions,
      );

      if (_mediaBytes != null && _mediaFilename != null) {
        final parts = _resolveMimeAndType(_mediaFilename!).split('|');
        await service.uploadExerciseMedia(
          exerciseId: created.summary.id,
          filename: _mediaFilename!,
          mimeType: parts[0],
          mediaType: parts[1],
          bytes: _mediaBytes!,
        );
      }

      ref.invalidate(catalogSearchProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Exercício customizado criado com sucesso!'),
            backgroundColor: Colors.green,
          ),
        );
        context.pop();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Erro ao criar exercício: $e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final filtersAsync = ref.watch(catalogFiltersProvider);

    return Scaffold(
      backgroundColor: const Color(0xFFF5F7FA),
      appBar: AppBar(
        title: const Text('Novo Exercício Customizado'),
        backgroundColor: Colors.white,
        foregroundColor: Colors.black87,
        elevation: 1,
      ),
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(20),
          child: Container(
            constraints: const BoxConstraints(maxWidth: 540),
            child: Form(
              key: _formKey,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  TextFormField(
                    decoration: const InputDecoration(
                      labelText: 'Nome do Exercício (PT-BR)',
                      filled: true,
                      fillColor: Colors.white,
                      border: OutlineInputBorder(),
                    ),
                    validator: (v) => v == null || v.trim().length < 2
                        ? 'Informe ao menos 2 caracteres'
                        : null,
                    onSaved: (v) => _namePT = v ?? '',
                  ),
                  const SizedBox(height: 14),
                  filtersAsync.when(
                    loading: () => const LinearProgressIndicator(),
                    error: (_, _) => const Text('Usando códigos padrão para filtros.'),
                    data: (filters) => Column(
                      children: [
                        DropdownButtonFormField<String>(
                          initialValue: filters.muscleGroups.any((m) => m.code == _primaryMuscle)
                              ? _primaryMuscle
                              : (filters.muscleGroups.isNotEmpty
                                  ? filters.muscleGroups.first.code
                                  : null),
                          decoration: const InputDecoration(
                            labelText: 'Grupo Muscular Principal',
                            filled: true,
                            fillColor: Colors.white,
                            border: OutlineInputBorder(),
                          ),
                          items: filters.muscleGroups
                              .map((m) => DropdownMenuItem(value: m.code, child: Text(m.namePT)))
                              .toList(),
                          onChanged: (v) => setState(() => _primaryMuscle = v ?? 'chest'),
                        ),
                        const SizedBox(height: 14),
                        DropdownButtonFormField<String>(
                          initialValue: filters.equipments.any((e) => e.code == _equipment)
                              ? _equipment
                              : (filters.equipments.isNotEmpty
                                  ? filters.equipments.first.code
                                  : null),
                          decoration: const InputDecoration(
                            labelText: 'Equipamento',
                            filled: true,
                            fillColor: Colors.white,
                            border: OutlineInputBorder(),
                          ),
                          items: filters.equipments
                              .map((e) => DropdownMenuItem(value: e.code, child: Text(e.namePT)))
                              .toList(),
                          onChanged: (v) => setState(() => _equipment = v ?? 'barbell'),
                        ),
                        const SizedBox(height: 14),
                        DropdownButtonFormField<String>(
                          initialValue: filters.categories.any((c) => c.code == _category)
                              ? _category
                              : (filters.categories.isNotEmpty
                                  ? filters.categories.first.code
                                  : null),
                          decoration: const InputDecoration(
                            labelText: 'Categoria',
                            filled: true,
                            fillColor: Colors.white,
                            border: OutlineInputBorder(),
                          ),
                          items: filters.categories
                              .map((c) => DropdownMenuItem(value: c.code, child: Text(c.namePT)))
                              .toList(),
                          onChanged: (v) => setState(() => _category = v ?? 'strength'),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 14),
                  DropdownButtonFormField<String>(
                    initialValue: _level,
                    decoration: const InputDecoration(
                      labelText: 'Nível',
                      filled: true,
                      fillColor: Colors.white,
                      border: OutlineInputBorder(),
                    ),
                    items: const [
                      DropdownMenuItem(value: 'beginner', child: Text('Iniciante')),
                      DropdownMenuItem(value: 'intermediate', child: Text('Intermediário')),
                      DropdownMenuItem(value: 'expert', child: Text('Avançado')),
                    ],
                    onChanged: (v) => setState(() => _level = v ?? 'intermediate'),
                  ),
                  const SizedBox(height: 14),
                  TextFormField(
                    decoration: const InputDecoration(
                      labelText: 'Músculos Secundários (separados por vírgula)',
                      filled: true,
                      fillColor: Colors.white,
                      border: OutlineInputBorder(),
                    ),
                    onSaved: (v) => _secondaryRaw = v ?? '',
                  ),
                  const SizedBox(height: 14),
                  TextFormField(
                    decoration: const InputDecoration(
                      labelText: 'Instruções de Execução (1 passo por linha)',
                      filled: true,
                      fillColor: Colors.white,
                      border: OutlineInputBorder(),
                    ),
                    maxLines: 4,
                    onSaved: (v) => _instructionsRaw = v ?? '',
                  ),
                  const SizedBox(height: 16),
                  OutlinedButton.icon(
                    onPressed: _pickMedia,
                    icon: const Icon(Icons.attach_file),
                    label: Text(
                      _mediaFilename == null
                          ? 'Anexar Vídeo, Foto ou GIF (Opcional)'
                          : 'Selecionado: $_mediaFilename',
                    ),
                  ),
                  const SizedBox(height: 20),
                  SizedBox(
                    height: 48,
                    child: ElevatedButton(
                      onPressed: _saving ? null : _submit,
                      child: _saving
                          ? const CircularProgressIndicator(color: Colors.white)
                          : const Text('Salvar Exercício Customizado'),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
