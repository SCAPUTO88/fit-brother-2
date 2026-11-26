import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../auth/auth_provider.dart';

class PrivacyCenterScreen extends ConsumerStatefulWidget {
  const PrivacyCenterScreen({super.key});

  @override
  ConsumerState<PrivacyCenterScreen> createState() => _PrivacyCenterScreenState();
}

class _PrivacyCenterScreenState extends ConsumerState<PrivacyCenterScreen> {
  bool _loading = true;
  bool _busy = false;
  List<Map<String, dynamic>> _consents = const [];
  List<Map<String, dynamic>> _requests = const [];
  String? _exportedJson;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final dio = ref.read(dioClientProvider).dio;
      final res = await Future.wait([
        dio.get<dynamic>('/privacy/consents'),
        dio.get<dynamic>('/privacy/requests'),
      ]);
      if (!mounted) return;
      setState(() {
        _consents = _castList(res[0].data);
        _requests = _castList(res[1].data);
        _loading = false;
      });
    } on DioException catch (e) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = (e.response?.data as Map?)?['error']?.toString() ?? 'Falha ao carregar Central LGPD';
      });
    }
  }

  List<Map<String, dynamic>> _castList(dynamic raw) {
    if (raw is! List) return const [];
    return raw.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<void> _exportData() async {
    setState(() => _busy = true);
    try {
      final res = await ref.read(dioClientProvider).dio.get<dynamic>('/privacy/export');
      final pretty = const JsonEncoder.withIndent('  ').convert(res.data);
      if (!mounted) return;
      setState(() => _exportedJson = pretty);
      await _load();
    } on DioException catch (e) {
      _showSnack((e.response?.data as Map?)?['error']?.toString() ?? 'Falha ao exportar dados', true);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _recordConsent(String purpose, bool granted) async {
    setState(() => _busy = true);
    try {
      await ref.read(dioClientProvider).dio.post<dynamic>(
        '/privacy/consents',
        data: {
          'purpose': purpose,
          'policy_version': '2026-11',
          'granted': granted,
        },
      );
      await _load();
      _showSnack('Consentimento registrado.', false);
    } on DioException catch (e) {
      _showSnack((e.response?.data as Map?)?['error']?.toString() ?? 'Erro ao atualizar consentimento', true);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _confirmAnonymize() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Excluir / Anonimizar Conta (LGPD)'),
        content: const Text(
          'Esta ação remove seus dados identificáveis, encerra vínculos ativos e revoga todas as sessões imediatamente. Deseja continuar?',
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Cancelar')),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red, foregroundColor: Colors.white),
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Confirmar Exclusão'),
          ),
        ],
      ),
    );
    if (confirmed != true) return;

    setState(() => _busy = true);
    try {
      await ref.read(dioClientProvider).dio.post<dynamic>('/privacy/anonymize');
      await ref.read(authProvider.notifier).clearLocalSession();
      if (!mounted) return;
      context.go('/login');
    } on DioException catch (e) {
      _showSnack((e.response?.data as Map?)?['error']?.toString() ?? 'Falha ao anonimizar conta', true);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  void _showSnack(String msg, bool isError) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(msg), backgroundColor: isError ? Colors.red : Colors.green),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Central de Privacidade & LGPD')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(child: Text(_error!, style: const TextStyle(color: Colors.red)))
              : SingleChildScrollView(
                  padding: const EdgeInsets.all(24),
                  child: Center(
                    child: Container(
                      constraints: const BoxConstraints(maxWidth: 720),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Wrap(
                            spacing: 12,
                            runSpacing: 12,
                            children: [
                              ElevatedButton.icon(
                                onPressed: _busy ? null : _exportData,
                                icon: const Icon(Icons.download),
                                label: const Text('Exportar Meus Dados (JSON)'),
                              ),
                              OutlinedButton.icon(
                                onPressed: _busy
                                    ? null
                                    : () => _recordConsent('MEDIA_UPLOAD', true),
                                icon: const Icon(Icons.check_circle_outline),
                                label: const Text('Autorizar Mídia de Treino'),
                              ),
                            ],
                          ),
                          if (_exportedJson != null) ...[
                            const SizedBox(height: 20),
                            const Text('Exportação LGPD', style: TextStyle(fontWeight: FontWeight.bold)),
                            const SizedBox(height: 8),
                            Container(
                              padding: const EdgeInsets.all(12),
                              decoration: BoxDecoration(
                                color: Colors.grey.shade100,
                                borderRadius: BorderRadius.circular(8),
                                border: Border.all(color: Colors.grey.shade300),
                              ),
                              child: SelectableText(
                                _exportedJson!,
                                style: const TextStyle(fontFamily: 'monospace', fontSize: 12),
                              ),
                            ),
                          ],
                          const SizedBox(height: 24),
                          const Text('Consentimentos Registrados', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                          const SizedBox(height: 8),
                          if (_consents.isEmpty)
                            const Text('Nenhum registro de consentimento ainda.')
                          else
                            ..._consents.map(
                              (c) => Card(
                                child: ListTile(
                                  leading: Icon(
                                    c['granted'] == true ? Icons.verified : Icons.block,
                                    color: c['granted'] == true ? Colors.green : Colors.orange,
                                  ),
                                  title: Text('${c['purpose']} (v${c['policy_version']})'),
                                  subtitle: Text('Registrado em: ${c['recorded_at']}'),
                                ),
                              ),
                            ),
                          const SizedBox(height: 24),
                          const Text('Solicitações LGPD', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                          const SizedBox(height: 8),
                          if (_requests.isEmpty)
                            const Text('Nenhuma solicitação registrada.')
                          else
                            ..._requests.map(
                              (r) => Card(
                                child: ListTile(
                                  title: Text('${r['request_type']} — ${r['status']}'),
                                  subtitle: Text('Solicitado em: ${r['requested_at']}'),
                                ),
                              ),
                            ),
                          const SizedBox(height: 32),
                          ElevatedButton.icon(
                            style: ElevatedButton.styleFrom(
                              backgroundColor: Colors.red.shade700,
                              foregroundColor: Colors.white,
                              padding: const EdgeInsets.symmetric(vertical: 14),
                            ),
                            onPressed: _busy ? null : _confirmAnonymize,
                            icon: const Icon(Icons.delete_forever),
                            label: const Text('EXCLUIR / ANONIMIZAR MINHA CONTA'),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
    );
  }
}
