import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'auth_provider.dart';

class PasswordResetScreen extends ConsumerStatefulWidget {
  final String? initialToken;

  const PasswordResetScreen({super.key, this.initialToken});

  @override
  ConsumerState<PasswordResetScreen> createState() => _PasswordResetScreenState();
}

class _PasswordResetScreenState extends ConsumerState<PasswordResetScreen> {
  final _emailController = TextEditingController();
  late final TextEditingController _tokenController;
  final _newPasswordController = TextEditingController();
  bool _confirmStep = false;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    final token = widget.initialToken?.trim() ?? '';
    _tokenController = TextEditingController(text: token);
    _confirmStep = token.isNotEmpty;
  }

  @override
  void dispose() {
    _emailController.dispose();
    _tokenController.dispose();
    _newPasswordController.dispose();
    super.dispose();
  }

  Future<void> _requestReset() async {
    final email = _emailController.text.trim();
    if (email.isEmpty || !email.contains('@')) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Informe um e-mail válido.')),
      );
      return;
    }

    setState(() => _isLoading = true);
    try {
      await ref.read(dioClientProvider).dio.post<dynamic>(
        '/auth/password-reset/request',
        data: {'email': email},
      );
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Se o e-mail estiver cadastrado, enviaremos as instruções.'),
          backgroundColor: Colors.green,
        ),
      );
      setState(() => _confirmStep = true);
    } on DioException catch (e) {
      final msg = (e.response?.data as Map?)?['error']?.toString() ?? 'Falha ao solicitar redefinição';
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(msg), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _confirmReset() async {
    final token = _tokenController.text.trim();
    final pass = _newPasswordController.text;
    if (token.isEmpty || pass.length < 8) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Informe o token e uma nova senha de pelo menos 8 caracteres.')),
      );
      return;
    }

    setState(() => _isLoading = true);
    try {
      await ref.read(dioClientProvider).dio.post<dynamic>(
        '/auth/password-reset/confirm',
        data: {'token': token, 'new_password': pass},
      );
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Senha redefinida! Faça login com a nova senha.'),
          backgroundColor: Colors.green,
        ),
      );
      context.go('/login');
    } on DioException catch (e) {
      final msg = (e.response?.data as Map?)?['error']?.toString() ?? 'Falha ao redefinir senha';
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(msg), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Recuperar Senha')),
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Container(
            constraints: const BoxConstraints(maxWidth: 420),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                SegmentedButton<bool>(
                  segments: const [
                    ButtonSegment<bool>(value: false, label: Text('Solicitar')),
                    ButtonSegment<bool>(value: true, label: Text('Confirmar')),
                  ],
                  selected: {_confirmStep},
                  onSelectionChanged: (s) => setState(() => _confirmStep = s.first),
                ),
                const SizedBox(height: 24),
                if (!_confirmStep) ...[
                  TextField(
                    controller: _emailController,
                    decoration: const InputDecoration(
                      labelText: 'Seu e-mail cadastrado',
                      border: OutlineInputBorder(),
                    ),
                    keyboardType: TextInputType.emailAddress,
                  ),
                  const SizedBox(height: 20),
                  ElevatedButton(
                    onPressed: _isLoading ? null : _requestReset,
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(vertical: 16),
                    ),
                    child: _isLoading
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('ENVIAR INSTRUÇÕES'),
                  ),
                ] else ...[
                  TextField(
                    controller: _tokenController,
                    decoration: const InputDecoration(
                      labelText: 'Token recebido por e-mail',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 16),
                  TextField(
                    controller: _newPasswordController,
                    decoration: const InputDecoration(
                      labelText: 'Nova Senha (mín. 8 caracteres)',
                      border: OutlineInputBorder(),
                    ),
                    obscureText: true,
                  ),
                  const SizedBox(height: 20),
                  ElevatedButton(
                    onPressed: _isLoading ? null : _confirmReset,
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(vertical: 16),
                    ),
                    child: _isLoading
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('REDEFINIR SENHA'),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}
