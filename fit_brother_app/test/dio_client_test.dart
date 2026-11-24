import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:fit_brother_app/core/api/dio_client.dart';
import 'package:fit_brother_app/core/api/token_storage.dart';

class InMemoryTokenStorage implements TokenStorage {
  final Map<String, String> values = {};

  @override
  Future<String?> read({required String key}) async => values[key];

  @override
  Future<void> write({required String key, required String? value}) async {
    if (value == null) {
      values.remove(key);
    } else {
      values[key] = value;
    }
  }

  @override
  Future<void> delete({required String key}) async {
    values.remove(key);
  }
}

typedef AdapterHandler = Future<ResponseBody> Function(RequestOptions options);

class FakeHttpClientAdapter implements HttpClientAdapter {
  final AdapterHandler handler;

  FakeHttpClientAdapter(this.handler);

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) {
    return handler(options);
  }

  @override
  void close({bool force = false}) {}
}

ResponseBody jsonResponse(int statusCode, Object? body) {
  return ResponseBody.fromString(
    jsonEncode(body),
    statusCode,
    headers: {
      Headers.contentTypeHeader: [Headers.jsonContentType],
    },
  );
}

void main() {
  test('Queues concurrent 401 responses into a single /auth/refresh and retries requests', () async {
    final storage = InMemoryTokenStorage()
      ..values['access_token'] = 'expired-access'
      ..values['refresh_token'] = 'refresh-1';

    var refreshCalls = 0;
    final retriedAuthHeaders = <String?>[];

    final adapter = FakeHttpClientAdapter((options) async {
      if (options.path == '/auth/refresh') {
        refreshCalls++;
        await Future<void>.delayed(const Duration(milliseconds: 25));
        return jsonResponse(200, {
          'access_token': 'fresh-access',
          'refresh_token': 'refresh-2',
          'expires_in': 900,
          'role': 'TRAINER',
        });
      }

      if (options.headers['Authorization'] == 'Bearer expired-access') {
        return jsonResponse(401, {'error': 'expired'});
      }

      retriedAuthHeaders.add(options.headers['Authorization'] as String?);
      return jsonResponse(200, {'ok': true, 'path': options.path});
    });

    final client = DioClient(
      tokenStorage: storage,
      useCookieSession: false,
      httpClientAdapter: adapter,
    );

    final responses = await Future.wait([
      client.dio.get<dynamic>('/workouts'),
      client.dio.get<dynamic>('/students'),
      client.dio.get<dynamic>('/workouts/abc'),
    ]);

    expect(refreshCalls, 1);
    expect(responses.every((r) => r.statusCode == 200), isTrue);
    expect(
      retriedAuthHeaders,
      equals(['Bearer fresh-access', 'Bearer fresh-access', 'Bearer fresh-access']),
    );
    expect(storage.values['access_token'], 'fresh-access');
    expect(storage.values['refresh_token'], 'refresh-2');
  });

  test('Clears tokens and triggers onSessionExpired when refresh fails', () async {
    final storage = InMemoryTokenStorage()
      ..values['access_token'] = 'expired-access'
      ..values['refresh_token'] = 'revoked-refresh';
    var expiredCallbacks = 0;

    final adapter = FakeHttpClientAdapter((options) async {
      if (options.path == '/auth/refresh') {
        return jsonResponse(401, {'error': 'revoked'});
      }
      return jsonResponse(401, {'error': 'unauthorized'});
    });

    final client = DioClient(
      tokenStorage: storage,
      useCookieSession: false,
      httpClientAdapter: adapter,
      onSessionExpired: () => expiredCallbacks++,
    );

    await expectLater(
      client.dio.get<dynamic>('/workouts'),
      throwsA(isA<DioException>()),
    );

    expect(expiredCallbacks, 1);
    expect(storage.values.containsKey('access_token'), isFalse);
    expect(storage.values.containsKey('refresh_token'), isFalse);
  });

  test('Web cookie mode enables withCredentials, attaches X-CSRF-Token on mutations, and avoids localStorage', () async {
    final storage = InMemoryTokenStorage();
    final capturedHeaders = <Map<String, dynamic>>[];
    final capturedWithCreds = <Object?>[];

    final adapter = FakeHttpClientAdapter((options) async {
      capturedHeaders.add(Map<String, dynamic>.from(options.headers));
      capturedWithCreds.add(options.extra['withCredentials']);
      if (options.path == '/auth/refresh') {
        return jsonResponse(200, {
          'expires_in': 900,
          'role': 'TRAINER',
          'session_mode': 'cookie',
        });
      }
      if (options.extra['_retried'] != true && options.method == 'POST') {
        return jsonResponse(401, {'error': 'expired'});
      }
      return jsonResponse(200, {'ok': true});
    });

    final client = DioClient(
      tokenStorage: storage,
      useCookieSession: true,
      csrfCookieReader: (name) => name == DioClient.csrfCookieName ? 'csrf-web-token' : null,
      httpClientAdapter: adapter,
    );

    await client.dio.get<dynamic>('/students');
    await client.dio.post<dynamic>('/workouts', data: {'title': 'Treino A'});

    expect(storage.values, isEmpty);
    expect(capturedWithCreds.every((v) => v == true), isTrue);
    expect(capturedHeaders[0].containsKey('Authorization'), isFalse);
    expect(capturedHeaders[0].containsKey(DioClient.csrfHeaderName), isFalse);
    // POST -> 401 -> /auth/refresh (with X-Session-Mode: cookie & X-CSRF-Token) -> retried POST (with X-CSRF-Token)
    expect(capturedHeaders[1][DioClient.csrfHeaderName], 'csrf-web-token');
    expect(capturedHeaders[2][DioClient.sessionModeHeaderName], 'cookie');
    expect(capturedHeaders[2][DioClient.csrfHeaderName], 'csrf-web-token');
    expect(capturedHeaders[3][DioClient.csrfHeaderName], 'csrf-web-token');
  });
}
