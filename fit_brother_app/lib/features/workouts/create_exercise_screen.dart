import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:file_picker/file_picker.dart';
import 'package:dio/dio.dart';

import 'workouts_provider.dart';
import '../auth/auth_provider.dart';

class CreateExerciseScreen extends ConsumerStatefulWidget {
  final String planId;
  final String sessionId;

  const CreateExerciseScreen({
    super.key,
    required this.planId,
    required this.sessionId,
  });

  @override
  ConsumerState<CreateExerciseScreen> createState() =>
      _CreateExerciseScreenState();
}

class _CreateExerciseScreenState extends ConsumerState<CreateExerciseScreen> {
  final _formKey = GlobalKey<FormState>();

  String _name = '';
  int _sets = 0;
  String _reps = '';
  int _restSeconds = 0;
  String _notes = '';
  int _position = 1;

  Uint8List? _videoBytes;
  String? _videoFileName;
  bool _isLoading = false;

  String _resolveMimeType(String filename) {
    final lower = filename.toLowerCase();
    if (lower.endsWith('.webm')) return 'video/webm';
    if (lower.endsWith('.mov')) return 'video/quicktime';
    return 'video/mp4';
  }

  Future<void> _pickVideo() async {
    final FilePickerResult? result = await FilePicker.pickFiles(
      type: FileType.video,
      withData: true,
    );

    if (result != null && result.files.single.bytes != null) {
      final sizeMb = result.files.single.bytes!.length / (1024 * 1024);
      if (sizeMb > 50) {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('O vídeo não pode ter mais de 50MB!')),
        );
        return;
      }

