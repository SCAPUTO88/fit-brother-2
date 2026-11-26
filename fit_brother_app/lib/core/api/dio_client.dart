import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import 'token_storage.dart';
import 'web_cookie_stub.dart' if (dart.library.js_interop) 'web_cookie_web.dart';

const defaultApiBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://localhost:8080/v1',
);

typedef CsrfCookieReader = String? Function(String cookieName);

class DioClient {
  static const csrfCookieName = 'fitbrother_csrf';
  static const csrfHeaderName = 'X-CSRF-Token';
  static const sessionModeHeaderName = 'X-Session-Mode';

  final Dio dio;
  final Dio _refreshDio;
  final TokenStorage tokenStorage;
  final bool useCookieSession;
  final CsrfCookieReader _readCookie;
  final FutureOr<void> Function()? onSessionExpired;

  Completer<bool>? _refreshCompleter;

  DioClient({
    TokenStorage? tokenStorage,
    FlutterSecureStorage? secureStorage,
    String baseUrl = defaultApiBaseUrl,
    this.useCookieSession = kIsWeb,
    CsrfCookieReader? csrfCookieReader,
    HttpClientAdapter? httpClientAdapter,
    this.onSessionExpired,
  }) : tokenStorage =
           tokenStorage ??
           (useCookieSession
               ? const NoOpTokenStorage()
               : SecureTokenStorage(
                   secureStorage ?? const FlutterSecureStorage(),
                 )),
       _readCookie = csrfCookieReader ?? readBrowserCookie,
       dio = Dio(
         BaseOptions(
           baseUrl: baseUrl,
           connectTimeout: const Duration(seconds: 10),
           receiveTimeout: const Duration(seconds: 10),
           headers: {'Content-Type': 'application/json'},
           extra: useCookieSession ? {'withCredentials': true} : const {},
         ),
       ),
       _refreshDio = Dio(
         BaseOptions(
           baseUrl: baseUrl,
           connectTimeout: const Duration(seconds: 10),
           receiveTimeout: const Duration(seconds: 10),
           headers: {'Content-Type': 'application/json'},
           extra: useCookieSession ? {'withCredentials': true} : const {},
         ),
       ) {
    if (httpClientAdapter != null) {
      dio.httpClientAdapter = httpClientAdapter;
      _refreshDio.httpClientAdapter = httpClientAdapter;
    }

    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          if (useCookieSession) {
            options.extra['withCredentials'] = true;
            _attachCsrfHeaderIfMutating(options);
          } else {
            final token = await this.tokenStorage.read(key: 'access_token');
            if (token != null &&
                token.isNotEmpty &&
                options.headers['Authorization'] == null) {
              options.headers['Authorization'] = 'Bearer $token';
            }
          }
          return handler.next(options);
        },
        onError: (DioException e, handler) async {
          if (e.response?.statusCode != 401) {
            return handler.next(e);
          }

          final requestOptions = e.requestOptions;
          if (_isAuthLifecyclePath(requestOptions.path) ||
              requestOptions.extra['_retried'] == true) {
            if (_isRefreshPath(requestOptions.path)) {
              await _notifySessionExpired();
            }
            return handler.next(e);
          }

          final refreshed = await _refreshAccessToken();
          if (!refreshed) {
            return handler.next(e);
          }

          requestOptions.extra['_retried'] = true;
          if (useCookieSession) {
            requestOptions.extra['withCredentials'] = true;
            _attachCsrfHeaderIfMutating(requestOptions);
          } else {
            final newToken = await this.tokenStorage.read(key: 'access_token');
            if (newToken == null || newToken.isEmpty) {
              return handler.next(e);
            }
            requestOptions.headers['Authorization'] = 'Bearer $newToken';
          }

          try {
            final retryResponse = await dio.fetch<dynamic>(requestOptions);
            return handler.resolve(retryResponse);
          } on DioException catch (retryError) {
            return handler.next(retryError);
          }
        },
      ),
    );
  }

  void _attachCsrfHeaderIfMutating(RequestOptions options) {
    final method = options.method.toUpperCase();
    if (method == 'POST' ||
        method == 'PUT' ||
        method == 'PATCH' ||
        method == 'DELETE') {
      final csrfToken = _readCookie(csrfCookieName);
      if (csrfToken != null && csrfToken.isNotEmpty) {
        options.headers[csrfHeaderName] = csrfToken;
      }
    }
  }

  Future<bool> _refreshAccessToken() async {
    final inFlight = _refreshCompleter;
    if (inFlight != null) {
      return inFlight.future;
    }

    final completer = Completer<bool>();
    _refreshCompleter = completer;

    var success = false;
    try {
      success = await _performRefresh();
      if (!completer.isCompleted) {
        completer.complete(success);
      }
      return success;
    } catch (_) {
      if (!completer.isCompleted) {
        completer.complete(false);
      }
      return false;
    } finally {
      _refreshCompleter = null;
    }
  }

  Future<bool> _performRefresh() async {
    try {
      if (useCookieSession) {
        final csrf = _readCookie(csrfCookieName);
        final response = await _refreshDio.post<dynamic>(
          '/auth/refresh',
          options: Options(
            headers: {
              sessionModeHeaderName: 'cookie',
              if (csrf != null && csrf.isNotEmpty) csrfHeaderName: csrf,
            },
            extra: const {'withCredentials': true},
          ),
        );
        final code = response.statusCode ?? 0;
        if (code >= 200 && code < 300) {
          return true;
        }
        await _notifySessionExpired();
        return false;
      }

      final refreshToken = await tokenStorage.read(key: 'refresh_token');
      if (refreshToken == null || refreshToken.isEmpty) {
        await _clearStoredTokens();
        await _notifySessionExpired();
        return false;
      }

      final response = await _refreshDio.post<dynamic>(
        '/auth/refresh',
        data: {'refresh_token': refreshToken},
      );
      final data = response.data is Map
          ? Map<String, dynamic>.from(response.data as Map)
          : <String, dynamic>{};
      final newAccess = data['access_token'] as String?;
      final newRefresh = data['refresh_token'] as String?;

      if (newAccess == null ||
          newAccess.isEmpty ||
          newRefresh == null ||
          newRefresh.isEmpty) {
        await _clearStoredTokens();
        await _notifySessionExpired();
        return false;
      }

      await tokenStorage.write(key: 'access_token', value: newAccess);
      await tokenStorage.write(key: 'refresh_token', value: newRefresh);
      return true;
    } catch (_) {
      if (!useCookieSession) {
        await _clearStoredTokens();
      }
      await _notifySessionExpired();
      return false;
    }
  }

  Future<void> _clearStoredTokens() async {
    await tokenStorage.delete(key: 'access_token');
    await tokenStorage.delete(key: 'refresh_token');
    await tokenStorage.delete(key: 'user_role');
  }

  Future<void> _notifySessionExpired() async {
    if (onSessionExpired != null) {
      await onSessionExpired!();
    }
  }

  bool _isRefreshPath(String path) => path.endsWith('/auth/refresh');

  bool _isAuthLifecyclePath(String path) =>
      path.endsWith('/auth/login') ||
      path.endsWith('/auth/refresh') ||
      path.endsWith('/auth/logout') ||
      path.endsWith('/auth/invitations/accept') ||
      path.endsWith('/auth/password-reset/request') ||
      path.endsWith('/auth/password-reset/confirm');
}
