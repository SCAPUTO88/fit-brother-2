# Estado Atual do Projeto (Checkpoint de Handoff)

**Última atualização:** 2026-11-24  
**Fase atual:** **Fase 1 — Security Baseline (Implementada, validada com 5/5 jobs verdes no GitHub Actions e aberta em PR `#1` para `develop`)**  
**Repositório conectado:** `https://github.com/SCAPUTO88/fit-brother-2`  
**Branch ativa:** `fix/fase-1-security-baseline` → **PR `#1`** (`https://github.com/SCAPUTO88/fit-brother-2/pull/1`) para `develop`  
**Branch `origin/main`:** `565e121` (Sprint 2 — aguardando promoção controlada após validação em `develop`)

---

## 1. O que foi implementado na Fase 1

- **Governança e Decisões:**
  - `.gitignore` atualizado conforme aprovado em `DEC-01` para versionar `AGENTS.md` e `docs/`.
  - `docs/07-project/decisions.md` atualizado com a aprovação de `DEC-01`–`DEC-09` em `2026-11-24`.
- **Entrega 1A — Isolamento de Dados e Banco (`migrations/006` + `student`):**
  - Criada migration `006_security_baseline_ownership_and_video_status` adicionando `students.trainer_id`, índice composto `(trainer_id, student_id)` em `workout_plans` e `videos.status` (`PENDING` / `CONFIRMED`).
  - Implementado `TxManager` em PostgreSQL e aplicadas transações atômicas em `CreateStudentUseCase` e `AnonymizeStudentUseCase`.
  - Protegido `POST /v1/students` para exigir autenticação com perfil de treinador, vinculando automaticamente `trainer_id` pela sessão e retornando `201 Created`.
  - Listagem (`ListByTrainer`), consulta detalhada e anonimização de alunos restritas ao treinador responsável ou ao próprio aluno; anonimização agora limpa `email`, `full_name`, `password_hash`, define `deleted_at`, revoga todos os `refresh_tokens` e remove o registro em `students`.
- **Entrega 1B — Treinos, Autenticação e Cabeçalhos HTTP:**
  - Corrigida resolução de identidade entre `users.id` e `students.id` em `ListPlans.ByStudent` e `GetPlanDetailsUseCase`, permitindo que alunos autenticados visualizem seus próprios treinos e bloqueando acesso cruzado.
  - Adicionada validação de vínculo e propriedade em `CreatePlanUseCase`, `CreateSessionUseCase` e `CreateExerciseUseCase`.
  - Mapeamento padronizado de erros de domínio para respostas HTTP `400`/`401`/`403`/`404`/`409`/`500` genérico sem exposição de detalhes internos.
  - `RefreshUseCase` tornado transacional com bloqueio de linha (`FOR UPDATE`), checagem de usuário ativo antes da rotação e revogação persistente da família de tokens em caso de reuso.
  - `RequireAuth` verifica em tempo real se o usuário está ativo e não excluído; adicionado middleware `SecurityHeaders`.
- **Entrega 1C — Vídeos, Scripts, Testes e Flutter:**
  - Upload de vídeos passou a registrar intenção `PENDING`, validar MIME (`video/mp4`, `video/webm`, `video/quicktime`) e propriedade sobre o exercício, gerar chave controlada exclusivamente no servidor e verificar existência/tamanho (até 50 MB) via `HeadObject` no R2 na confirmação (`POST /v1/videos`).
  - `GET /v1/videos/exercises/:exercise_id/signed-url` passa a validar posse pelo treinador ou aluno vinculado.
  - Removidos segredos e senhas fixas de `scripts/debug_crypto.go` e `scripts/seed_trainer.go`.
  - Criada suíte de integração real com PostgreSQL (`tests/integration/security_baseline_test.go`) e expandidos os testes unitários/HTTP.
  - Corrigidos no app Flutter (`fit_brother_app`): validação de tokens no `auth_provider.dart`, envio de `refresh_token` no `logout()`, cache reativo por sessão, guarda de rotas autenticadas em `app_router.dart` e fluxo correto de criação de exercício + upload de vídeo em `create_exercise_screen.dart`.

---

## 2. Estado do Pipeline e Próximos Passos

1. **Esteira CI (`.github/workflows/ci.yml`):**
   - Pipeline ampliado para cobrir `Lint (Go)`, `Security Scan (Go)` (`govulncheck`), `Test (Go + Postgres)` (unitários `-race`, integração PostgreSQL `001`–`006` e cobertura), `Flutter (Analyze, Test & Web Build)` e `Build (Go & Docker)`.
   - Todos os 5 jobs validados com sucesso na run [`45461877292`](https://github.com/SCAPUTO88/fit-brother-2/actions/runs/45461877292) (detalhes em [`../06-quality/test-results.md`](../06-quality/test-results.md)).
2. **Revisão e merge:**
   - PR [`#1`](https://github.com/SCAPUTO88/fit-brother-2/pull/1) (`fix/fase-1-security-baseline` → `develop`) pronto para aprovação e merge pelo usuário antes de avançar para a **Fase 2 — Workout & Video Core**.