      setState(() {
        _videoBytes = result.files.single.bytes;
        _videoFileName = result.files.single.name;
      });
    }
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    _formKey.currentState!.save();

    setState(() => _isLoading = true);

    try {
      final dio = ref.read(dioClientProvider).dio;

      // 1. Criar o exercício primeiro para obter exercise_id válido para vínculo do vídeo
      final exResponse = await dio.post(
        '/workouts/sessions/${widget.sessionId}/exercises',
        data: {
          'workout_session_id': widget.sessionId,
          'name': _name,
          'sets': _sets,
          'reps': _reps,
          'rest_seconds': _restSeconds,
          'notes': _notes,
          'video_url': '',
          'position': _position,
        },
      );

      final exerciseId = (exResponse.data as Map)['id'] as String?;

      // 2. Se houver vídeo selecionado, executar fluxo alinhado ao contrato da API Go
      if (_videoBytes != null &&
          _videoFileName != null &&
          exerciseId != null &&
          exerciseId.isNotEmpty) {
        final mimeType = _resolveMimeType(_videoFileName!);

        // Passo A: Solicitar intenção de upload e URL pré-assinada
        final reqResponse = await dio.post(
          '/videos/upload-url',
          data: {
            'exercise_id': exerciseId,
            'filename': _videoFileName,
            'mime_type': mimeType,
          },
        );

        final reqData = reqResponse.data as Map;
        final presignedUrl = reqData['presigned_url'] as String;
        final r2Key = reqData['r2_key'] as String;

        // Passo B: Upload binário direto para a presigned URL (sem header Authorization da API)
        final uploadDio = Dio();
        await uploadDio.put(
          presignedUrl,
          data: _videoBytes!,
          options: Options(
            headers: {
              'Content-Type': mimeType,
              'Content-Length': _videoBytes!.length,
            },
          ),
        );

        // Passo C: Confirmar upload na API para validação HeadObject e persistência CONFIRMED
        await dio.post(
          '/videos',
          data: {
            'exercise_id': exerciseId,
            'r2_key': r2Key,
            'filename': _videoFileName,
            'size_bytes': _videoBytes!.length,
            'mime_type': mimeType,
          },
        );
      }

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Exercício criado com sucesso!'),
            backgroundColor: Colors.green,
          ),
        );
        ref.invalidate(planDetailsProvider(widget.planId));
        context.pop();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Erro ao salvar exercício ou vídeo: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Novo Exercício')),
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24.0),
          child: Container(
            constraints: const BoxConstraints(maxWidth: 500),
            child: Form(
              key: _formKey,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  TextFormField(
                    decoration: const InputDecoration(
                      labelText: 'Nome do Exercício (Ex: Supino Reto)',
                      border: OutlineInputBorder(),
                    ),
                    validator: (val) => val == null || val.length < 2
                        ? 'Mínimo de 2 caracteres'
                        : null,
                    onSaved: (val) => _name = val ?? '',
                  ),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Expanded(
                        child: TextFormField(
                          decoration: const InputDecoration(
                            labelText: 'Séries',
                            border: OutlineInputBorder(),
                          ),
                          keyboardType: TextInputType.number,
                          validator: (val) =>
                              val == null || int.tryParse(val) == null
                                  ? 'Inválido'
                                  : null,
                          onSaved: (val) => _sets = int.parse(val!),
                        ),
                      ),
                      const SizedBox(width: 16),
                      Expanded(
                        child: TextFormField(
                          decoration: const InputDecoration(
                            labelText: 'Repetições (Ex: 10-12)',
                            border: OutlineInputBorder(),
                          ),
                          onSaved: (val) => _reps = val ?? '',
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Expanded(
                        child: TextFormField(
                          decoration: const InputDecoration(
                            labelText: 'Descanso em segundos (Ex: 60)',
                            border: OutlineInputBorder(),
                          ),
                          keyboardType: TextInputType.number,
                          validator: (val) =>
                              val == null || int.tryParse(val) == null
                                  ? 'Inválido'
                                  : null,
                          onSaved: (val) => _restSeconds = int.parse(val!),
                        ),
                      ),
                      const SizedBox(width: 16),
                      Expanded(
                        child: TextFormField(
                          decoration: const InputDecoration(
                            labelText: 'Posição na fila (Ex: 1)',
                            border: OutlineInputBorder(),
                          ),
                          keyboardType: TextInputType.number,
                          initialValue: '1',
                          validator: (val) =>
                              val == null || int.tryParse(val) == null
                                  ? 'Inválido'
                                  : null,
                          onSaved: (val) => _position = int.parse(val!),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    decoration: const InputDecoration(
                      labelText: 'Instruções Adicionais',
                      border: OutlineInputBorder(),
                    ),
                    maxLines: 2,
                    onSaved: (val) => _notes = val ?? '',
                  ),
                  const SizedBox(height: 24),

                  // Bloco de Upload do Vídeo
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      border: Border.all(color: Colors.grey.shade300),
                      borderRadius: BorderRadius.circular(8),
                      color: Colors.blue.withValues(alpha: 0.05),
                    ),
                    child: Column(
                      children: [
                        const Icon(
                          Icons.video_camera_back,
                          size: 48,
                          color: Colors.blue,
                        ),
                        const SizedBox(height: 16),
                        const Text(
                          'Demonstração (Opcional)',
                          style: TextStyle(fontWeight: FontWeight.bold),
                        ),
                        const SizedBox(height: 8),
                        if (_videoFileName != null)
                          Padding(
                            padding: const EdgeInsets.only(bottom: 12.0),
                            child: Text(
                              '✅ $_videoFileName',
                              style: const TextStyle(color: Colors.green),
                            ),
                          ),
                        ElevatedButton.icon(
                          onPressed: _pickVideo,
                          icon: const Icon(Icons.attach_file),
                          label: const Text('Selecionar Vídeo'),
                        ),
                        const SizedBox(height: 8),
                        const Text(
                          'Máx 50MB. MP4, WebM ou MOV.',
                          style: TextStyle(fontSize: 12, color: Colors.grey),
                        ),
                      ],
                    ),
                  ),

                  const SizedBox(height: 32),
                  SizedBox(
                    height: 48,
                    child: ElevatedButton(
                      onPressed: _isLoading ? null : _submit,
                      child: _isLoading
                          ? const CircularProgressIndicator(color: Colors.white)
                          : const Text(
                              'Salvar Exercício',
                              style: TextStyle(fontSize: 16),
                            ),
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
