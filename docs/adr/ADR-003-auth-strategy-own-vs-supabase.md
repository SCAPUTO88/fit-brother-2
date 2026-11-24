# ADR-003 — Estratégia de Autenticação: Auth Própria (Go) vs Supabase Auth

**Status:** Proposto (Aguardando aprovação na Fase 0)  
**Data:** 2026-11-23

## Contexto
O Fit Brother já implementa autenticação JWT RS256, refresh tokens e hashing `HMAC-SHA256(PASSWORD_PEPPER, senha) -> bcrypt` no backend Go (`internal/application/auth`, `internal/infrastructure/crypto`). Com a adoção do Supabase PostgreSQL, avaliou-se manter a autenticação própria no Go ou migrar para o Supabase Auth.

## Comparação Objetiva

| Critério | Auth Própria no Backend Go (Evoluída) | Supabase Auth (`auth.users` + JWKS Go) |
| :--- | :--- | :--- |
| **Portabilidade / Lock-in** | 100% portável (qualquer PostgreSQL). | Acopla identidade, convites, e-mails e tokens ao Supabase Auth. |
| **Limites Free-Tier** | Sem limite artificial de MAU; limitado apenas por DB/API. | 50.000 MAUs grátis; limites rígidos de envio de e-mail interno do Supabase sem SMTP customizado (2–4 e-mails/hora no SMTP padrão Supabase!). |
| **Compatibilidade com hashes atuais** | **100% compatível.** | **Incompatível diretamente** com `PASSWORD_PEPPER` (Supabase valida senha pura contra hash bcrypt padrão, não pré-hash HMAC+pepper; exigiria re-hash lazy ou reset de senha). |
| **Sessão Web `HttpOnly` (ADR-008)** | Direto e nativo na API Go (emite cookies `HttpOnly; SameSite` diretamente). | SDK client-side do Supabase armazena tokens em `localStorage`/cookies JS por padrão; exigiria BFF extra para `HttpOnly` real. |
| **Esforço de Evolução (Convite/Reset/MFA)** | Requer implementar convites, reset via Resend/Brevo e TOTP no Go (`pquerna/otp`). | Oferece fluxos prontos, mas exige integração JWKS (`MicahParks/keyfunc`) + sincronização `auth.users` ↔ `public.users`. |

## Decisão
**Manter e evoluir a Autenticação Própria no Backend Go** (corrigindo transações, revogação, checagem de `is_active`, `kid` no JWT, cookies `HttpOnly` na Web, convites via SMTP próprio e MFA TOTP opcional na Fase 3).

## Plano de Contingência
Caso no futuro se decida adotar um IdP externo ou Supabase Auth, a interface de middleware JWT já isola claims e permite validar JWKS assimétrico sem alterar casos de uso de negócio.
