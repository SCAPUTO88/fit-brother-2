import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../../core/api/dio_client.dart';

final secureStorageProvider = Provider((ref) => const FlutterSecureStorage());

final dioClientProvider = Provider((ref) {
  final storage = ref.watch(secureStorageProvider);
  return DioClient(secureStorage: storage);
});

final authProvider = StateNotifierProvider<AuthNotifier, AsyncValue<void>>((ref) {
  final dioClient = ref.watch(dioClientProvider);
  final storage = ref.watch(secureStorageProvider);
  return AuthNotifier(dioClient.dio, storage);
});

class AuthNotifier extends StateNotifier<AsyncValue<void>> {
  final Dio _dio;
  final FlutterSecureStorage _storage;

  AuthNotifier(this._dio, this._storage) : super(const AsyncData(null));

  Future<void> login(String email, String password) async {
    state = const AsyncLoading();
    try {
      final response = await _dio.post('/auth/login', data: {
        'email': email,
        'password': password,
      });

      final accessToken = response.data['access_token'];
      final refreshToken = response.data['refresh_token'];

      if (accessToken != null && refreshToken != null) {
        await _storage.write(key: 'access_token', value: accessToken);
        await _storage.write(key: 'refresh_token', value: refreshToken);
      }
      
      state = const AsyncData(null);
    } on DioException catch (e) {
      final message = e.response?.data['error'] ?? 'Erro de conexão';
      state = AsyncError(message, e.stackTrace ?? StackTrace.current);
    } catch (e, st) {
      state = AsyncError('Ocorreu um erro inesperado.', st);
    }
  }

  Future<void> logout() async {
    try {
      await _dio.post('/auth/logout');
    } catch (_) {
      // Ignora erro no logout da API, mas limpa localmente
    } finally {
      await _storage.delete(key: 'access_token');
      await _storage.delete(key: 'refresh_token');
      // Força um recarregamento se necessário
    }
  }
}
