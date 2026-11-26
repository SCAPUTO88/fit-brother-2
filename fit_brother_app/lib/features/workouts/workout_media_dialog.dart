import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'workouts_provider.dart';

Future<void> showExerciseMediaDialog({
  required BuildContext context,
  required WidgetRef ref,
  required String exerciseName,
  String? legacyExerciseId,
  String? catalogExerciseId,
  String? mediaId,
  String? fallbackVideoUrl,
}) {
  return showDialog<void>(
    context: context,
    builder: (ctx) => _ExerciseMediaDialog(
      ref: ref,
      exerciseName: exerciseName,
      legacyExerciseId: legacyExerciseId,
      catalogExerciseId: catalogExerciseId,
      mediaId: mediaId,
      fallbackVideoUrl: fallbackVideoUrl,
    ),
  );
}

class _ExerciseMediaDialog extends StatefulWidget {
  final WidgetRef ref;
  final String exerciseName;
  final String? legacyExerciseId;
  final String? catalogExerciseId;
  final String? mediaId;
  final String? fallbackVideoUrl;

  const _ExerciseMediaDialog({
    required this.ref,
    required this.exerciseName,
    this.legacyExerciseId,
    this.catalogExerciseId,
    this.mediaId,
    this.fallbackVideoUrl,
  });

  @override
  State<_ExerciseMediaDialog> createState() => _ExerciseMediaDialogState();
}

class _ExerciseMediaDialogState extends State<_ExerciseMediaDialog> {
  late Future<ResolvedExerciseMedia> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<ResolvedExerciseMedia> _load() {
    return widget.ref.read(workoutBuilderServiceProvider).resolveExerciseMedia(
          legacyExerciseId: widget.legacyExerciseId,
          catalogExerciseId: widget.catalogExerciseId,
          mediaId: widget.mediaId,
          fallbackVideoUrl: widget.fallbackVideoUrl,
        );
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(widget.exerciseName),
      content: SizedBox(
        width: 460,
        child: FutureBuilder<ResolvedExerciseMedia>(
          future: _future,
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return const SizedBox(
                height: 140,
                child: Center(child: CircularProgressIndicator()),
              );
            }
            if (snapshot.hasError) {
              return Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Não foi possível carregar a demonstração: ${snapshot.error}',
                    style: const TextStyle(color: Colors.red),
                  ),
                  const SizedBox(height: 12),
                  OutlinedButton.icon(
                    onPressed: () {
                      setState(() {
                        _future = _load();
                      });
                    },
                    icon: const Icon(Icons.refresh),
                    label: const Text('Tentar novamente'),
                  ),
                ],
              );
            }

            final media = snapshot.data!;
            return SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (media.signedUrl.isNotEmpty &&
                      (media.mediaType == 'IMAGE' || media.mediaType == 'GIF'))
                    ClipRRect(
                      borderRadius: BorderRadius.circular(8),
                      child: Image.network(
                        media.signedUrl,
                        height: 220,
                        width: double.infinity,
                        fit: BoxFit.cover,
                        errorBuilder: (_, _, _) => SelectableText(
                          'URL assinada gerada: ${media.signedUrl}',
                          style: const TextStyle(fontSize: 12, color: Colors.blue),
                        ),
                      ),
                    )
                  else if (media.signedUrl.isNotEmpty) ...[
                    Container(
                      width: double.infinity,
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.blue.withValues(alpha: 0.08),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Row(
                            children: [
                              Icon(Icons.play_circle_fill, color: Colors.blue),
                              SizedBox(width: 8),
                              Expanded(
                                child: Text(
                                  'URL assinada de vídeo gerada com segurança',
                                  style: TextStyle(fontWeight: FontWeight.bold),
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: 8),
                          SelectableText(
                            media.signedUrl,
                            style: const TextStyle(fontSize: 12, color: Colors.blue),
                          ),
                          const SizedBox(height: 8),
                          Text(
                            'Expira em ${media.expiresIn ~/ 60} min.',
                            style: const TextStyle(fontSize: 12, color: Colors.black54),
                          ),
                          const SizedBox(height: 8),
                          OutlinedButton.icon(
                            onPressed: () async {
                              await Clipboard.setData(ClipboardData(text: media.signedUrl));
                              if (context.mounted) {
                                ScaffoldMessenger.of(context).showSnackBar(
                                  const SnackBar(content: Text('URL copiada para a área de transferência.')),
                                );
                              }
                            },
                            icon: const Icon(Icons.copy, size: 16),
                            label: const Text('Copiar URL do Vídeo'),
                          ),
                        ],
                      ),
                    ),
                  ],
                  if (media.instructionsPT.isNotEmpty) ...[
                    const SizedBox(height: 16),
                    const Text(
                      'Instruções de Execução',
                      style: TextStyle(fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 6),
                    ...media.instructionsPT.asMap().entries.map(
                          (entry) => Padding(
                            padding: const EdgeInsets.only(bottom: 4),
                            child: Text('${entry.key + 1}. ${entry.value}'),
                          ),
                        ),
                  ],
                  if (media.signedUrl.isEmpty && media.instructionsPT.isEmpty)
                    const Text('Nenhuma mídia ou instrução cadastrada para este exercício.'),
                ],
              ),
            );
          },
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Fechar'),
        ),
      ],
    );
  }
}
