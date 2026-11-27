# Estado Atual do Projeto (Checkpoint de Handoff)

**Última atualização:** 2026-11-27  
**Fase atual:** **Fase 5 — Workout Builder (Implementada, validada no GitHub Actions e aberta em PR `#5` para `develop`)**  
**Repositório conectado:** `https://github.com/SCAPUTO88/fit-brother-2`  
**Branch ativa:** `feat/fase-5-workout-builder` → **PR `#5`** (`https://github.com/SCAPUTO88/fit-brother-2/pull/5`) para `develop`  
**Branch `origin/main`:** `565e121` (Sprint 2 — aguardando promoção controlada após validação em `develop`)

---

## 1. Resumo das Fases Concluídas

- **Fase 1 — Security Baseline:** Merged em `develop` via PR `#1` (isolamento trainer/aluno com `migrations/006`, transações atômicas em Auth/Student/Anonymize, `HeadObject` em vídeos, `SecurityHeaders`, suíte de integração PostgreSQL e correções iniciais no Flutter).
- **Fase 2 — Simplificação & Modernização (`BL-07`, `BL-08`, `BL-09`):** Merged em `develop` via PR `#2` (`StoragePort` + `LocalStorage` sem Redis/R2 obrigatórios, `RateLimiter` em memória, sessão Web com cookies `HttpOnly` + Double-Submit CSRF, fila de refresh no Flutter e esteira DevSecOps completa).
- **Fase 3 — Identidade, Multi-Tenancy & LGPD (`BL-10`, `BL-11`, `BL-12`, `ADR-006`):** Merged em `develop` (`e8d92bc`) via PR `#3` (`organizations`, `memberships`, `trainer_student_relationships`, convites via `MailerPort`, reset de senha anti-enumeração, MFA TOTP com anti-replay, CPF opcional e Central LGPD).
- **Fase 4 — Biblioteca de Exercícios (`BL-13`, `RF-05`, `ADR-005`):** Merged em `develop` (`2b1d8b9`) via PR `#4` (catálogo PT-BR global `free-exercise-db` + customizados por organização, importador idempotente `ImportFreeExerciseDB`, mídia `PENDING`/`CONFIRMED` com `HeadObject` e URLs assinadas e telas Flutter `/catalog`).
- **Fase 5 — Workout Builder (`BL-14`):**
  - **Banco de Dados & Backfill (`migrations/009_workout_templates_and_versions`):**
    - Tabelas `workout_templates`, `workout_versions` (snapshots `JSONB` imutáveis protegidos pelo trigger `trg_workout_versions_immutable` no PostgreSQL) e `workout_assignments` (vínculo de atribuição versionada para alunos com índice parcial único em `legacy_workout_plan_id`), além de FK opcional `exercises.catalog_exercise_id`.
    - Backfill idempotente dos `workout_plans` legados para templates + `v1` + atribuições sem quebrar endpoints anteriores.
  - **Domínio, Aplicação, Postgres & HTTP (`workout`):**
    - Endpoints `/v1/workout-templates` (CRUD de rascunho, clonagem `POST /:id/clone`, publicação imutável `POST /:id/publish`, histórico `/versions`, atribuição `POST /:id/assign`) e `/v1/workout-assignments` (`GET ""`, `GET /:id`, `PATCH /:id/status`) com isolamento multi-tenant, validação de vínculo `ACTIVE`, validação cross-tenant de exercícios customizados e auditoria.
    - Alinhamento de playback de mídia: `/v1/workouts/:id` expõe `has_video` (derivado da tabela `videos` com status `CONFIRMED`) e `catalog_exercise_id` opcional.
  - **Flutter (`fit_brother_app`):**
    - Rotas `/workout-builder`, `/workout-builder/new`, `/workout-builder/:id` e `/workout-assignments/:id` com `WorkoutTemplatesScreen`, `TemplateEditorScreen`, `AssignmentDetailScreen` e `workout_media_dialog.dart`.
    - Atualização de `PlanDetailScreen`, `WorkoutPlansScreen` e `DashboardScreen` para consumir versões publicadas e resolver URLs assinadas (`/v1/videos/exercises/:id/signed-url` e `/v1/catalog/exercises/:id/media/:media_id/signed-url`).

---

## 2. Estado do Pipeline e Próximos Passos

1. **Esteira CI (`.github/workflows/ci.yml`):**
   - Pipeline `6/6` validada na run [`46015141865`](https://github.com/SCAPUTO88/fit-brother-2/actions/runs/46015141865) (`8d86a43`, detalhes em [`../06-quality/test-results.md`](../06-quality/test-results.md)).
2. **Revisão e merge:**
   - PR [`#5`](https://github.com/SCAPUTO88/fit-brother-2/pull/5) (`feat/fase-5-workout-builder` → `develop`) pronto para revisão e aprovação de merge pelo usuário antes de avançar para a **Fase 6 — Experiência do Aluno**.
