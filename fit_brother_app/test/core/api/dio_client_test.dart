import 'package:fit_brother_app/core/api/dio_client.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';

const configuredApiBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://localhost:8080/v1',
);

void main() {
  test('uses API_BASE_URL supplied at build time', () {
    final client = DioClient(secureStorage: const FlutterSecureStorage());

    expect(client.dio.options.baseUrl, configuredApiBaseUrl);
  });
}
