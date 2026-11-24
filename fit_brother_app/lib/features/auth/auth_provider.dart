import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../../core/api/dio_client.dart';

final secureStorageProvider = Provider<FlutterSecureStorage>((ref) {
  return const FlutterSecureStorage();
});

final dioClientProvider = Provider<DioClient>((ref) {
  final storage = ref.watch(secureStorageProvider);
  return DioClient(secureStorage: storage);
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

  Dio get _dio => ref.read(dioClientProvider).dio;
  FlutterSecureStorage get _storage => ref.read(secureStorageProvider);

  Future<void> login(String email, String password) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final response = await _dio.post(
        '/auth/login',
        data: {'email': email, 'password': password},
      );

      final data = response.data is Map<String, dynamic>
          ? response.data as Map<String, dynamic>
          : <String, dynamic>{};
      final accessToken = data['access_token'] as String?;
      final refreshToken = data['refresh_token'] as String?;
      final role = data['role'] as String?;

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
      final refreshToken = await _storage.read(key: 'refresh_token');
      if (refreshToken != null && refreshToken.isNotEmpty) {
        await _dio.post('/auth/logout', data: {'refresh_token': refreshToken});
      }
    } catch (_) {}
    await _storage.delete(key: 'access_token');
    await _storage.delete(key: 'refresh_token');
    await _storage.delete(key: 'user_role');
    state = const AuthState();
  }
}

final authProvider = NotifierProvider<AuthNotifier, AuthState>(AuthNotifier.new);
