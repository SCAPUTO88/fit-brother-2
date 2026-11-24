import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../../core/api/dio_client.dart';
import '../../core/api/token_storage.dart';

final secureStorageProvider = Provider<FlutterSecureStorage>((ref) {
  return const FlutterSecureStorage();
});

final tokenStorageProvider = Provider<TokenStorage>((ref) {
  if (kIsWeb) {
    return const NoOpTokenStorage();
  }
  return SecureTokenStorage(ref.watch(secureStorageProvider));
});

final dioClientProvider = Provider<DioClient>((ref) {
  final storage = ref.watch(tokenStorageProvider);
  return DioClient(
    tokenStorage: storage,
    useCookieSession: kIsWeb,
    onSessionExpired: () {
      ref.read(authProvider.notifier).expireSession();
    },
  );
});

class AuthState {
  final bool isLoading;
  final String? error;
  final bool isLoggedIn;
  final String? role;

  const AuthState({
    this.isLoading = false,
    this.error,
    this.isLoggedIn = false,
    this.role,
  });

  AuthState copyWith({
    bool? isLoading,
    String? error,
    bool? isLoggedIn,
    String? role,
  }) {
    return AuthState(
      isLoading: isLoading ?? this.isLoading,
      error: error,
      isLoggedIn: isLoggedIn ?? this.isLoggedIn,
      role: role ?? this.role,
    );
  }
}

class AuthNotifier extends Notifier<AuthState> {
  @override
  AuthState build() => const AuthState();

  DioClient get _client => ref.read(dioClientProvider);
  Dio get _dio => _client.dio;
  TokenStorage get _storage => ref.read(tokenStorageProvider);

  Future<void> login(String email, String password) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final useCookie = _client.useCookieSession;
      final response = await _dio.post<dynamic>(
        '/auth/login',
        data: {'email': email, 'password': password},
        options: useCookie
            ? Options(headers: {DioClient.sessionModeHeaderName: 'cookie'})
            : null,
      );

      final data = response.data is Map
          ? Map<String, dynamic>.from(response.data as Map)
          : <String, dynamic>{};
      final role = data['role'] as String?;
      final sessionMode = data['session_mode'] as String?;

      if (useCookie) {
        if (sessionMode != 'cookie' &&
            (role == null || role.isEmpty)) {
          state = const AuthState(
            isLoading: false,
            isLoggedIn: false,
            error: 'Resposta de autenticação inválida do servidor.',
          );
          return;
        }

        state = AuthState(
          isLoading: false,
          isLoggedIn: true,
          role: role,
        );
        return;
      }

      final accessToken = data['access_token'] as String?;
      final refreshToken = data['refresh_token'] as String?;

      if (accessToken == null ||
          accessToken.isEmpty ||
          refreshToken == null ||
          refreshToken.isEmpty) {
        state = const AuthState(
          isLoading: false,
          isLoggedIn: false,
          error: 'Resposta de autenticação inválida do servidor.',
        );
        return;
      }

      await _storage.write(key: 'access_token', value: accessToken);
      await _storage.write(key: 'refresh_token', value: refreshToken);
      if (role != null && role.isNotEmpty) {
        await _storage.write(key: 'user_role', value: role);
      }

      state = AuthState(
        isLoading: false,
        isLoggedIn: true,
        role: role,
      );
    } on DioException catch (e) {
      final message = (e.response?.data as Map?)?['error'] as String? ??
          'Erro de conexão';
      state = state.copyWith(isLoading: false, isLoggedIn: false, error: message);
    } catch (_) {
      state = state.copyWith(
        isLoading: false,
        isLoggedIn: false,
        error: 'Ocorreu um erro inesperado.',
      );
    }
  }

  Future<void> logout() async {
    try {
      if (_client.useCookieSession) {
        await _dio.post<dynamic>(
          '/auth/logout',
          options: Options(headers: {DioClient.sessionModeHeaderName: 'cookie'}),
        );
      } else {
        final refreshToken = await _storage.read(key: 'refresh_token');
        if (refreshToken != null && refreshToken.isNotEmpty) {
          await _dio.post<dynamic>(
            '/auth/logout',
            data: {'refresh_token': refreshToken},
          );
        }
      }
    } catch (_) {}

    await _storage.delete(key: 'access_token');
    await _storage.delete(key: 'refresh_token');
    await _storage.delete(key: 'user_role');
    state = const AuthState();
  }

  void expireSession() {
    state = const AuthState();
  }
}

final authProvider = NotifierProvider<AuthNotifier, AuthState>(AuthNotifier.new);
