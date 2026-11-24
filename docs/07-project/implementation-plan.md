# Plano de Implementação Detalhado (Próximas Fases)

Este plano orienta a execução sequencial após aprovação da Fase 0. **Nenhuma tarefa abaixo foi iniciada.**

---

## 1. Fase 1 — Security Baseline (Detalhamento Executável)

1. **Entrega 1A — DB & Isolamento de Alunos:**
   - Criar `migrations/006_add_trainer_ownership.up.sql` (+ `.down.sql`).
   - Proteger `POST /v1/students` em `student_handler.go` exigindo `TRAINER`; restringir `GET /v1/students` a `TRAINER` filtrando por `trainer_id`; validar vínculo em `GET/DELETE /students/:id`.
   - Envolver `CreateStudentUseCase` e `AnonymizeStudentUseCase` em transação SQL; anonimizar `users.email`/`full_name`, setar `deleted_at` e revogar `refresh_tokens`.
2. **Entrega 1B — Correção Crítica de Workouts & Auth Middleware:**
   - Resolver `students.id` via `user_id` ao listar/consultar planos como `STUDENT`.
   - Validar posse do plano em `CreatePlan`, `CreateSession` e `CreateExercise`; mapear `403` e `404` corretamente sem `err.Error()`.
   - Envolver `RefreshUseCase` em transação e validar `is_active` na rotação e nas operações sensíveis.
3. **Entrega 1C — Hardening de Vídeos & Higiene:**
   - Validar posse do exercício em upload/confirm/playback de vídeo; gerar `r2_key` controlada no servidor; adicionar tabela/estado `PENDING` + `HeadObject` antes de confirmar.
   - Remover/sanitizar print de `PASSWORD_PEPPER` em `scripts/debug_crypto.go`; corrigir `auth_provider.dart` e `logout()` no Flutter; implementar testes reais em `tests/integration/`.

## 2. Fase 2 — Simplificação e Modernização
- Tornar `REDIS_URL` e `R2_*` opcionais em `config.go`; extrair `domain.StoragePort` + `LocalStorageAdapter`; limpar menções GCP/Railway; implementar refresh queue no `DioClient` e arquitetura de sessão Web (`ADR-008`); atualizar CI (`.github/workflows/ci.yml`) com Flutter e ferramentas DevSecOps com versões fixas.

## 3. Fases 3 a 10
- Seguem exatamente a sequência do [`roadmap.md`](roadmap.md) e [`../01-product/backlog.md`](../01-product/backlog.md), sempre mediante validação de critérios de aceitação, atualização de `current-state.md` e revisão antes de qualquer merge ou deploy.
