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

// Estado simples de autenticação
class AuthState {
  final bool isLoading;
  final String? error;
  final bool isLoggedIn;

  const AuthState({
    this.isLoading = false,
    this.error,
    this.isLoggedIn = false,
  });

  AuthState copyWith({bool? isLoading, String? error, bool? isLoggedIn}) {
    return AuthState(
      isLoading: isLoading ?? this.isLoading,
      error: error,
      isLoggedIn: isLoggedIn ?? this.isLoggedIn,
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
      final response = await _dio.post('/auth/login', data: {
        'email': email,
        'password': password,
      });

      final accessToken = response.data['access_token'] as String?;
      final refreshToken = response.data['refresh_token'] as String?;

      if (accessToken != null && refreshToken != null) {
        await _storage.write(key: 'access_token', value: accessToken);
        await _storage.write(key: 'refresh_token', value: refreshToken);
      }

      state = state.copyWith(isLoading: false, isLoggedIn: true);
    } on DioException catch (e) {
      final message = (e.response?.data as Map?)?['error'] as String? ??
          'Erro de conexão';
      state = state.copyWith(isLoading: false, error: message);
    } catch (_) {
      state = state.copyWith(isLoading: false, error: 'Ocorreu um erro inesperado.');
    }
  }

  Future<void> logout() async {
    try {
      await _dio.post('/auth/logout');
    } catch (_) {}
    await _storage.delete(key: 'access_token');
    await _storage.delete(key: 'refresh_token');
    state = const AuthState();
  }
}

final authProvider = NotifierProvider<AuthNotifier, AuthState>(AuthNotifier.new);
