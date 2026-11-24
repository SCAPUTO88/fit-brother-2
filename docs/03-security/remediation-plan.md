# Plano de Remediação de Segurança

Plano ordenado por prioridade de risco (abordagem RED/GREEN com teste de regressão obrigatório por item).

---

## Onda 1 — Fase 1 (Security Baseline — Bloqueadores de Qualquer Uso Real)

1. **[SEC-01/SEC-02/SEC-05] Isolamento e Auth de Alunos:**
   - Proteger `POST /v1/students` com `RequireAuth` + role `TRAINER` (e `trainer_id` da sessão).
   - Migration `006`: incluir `trainer_id` em `students` com índice composto e preparar transição para `trainer_student_relationships`.
   - Filtrar `GET /v1/students` somente para `TRAINER` e apenas seus alunos; impedir `STUDENT` (`403`).
   - Validar vínculo do trainer em `GET /v1/students/:id` e `DELETE /v1/students/:id`.
   - Remover exposição de CPF em listagens e começar plano de depreciação.
2. **[SEC-03/SEC-07] Correção Crítica e BOLA em Workouts:**
   - Resolver `students.id` a partir de `claims.UserID` no `WorkoutHandler` / use cases (corrigindo erro que impede aluno de ver seus treinos).
   - Validar posse do plano pai em `CreatePlan`, `CreateSession`, `CreateExercise`.
   - Mapear corretamente `"acesso negado"` → `403` e `"não encontrado"` → `404` sem `err.Error()` cru.
3. **[SEC-04/SEC-08] Segurança no Fluxo de Vídeos:**
   - Validar vínculo ativo/posse do exercício em `RequestUploadURL`, `ConfirmUpload` e `GetSignedURL`.
   - Gerar `r2_key` controlada no backend sem aceitar path arbitrário; exigir MIME permitido e tamanho máximo.
   - Persistir intenção de upload (`PENDING`) e exigir `HeadObject` no R2 antes de confirmar (`CONFIRMED`).
   - Alinhar Flutter `create_exercise_screen.dart` com o contrato real do backend.
4. **[SEC-06/SEC-09/SEC-10/SEC-13] Auth, Transações e Higiene:**
   - Colocar `CreateStudent`, `Anonymize` (anonimizando `users.email`/`full_name`, setando `deleted_at`, revogando tokens) e `Refresh` dentro de transações SQL explícitas.
   - Checar `is_active` / revogação de usuário no refresh e no fluxo crítico.
   - Corrigir `auth_provider.dart` e `logout()` do Flutter enviando `refresh_token` e protegendo rotas no `GoRouter`.
   - Remover `scripts/debug_crypto.go` (ou eliminar print de `PASSWORD_PEPPER`) e exigir senha via argumento/prompt seguro em `seed_trainer.go`.
   - Implementar testes reais de integração em `tests/integration/` contra PostgreSQL.

---

## Onda 2 — Fase 2 (Arquitetura Segura Web & Pipeline DevSecOps)

1. Implementar interceptor de refresh automático com fila no `DioClient` e migrar sessão Web para cookies `HttpOnly; Secure; SameSite` via mesmo domínio/BFF (ADR-008) + proteção CSRF.
2. Adicionar rate limiter por IP/usuário em login, refresh e convites; adicionar cabeçalhos HTTP de segurança (`X-Content-Type-Options`, `X-Frame-Options`/CSP, `Referrer-Policy`).
3. Adicionar `kid` no JWT RS256 e estratégia de rotação sem derrubar sessões ativas abruptamente.
4. Endurecer CI/CD com versões fixas, `gitleaks`, `semgrep`, `govulncheck`, `osv-scanner`, `trivy`, SBOM e gates Flutter.

---

## Onda 3 — Fases 3, 9 e 10 (Identidade, Auditoria e DAST)

1. Convites com hash único e expiração; reset de senha com invalidação de sessões anteriores; MFA TOTP para treinadores.
2. Registro de `audit_events`, `consent_records` e `privacy_requests`.
3. DAST (OWASP ZAP) em ambiente de homologação autorizado antes do Go/No-Go comercial.
