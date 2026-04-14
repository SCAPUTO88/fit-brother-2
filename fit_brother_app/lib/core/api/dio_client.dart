import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'dart:io';

class DioClient {
  final Dio dio;
  final FlutterSecureStorage secureStorage;

  DioClient({
    required this.secureStorage,
    String baseUrl = 'http://localhost:8080/v1',
  }) : dio = Dio(
          BaseOptions(
            baseUrl: baseUrl,
            connectTimeout: const Duration(seconds: 10),
            receiveTimeout: const Duration(seconds: 10),
            headers: {'Content-Type': 'application/json'},
          ),
        ) {
    dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await secureStorage.read(key: 'access_token');
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        return handler.next(options);
      },
      onError: (DioException e, handler) async {
        // Implement Refresh Token logic here later
        if (e.response?.statusCode == 401) {
          // Token expirado, forçar logout ou tentar refresh
        }
        return handler.next(e);
      },
    ));
  }
}
