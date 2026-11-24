import 'package:flutter_secure_storage/flutter_secure_storage.dart';

abstract class TokenStorage {
  Future<String?> read({required String key});
  Future<void> write({required String key, required String? value});
  Future<void> delete({required String key});
}

class SecureTokenStorage implements TokenStorage {
  final FlutterSecureStorage _secureStorage;

  const SecureTokenStorage(this._secureStorage);

  @override
  Future<String?> read({required String key}) => _secureStorage.read(key: key);

  @override
  Future<void> write({required String key, required String? value}) =>
      _secureStorage.write(key: key, value: value);

  @override
  Future<void> delete({required String key}) => _secureStorage.delete(key: key);
}

/// No-op storage para Flutter Web em modo HttpOnly cookie (ADR-008),
/// impedindo qualquer persistência de JWT/refresh tokens em localStorage.
class NoOpTokenStorage implements TokenStorage {
  const NoOpTokenStorage();

  @override
  Future<String?> read({required String key}) async => null;

  @override
  Future<void> write({required String key, required String? value}) async {}

  @override
  Future<void> delete({required String key}) async {}
}
