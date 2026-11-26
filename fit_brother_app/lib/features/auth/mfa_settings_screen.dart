import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'auth_provider.dart';

class MFASettingsScreen extends ConsumerStatefulWidget {
  const MFASettingsScreen({super.key});

  @override
  ConsumerState<MFASettingsScreen> createState() => _MFASettingsScreenState();
}

class _MFASettingsScreenState extends ConsumerState<MFASettingsScreen> {
  final _codeController = TextEditingController();
  final _passwordController = TextEditingController();

  bool _loadingStatus = true;
  bool _busy = false;
  bool _enabled = false;
  String? _secret;
  String? _otpauthUri;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadStatus();
  }

  @override
  void dispose() {
    _codeController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _loadStatus() async {
    setState(() {
      _loadingStatus = true;
      _error = null;
    });
    try {
      final res = await ref.read(dioClientProvider).dio.get<dynamic>('/auth/mfa/status');
      final data = res.data is Map ? Map<String, dynamic>.from(res.data as Map) : <String, dynamic>{};
      if (!mounted) return;
      setState(() {
        _enabled = data['enabled'] == true;
        _loadingStatus = false;
      });
    } on DioException catch (e) {
      if (!mounted) return;
      setState(() {
        _loadingStatus = false;
        _error = (e.response?.data as Map?)?['error']?.toString() ?? 'Falha ao consultar MFA';
      });
    }
  }

  Future<void> _setupMFA() async {
    setState(() => _busy = true);
    try {
      final res = await ref.read(dioClientProvider).dio.post<dynamic>('/auth/mfa/setup');
      final data = res.data is Map ? Map<String, dynamic>.from(res.data as Map) : <String, dynamic>{};
      if (!mounted) return;
      setState(() {
        _secret = data['secret'] as String?;
        _otpauthUri = data['otpauth_uri'] as String?;
      });
    } on DioException catch (e) {
      _showSnack((e.response?.data as Map?)?['error']?.toString() ?? 'Falha ao configurar MFA', true);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _enableMFA() async {
    final code = _codeController.text.trim();
    if (code.length != 6) {
      _showSnack('Informe o código de 6 dígitos do autenticador.', true);
      return;
    }
    setState(() => _busy = true);
    try {
      await ref.read(dioClientProvider).dio.post<dynamic>(
        '/auth/mfa/enable',
        data: {'code': code},
      );
      _codeController.clear();
      if (!mounted) return;
      setState(() {
        _enabled = true;
        _secret = null;
        _otpauthUri = null;
      });
      _showSnack('MFA TOTP ativado com sucesso.', false);
    } on DioException catch (e) {
      _showSnack((e.response?.data as Map?)?['error']?.toString() ?? 'Código TOTP inválido', true);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _disableMFA() async {
    final code = _codeController.text.trim();
    final pass = _passwordController.text;
    if (pass.isEmpty || code.length != 6) {
      _showSnack('Informe sua senha atual e o código TOTP de 6 dígitos.', true);
      return;
    }
    setState(() => _busy = true);
    try {
      await ref.read(dioClientProvider).dio.post<dynamic>(
        '/auth/mfa/disable',
        data: {'password': pass, 'code': code},
      );
      _codeController.clear();
      _passwordController.clear();
      if (!mounted) return;
      setState(() => _enabled = false);
      _showSnack('MFA TOTP desativado.', false);
    } on DioException catch (e) {
      _showSnack((e.response?.data as Map?)?['error']?.toString() ?? 'Falha ao desativar MFA', true);
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
      appBar: AppBar(title: const Text('Segurança da Conta (MFA TOTP)')),
      body: _loadingStatus
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(child: Text(_error!, style: const TextStyle(color: Colors.red)))
              : Center(
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.all(24),
                    child: Container(
                      constraints: const BoxConstraints(maxWidth: 480),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Card(
                            child: ListTile(
                              leading: Icon(
                                _enabled ? Icons.verified_user : Icons.shield_outlined,
                                color: _enabled ? Colors.green : Colors.orange,
                              ),
                              title: Text(_enabled ? 'MFA TOTP Ativo' : 'MFA TOTP Desativado'),
                              subtitle: const Text(
                                'Proteja o acesso de treinador com códigos temporários de 6 dígitos.',
                              ),
                            ),
                          ),
                          const SizedBox(height: 20),
                          if (!_enabled && _secret == null)
                            ElevatedButton.icon(
                              onPressed: _busy ? null : _setupMFA,
                              icon: const Icon(Icons.qr_code),
                              label: const Text('CONFIGURAR AUTENTICADOR'),
                            ),
                          if (!_enabled && _secret != null) ...[
                            const Text(
                              'Adicione esta chave no Google Authenticator ou 1Password:',
                              style: TextStyle(fontWeight: FontWeight.w600),
                            ),
                            const SizedBox(height: 8),
                            SelectableText(
                              _secret!,
                              style: const TextStyle(
                                fontFamily: 'monospace',
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            if (_otpauthUri != null) ...[
                              const SizedBox(height: 8),
                              SelectableText(
                                _otpauthUri!,
                                style: const TextStyle(fontSize: 12, color: Colors.black54),
                              ),
                            ],
                            const SizedBox(height: 16),
                            TextField(
                              controller: _codeController,
                              decoration: const InputDecoration(
                                labelText: 'Código de confirmação (6 dígitos)',
                                border: OutlineInputBorder(),
                              ),
                              keyboardType: TextInputType.number,
                              maxLength: 6,
                            ),
                            const SizedBox(height: 12),
                            ElevatedButton(
                              onPressed: _busy ? null : _enableMFA,
                              child: const Text('ATIVAR MFA'),
                            ),
                          ],
                          if (_enabled) ...[
                            TextField(
                              controller: _passwordController,
                              decoration: const InputDecoration(
                                labelText: 'Senha atual',
                                border: OutlineInputBorder(),
                              ),
                              obscureText: true,
                            ),
                            const SizedBox(height: 12),
                            TextField(
                              controller: _codeController,
                              decoration: const InputDecoration(
                                labelText: 'Código TOTP atual (6 dígitos)',
                                border: OutlineInputBorder(),
                              ),
                              keyboardType: TextInputType.number,
                              maxLength: 6,
                            ),
                            const SizedBox(height: 12),
                            OutlinedButton(
                              onPressed: _busy ? null : _disableMFA,
                              child: const Text('DESATIVAR MFA'),
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
