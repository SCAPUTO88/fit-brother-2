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

  Future<void> _pickVideo() async {
    // Pick video specifically on Web (returns bytes)
    FilePickerResult? result = await FilePicker.pickFiles(
      type: FileType.video,
      withData: true,
    );

    if (result != null && result.files.single.bytes != null) {
      // Basic check to see if it's too large (ex: > 50MB)
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
      String finalVideoUrl = '';

      // 1. Lógica de Upload do Vídeo via R2 (Presigned URL)
      if (_videoBytes != null && _videoFileName != null) {
        // Passo 1: Solicitar URL pré-assinada
        final reqResponse = await dio.post(
          '/videos/request-upload',
          data: {
            'filename': _videoFileName,
            'content_type':
                'video/mp4', // Genérico, o ideal seria inspecionar o arquivo real
          },
        );

        final uploadUrl = reqResponse.data['upload_url'];
        final videoId = reqResponse.data['video_id'];
        final publicUrl = reqResponse.data['public_url'];

        // Passo 2: Fazer o upload binário direto para a Cloudflare R2
        final uploadDio =
            Dio(); // Dio virgem sem os interceptors de Headers de Auth para evitar conflitos com a AWS/R2
        await uploadDio.put(
          uploadUrl,
          data: Stream.fromIterable([_videoBytes!]),
          options: Options(
            headers: {
              'Content-Type': 'video/mp4',
              'Content-Length': _videoBytes!.length.toString(),
            },
          ),
        );

        // Passo 3: Confirmar Upload na API
        await dio.post('/videos/confirm', data: {'video_id': videoId});

        finalVideoUrl = publicUrl;
      }

      // 2. Salvar Exercício
      await dio.post(
        '/workouts/sessions/${widget.sessionId}/exercises',
        data: {
          'workout_session_id': widget.sessionId,
          'name': _name,
          'sets': _sets,
          'reps': _reps,
          'rest_seconds': _restSeconds,
          'notes': _notes,
          'video_url': finalVideoUrl,
          'position': _position,
        },
      );

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
                          'Máx 50MB. MP4 recomendado.',
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
