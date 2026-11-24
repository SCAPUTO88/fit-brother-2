# ADR-008 — Segurança de Sessão Web (Cookies HttpOnly sob Mesmo Domínio vs `flutter_secure_storage`)

**Status:** Proposto (Aguardando aprovação na Fase 0)  
**Data:** 2026-11-23

## Contexto
Atualmente o cliente Flutter usa `flutter_secure_storage: ^10.0.0` (`dio_client.dart`, `auth_provider.dart`). No ambiente Web, essa biblioteca não utiliza Keychain/Keystore de hardware, recorrendo a WebCrypto com chaves/dados persistidos no `localStorage` do navegador, ficando exposta caso ocorra XSS em scripts ou dependências.

## Decisão
Para o **Flutter Web / PWA em produção**:
1. Servir o frontend (`app.seudominio`) e fazer proxy da API (`/v1/*`) sob a **mesma origem/domínio** via Cloudflare Workers/Routes ou proxy reverso (eliminando problemas de cookies third-party / Safari ITP).
2. A API Go passará a suportar modo de sessão Web emitindo **`access_token` e `refresh_token` em cookies `HttpOnly; Secure; SameSite=Lax`**, com Path restrito (`/` ou `/v1/auth/refresh` para o refresh cookie).
3. Implementar **proteção contra CSRF** em todas as requisições mutáveis (`POST`, `PUT`, `PATCH`, `DELETE`) quando autenticadas via cookie: validação estrita de `Origin`/`Referer` + token CSRF (Double-Submit Cookie ou header customizado validado pelo backend) + CORS estrito sem wildcard (`*`) quando `AllowCredentials=true`.
4. No Flutter Web (`kIsWeb`), o `DioClient` enviará `extra['withCredentials'] = true` e **não armazenará tokens no `localStorage`/`flutter_secure_storage`**. Para futuros builds móveis nativos (`!kIsWeb`), mantém-se `Bearer` via Keychain/Keystore nativo.
5. No **Service Worker** da PWA (Fase 8), proibir estritamente o cache de `/v1/*` e de URLs assinadas de vídeo, garantindo limpeza total de estado em memória no `logout`.

## Consequências
- Elimina exfiltração direta de tokens via XSS no `localStorage`, resolve restrições cross-site em iOS Safari PWA e mantém compatibilidade futura com apps móveis usando o mesmo backend Go.
