import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../auth/auth_provider.dart';
import 'catalog_provider.dart';

class ExerciseDetailScreen extends ConsumerStatefulWidget {
  final String exerciseId;

  const ExerciseDetailScreen({super.key, required this.exerciseId});

  @override
  ConsumerState<ExerciseDetailScreen> createState() =>
      _ExerciseDetailScreenState();
}

class _ExerciseDetailScreenState extends ConsumerState<ExerciseDetailScreen> {
  final Map<String, String> _signedUrls = {};
  bool _uploading = false;

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

  int _maxBytesForType(String mediaType) {
    switch (mediaType) {
      case 'VIDEO':
        return 50 * 1024 * 1024;
      case 'GIF':
        return 15 * 1024 * 1024;
      default:
        return 10 * 1024 * 1024;
    }
  }

  Future<void> _pickAndUploadMedia() async {
    final result = await FilePicker.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['mp4', 'webm', 'mov', 'jpg', 'jpeg', 'png', 'webp', 'gif'],
      withData: true,
    );
    if (result == null || result.files.single.bytes == null) return;

    final file = result.files.single;
    final resolved = _resolveMimeAndType(file.name).split('|');
    final mimeType = resolved[0];
    final mediaType = resolved[1];
    final maxBytes = _maxBytesForType(mediaType);

