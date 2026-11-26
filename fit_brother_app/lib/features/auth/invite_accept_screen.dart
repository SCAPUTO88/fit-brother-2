import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/api/dio_client.dart';
import 'auth_provider.dart';

class InviteAcceptScreen extends ConsumerStatefulWidget {
  final String? initialToken;

  const InviteAcceptScreen({super.key, this.initialToken});

  @override
  ConsumerState<InviteAcceptScreen> createState() => _InviteAcceptScreenState();
}

class _InviteAcceptScreenState extends ConsumerState<InviteAcceptScreen> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _tokenController;
  final _passwordController = TextEditingController();
  final _confirmController = TextEditingController();
  bool _consentTerms = false;
  bool _consentHealth = false;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _tokenController = TextEditingController(text: widget.initialToken ?? '');
  }

  @override
  void dispose() {
    _tokenController.dispose();
    _passwordController.dispose();
    _confirmController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    if (!_consentTerms || !_consentHealth) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Aceite os Termos de Uso e o consentimento de dados de saúde (LGPD).'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    setState(() => _isLoading = true);
    try {
      final client = ref.read(dioClientProvider);
      final res = await client.dio.post<dynamic>(
        '/auth/invitations/accept',
        data: {
          'token': _tokenController.text.trim(),
          'password': _passwordController.text,
          'consent_terms': _consentTerms,
          'consent_health_data': _consentHealth,
          'policy_version': '2026-11',
        },
        options: client.useCookieSession
            ? Options(headers: {DioClient.sessionModeHeaderName: 'cookie'})
            : null,
      );

      final data = res.data is Map
          ? Map<String, dynamic>.from(res.data as Map)
          : <String, dynamic>{};
      final session = data['session'] is Map
          ? Map<String, dynamic>.from(data['session'] as Map)
          : <String, dynamic>{};
      await ref.read(authProvider.notifier).applySessionFromPayload(session);

      if (!mounted) return;
      if (ref.read(authProvider).isLoggedIn) {
        context.go('/dashboard');
      }
    } on DioException catch (e) {
      final msg = (e.response?.data as Map?)?['error']?.toString() ?? 'Falha ao ativar convite';
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(msg), backgroundColor: Colors.red),
        );
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Erro inesperado ao ativar convite.'), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Ativar Convite')),
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Container(
            constraints: const BoxConstraints(maxWidth: 460),
            child: Form(
              key: _formKey,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const Text(
                    'Defina sua senha e registre seu consentimento LGPD para acessar seus treinos.',
                    style: TextStyle(color: Colors.black54),
                  ),
                  const SizedBox(height: 20),
                  TextFormField(
                    controller: _tokenController,
                    decoration: const InputDecoration(
                      labelText: 'Token do Convite',
                      border: OutlineInputBorder(),
                    ),
                    validator: (v) => (v == null || v.trim().isEmpty) ? 'Token obrigatório' : null,
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _passwordController,
                    decoration: const InputDecoration(
                      labelText: 'Nova Senha (mín. 8 caracteres)',
                      border: OutlineInputBorder(),
                    ),
                    obscureText: true,
                    validator: (v) => (v == null || v.length < 8) ? 'Mínimo de 8 caracteres' : null,
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _confirmController,
                    decoration: const InputDecoration(
                      labelText: 'Confirmar Senha',
                      border: OutlineInputBorder(),
                    ),
                    obscureText: true,
                    validator: (v) =>
                        v != _passwordController.text ? 'As senhas não conferem' : null,
                  ),
                  const SizedBox(height: 16),
                  CheckboxListTile(
                    contentPadding: EdgeInsets.zero,
                    value: _consentTerms,
                    onChanged: (v) => setState(() => _consentTerms = v ?? false),
                    title: const Text('Li e aceito os Termos de Uso e Política de Privacidade'),
                  ),
                  CheckboxListTile(
                    contentPadding: EdgeInsets.zero,
                    value: _consentHealth,
                    onChanged: (v) => setState(() => _consentHealth = v ?? false),
                    title: const Text('Autorizo o tratamento dos meus dados de saúde e treinos (LGPD)'),
                  ),
                  const SizedBox(height: 20),
                  ElevatedButton(
                    onPressed: _isLoading ? null : _submit,
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(vertical: 16),
                    ),
                    child: _isLoading
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('ATIVAR CONTA'),
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
