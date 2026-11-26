# Estado Atual do Projeto (Checkpoint de Handoff)

**Última atualização:** 2026-11-26  
**Fase atual:** **Fase 3 — Identidade, Multi-Tenancy & LGPD (Implementada, validada no GitHub Actions e aberta em PR `#3` para `develop`)**  
**Repositório conectado:** `https://github.com/SCAPUTO88/fit-brother-2`  
**Branch ativa:** `feat/fase-3-identidade-multitenancy-lgpd` → **PR `#3`** (`https://github.com/SCAPUTO88/fit-brother-2/pull/3`) para `develop`  
**Branch `origin/main`:** `565e121` (Sprint 2 — aguardando promoção controlada após validação em `develop`)

---

## 1. Resumo das Fases Concluídas

- **Fase 1 — Security Baseline:** Merged em `develop` via PR `#1` (isolamento trainer/aluno com `migrations/006`, transações atômicas em Auth/Student/Anonymize, `HeadObject` em vídeos, `SecurityHeaders`, suíte de integração PostgreSQL e correções iniciais no Flutter).
- **Fase 2 — Simplificação & Modernização (`BL-07`, `BL-08`, `BL-09`):** Merged em `develop` via PR `#2` (`StoragePort` + `LocalStorage` sem Redis/R2 obrigatórios, `RateLimiter` em memória, sessão Web com cookies `HttpOnly` + Double-Submit CSRF, fila de refresh no Flutter e esteira DevSecOps completa).
- **Fase 3 — Identidade, Multi-Tenancy & LGPD (`BL-10`, `BL-11`, `BL-12`, `ADR-006`):**
  - **BL-10 — Organizações, Membros & Vínculos (`migrations/007`):**
    - Tabelas `organizations`, `memberships` (`OWNER`, `ADMIN`, `TRAINER`) e `trainer_student_relationships` (`ACTIVE`, `PAUSED`, `ENDED`) com backfill idempotente de organizações `SOLO` e vínculos para dados legados.
    - Endpoints `/v1/organizations` (`POST`, `GET`, `/:id/members`, `/:id/relationships/:student_id`) e autorização baseada em vínculo ativo em `students` e `workouts`.
  - **BL-11 — Convites, Redefinição de Senha via SMTP & MFA TOTP:**
    - Fluxo de convite com token aleatório armazenado via hash SHA-256 + `pepper` (`/v1/invitations` e `POST /v1/auth/invitations/accept`), revogação automática de convites pendentes no reenvio e suporte a modo Web `HttpOnly` cookie + CSRF.
    - `MailerPort` + `SMTPMailer` (STARTTLS/TLS >= 1.2 + sanitização anti-injeção de cabeçalhos CRLF) e `NoopMailer` em dev/test.
    - Redefinição de senha anti-enumeração (`POST /v1/auth/password-reset/request` e `/confirm`) com invalidação atômica de tokens e revogação de todos os `refresh_tokens`.
    - MFA TOTP RFC 6238 (`internal/infrastructure/totp`) com proteção anti-replay (`last_used_step` com lock `FOR UPDATE`), `POST /v1/auth/mfa/{setup,enable,disable}`, `GET /v1/auth/mfa/status` e retorno `401 mfa_required: true` no login.
  - **BL-12 & ADR-006 — CPF Opcional/Deprecado & Central LGPD:**
    - `students.cpf_encrypted` e `students.cpf_hash` alterados para opcionais (`NULL`); criação de alunos via convite ou cadastro direto sem exigir CPF no backend e no Flutter (`CreateStudentScreen`).
    - Tabelas `consent_records`, `privacy_requests` e `audit_events` com endpoints `/v1/privacy/{consents,requests,export,anonymize}` e `GET /v1/audit-events` (restrito a treinadores/estúdios).
    - Flutter (`fit_brother_app`): rotas públicas `/activate` (aceite de convite + consentimentos LGPD) e `/forgot-password`, desafio TOTP no login, telas `MfaSettingsScreen` (`/mfa`) e `PrivacyCenterScreen` (`/privacy`), e `DashboardScreen` sensível ao papel (`TRAINER` vs `STUDENT`).

---

## 2. Estado do Pipeline e Próximos Passos

1. **Esteira CI (`.github/workflows/ci.yml`):**
   - Pipeline validada na run [`45735271644`](https://github.com/SCAPUTO88/fit-brother-2/actions/runs/45735271644) (`e41c7de`, detalhes em [`../06-quality/test-results.md`](../06-quality/test-results.md)).
2. **Revisão e merge:**
   - PR [`#3`](https://github.com/SCAPUTO88/fit-brother-2/pull/3) (`feat/fase-3-identidade-multitenancy-lgpd` → `develop`) pronto para revisão e aprovação de merge pelo usuário antes de avançar para a **Fase 4 — Catálogo de Exercícios**.