    if (file.bytes!.length > maxBytes) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Arquivo excede o limite para $mediaType.')),
      );
      return;
    }

    setState(() => _uploading = true);
    try {
      await ref.read(catalogServiceProvider).uploadExerciseMedia(
            exerciseId: widget.exerciseId,
            filename: file.name,
            mimeType: mimeType,
            mediaType: mediaType,
            bytes: file.bytes!,
          );
      ref.invalidate(catalogDetailProvider(widget.exerciseId));
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Mídia enviada e confirmada com sucesso!'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Falha ao enviar mídia: $e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _uploading = false);
    }
  }

  Future<void> _loadSignedUrl(CatalogMediaItem media) async {
    try {
      final res = await ref
          .read(catalogServiceProvider)
          .getSignedMediaUrl(widget.exerciseId, media.id);
      if (mounted) {
        setState(() {
          _signedUrls[media.id] = res.signedUrl;
        });
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Erro ao gerar URL assinada: $e')),
        );
      }
    }
  }

  Future<void> _confirmDelete() async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Remover Exercício Customizado'),
        content: const Text(
          'Deseja remover este exercício da biblioteca da sua organização?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancelar'),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Remover'),
          ),
        ],
      ),
    );
    if (ok != true) return;

    try {
      await ref.read(catalogServiceProvider).deleteCustomExercise(widget.exerciseId);
      ref.invalidate(catalogSearchProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Exercício removido.')),
        );
        context.pop();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Erro ao remover exercício: $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final isTrainer = ref.watch(authProvider.select((s) => s.role)) == 'TRAINER';
    final detailAsync = ref.watch(catalogDetailProvider(widget.exerciseId));

    return Scaffold(
      backgroundColor: const Color(0xFFF5F7FA),
      appBar: AppBar(
        title: const Text('Detalhes do Exercício'),
        backgroundColor: Colors.white,
        foregroundColor: Colors.black87,
        elevation: 1,
        actions: detailAsync.maybeWhen(
          data: (detail) => isTrainer && detail.summary.isCustom
              ? [
                  IconButton(
                    icon: const Icon(Icons.delete_outline, color: Colors.red),
                    tooltip: 'Remover exercício customizado',
                    onPressed: _confirmDelete,
                  ),
                ]
              : null,
          orElse: () => null,
        ),
      ),
      body: detailAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.error_outline, color: Colors.red, size: 48),
              const SizedBox(height: 12),
              Text('Erro ao carregar detalhes:\n$err', textAlign: TextAlign.center),
              const SizedBox(height: 12),
              ElevatedButton(
                onPressed: () => ref.invalidate(catalogDetailProvider(widget.exerciseId)),
                child: const Text('Tentar novamente'),
              ),
            ],
          ),
        ),
        data: (detail) {
          final ex = detail.summary;
          return ListView(
            padding: const EdgeInsets.all(20),
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
                      Text(
                        ex.namePT,
                        style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
                      ),
                      if (ex.nameEN.isNotEmpty && ex.nameEN != ex.namePT)
                        Padding(
                          padding: const EdgeInsets.only(top: 4),
                          child: Text(
                            ex.nameEN,
                            style: const TextStyle(color: Colors.black54),
                          ),
                        ),
                      const SizedBox(height: 12),
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: [
                          Chip(label: Text('Músculo: ${ex.primaryMusclePT}')),
                          Chip(label: Text('Equipamento: ${ex.equipmentPT}')),
                          Chip(label: Text('Categoria: ${ex.categoryPT}')),
                          Chip(label: Text('Nível: ${ex.level}')),
                        ],
                      ),
                      if (ex.secondaryMuscles.isNotEmpty) ...[
                        const SizedBox(height: 10),
                        Text(
                          'Músculos secundários: ${ex.secondaryMuscles.join(', ')}',
                          style: const TextStyle(fontSize: 13, color: Colors.black54),
                        ),
                      ],
                      const SizedBox(height: 10),
                      Text(
                        'Fonte / Licença: ${ex.licenseInfo}',
                        style: const TextStyle(fontSize: 12, color: Colors.black45),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),
              const Text(
                'Instruções de Execução (PT-BR)',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              if (detail.instructionsPT.isEmpty)
                const Text(
                  'Nenhuma instrução cadastrada.',
                  style: TextStyle(color: Colors.black54),
                )
              else
                ...detail.instructionsPT.asMap().entries.map(
                      (entry) => Card(
                        elevation: 0,
                        margin: const EdgeInsets.only(bottom: 8),
                        child: ListTile(
                          leading: CircleAvatar(
                            radius: 14,
                            backgroundColor: Colors.blue.shade50,
                            child: Text(
                              '${entry.key + 1}',
                              style: const TextStyle(
                                color: Colors.blue,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                          title: Text(entry.value),
                        ),
                      ),
                    ),
              const SizedBox(height: 20),
              Wrap(
                alignment: WrapAlignment.spaceBetween,
                crossAxisAlignment: WrapCrossAlignment.center,
                spacing: 8,
                runSpacing: 8,
                children: [
                  const Text(
                    'Mídias Demonstrativas',
                    style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                  ),
                  if (isTrainer && ex.isCustom)
                    ElevatedButton.icon(
                      onPressed: _uploading ? null : _pickAndUploadMedia,
                      icon: const Icon(Icons.upload_file),
                      label: Text(_uploading ? 'Enviando...' : 'Adicionar Mídia'),
                    ),
                ],
              ),
              const SizedBox(height: 10),
              if (detail.media.isEmpty)
                const Text(
                  'Nenhuma mídia demonstrativa confirmada ainda.',
                  style: TextStyle(color: Colors.black54),
                )
              else
                ...detail.media.map((m) => _buildMediaCard(m)),
            ],
          );
        },
      ),
    );
  }

  Widget _buildMediaCard(CatalogMediaItem media) {
    final signedUrl = _signedUrls[media.id];
    final sizeKb = (media.sizeBytes / 1024).toStringAsFixed(1);

    return Card(
      elevation: 0,
      margin: const EdgeInsets.only(bottom: 10),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(10),
        side: const BorderSide(color: Color(0xFFE4E9EF)),
      ),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  media.mediaType == 'VIDEO' ? Icons.videocam : Icons.image,
                  color: Colors.blue,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    media.filename,
                    style: const TextStyle(fontWeight: FontWeight.w600),
                  ),
                ),
                Text(
                  '${media.mediaType} • $sizeKb KB',
                  style: const TextStyle(fontSize: 12, color: Colors.black54),
                ),
              ],
            ),
            const SizedBox(height: 8),
            if (signedUrl == null)
              OutlinedButton.icon(
                onPressed: () => _loadSignedUrl(media),
                icon: const Icon(Icons.visibility),
                label: Text(
                  media.mediaType == 'VIDEO' ? 'Carregar URL do Vídeo' : 'Visualizar Mídia',
                ),
              )
            else if (media.mediaType == 'IMAGE' || media.mediaType == 'GIF')
              ClipRRect(
                borderRadius: BorderRadius.circular(8),
                child: Image.network(
                  signedUrl,
                  height: 220,
                  width: double.infinity,
                  fit: BoxFit.cover,
                  errorBuilder: (_, _, _) => SelectableText(
                    'URL assinada gerada: $signedUrl',
                    style: const TextStyle(fontSize: 12, color: Colors.blueGrey),
                  ),
                ),
              )
            else
              SelectableText(
                'URL assinada de vídeo pronta (expira em minutos):\n$signedUrl',
                style: const TextStyle(fontSize: 12, color: Colors.blue),
              ),
          ],
        ),
      ),
    );
  }
}
