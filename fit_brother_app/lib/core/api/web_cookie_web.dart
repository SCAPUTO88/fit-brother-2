import 'dart:js_interop';

@JS('document')
external _JsDocument get _document;

extension type _JsDocument(JSObject _) implements JSObject {
  external String get cookie;
}

String? readBrowserCookie(String name) {
  final raw = _document.cookie;
  if (raw.isEmpty) return null;

  for (final part in raw.split(';')) {
    final trimmed = part.trim();
    final sepIndex = trimmed.indexOf('=');
    if (sepIndex <= 0) continue;
    final cookieName = trimmed.substring(0, sepIndex).trim();
    if (cookieName == name) {
      return Uri.decodeComponent(trimmed.substring(sepIndex + 1).trim());
    }
  }
  return null;
}
