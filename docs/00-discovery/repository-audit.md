# Auditoria do Repositório — Fase 0

**Data da auditoria:** 2026-11-23  
**Repositório inspecionado:** `github.com/SCAPUTO88/fit-brother` (clonado localmente a partir do bundle `fit-brother.bundle`)  
**Branch auditada (HEAD):** `develop` (`16a8e6934475b3bf1491603926c8a3955b87043a`)

---

## 1. Resumo da Inspeção Git

| Referência no Bundle | Commit | Data | Descrição / Estado |
| :--- | :--- | :--- | :--- |
| `HEAD -> develop`, `origin/develop`, `origin/HEAD` | `16a8e69` | 2026-09-15 | `prod env adjustments` — código mais completo e atual |
| `origin/main` | `565e121` | 2026-04-14 | `Merge pull request #1 from SCAPUTO88/develop` — parado na Sprint 2 |
| Tags | — | — | Nenhuma tag de release encontrada |

### Conclusão sobre Branch Base
- `origin/main` (`565e121`) contém apenas as Sprints 0, 1 e 2 (Auth base + módulo inicial de Students). Faltam integralmente as Sprints 3 (Workouts), 4 (R2 Video), 5 (Flutter Web + CORS), scripts locais e ajustes de ambiente de produção/Cloud Run.
- `develop` (`16a8e69`) concentra todos os **30 commits** do repositório até o momento e é a única base viável para retomada (ver [ADR-001](../adr/ADR-001-base-branch-resumption.md)).

---

## 2. Cronologia dos Commits Principais (`develop`)

| Fase / Marco | Commits | Resumo do Entregue |
| :--- | :--- | :--- |
| **Sprint 0 — Fundação** | `3362d15` | Estrutura Go, `cmd/api/main.go`, `config.go`, `logger.go`, `Dockerfile`, `docker-compose.yml`, `.env.example`, `.github/workflows/ci.yml`, `README.md`, `SAD.md`. |
| **Sprint 1 — Auth** | `79e703e`, `54f71ad` | Migrations `001`–`002`, JWT RS256, Crypto (AES-256-GCM, HMAC-SHA256, bcrypt+pepper), `/v1/auth/{login,refresh,logout}`, testes de infra. |
| **Sprint 2 — Students** | `388dac9`, `1417405`, `2a2ed0b`, `f822dcd`..`565e121` | Migration `003`, CRUD parcial + anonimização LGPD de alunos, ajustes de CI/gofmt/Makefile, PR #1 para `main`. |
| **Sprint 3 — Workouts** | `a02b953`, `16a16ae` | Migration `004`, domínio/usecases/repositório/handler de `workout_plans`, `workout_sessions` e `exercises`. |
| **Sprint 4 — Videos R2** | `0bf7d1f`, `b0fc3d1` | Migration `005`, `R2Storage` via AWS SDK v2 S3-compatible, presigned upload/playback URLs e correção `staticcheck SA1019`. |
| **Sprint 5 — Flutter & Prod Prep** | `05d676b`, `ff06721`, `3684bad`, `cbb2442`, `36fe5db`, `8861815`..`16a8e69` | App Flutter (`fit_brother_app`), CORS configurável, scripts (`migrate.go`, `seed_trainer.go`, `debug_crypto.go`), atualização de `README.md`/`SAD.md`, ajustes CI/Compose/Cloud Run/CORS/env. |

---

## 3. Auditoria de Segredos no Histórico Git

Varredura completa nos 30 commits (`git log`) para detecção de chaves privadas, tokens de nuvem, senhas de produção e credenciais R2/Supabase/GCP:
- **Resultado:** Nenhum segredo real de produção, chave privada PEM, service account JSON GCP ou credencial R2 foi encontrado no histórico.
- **Achados locais / CI (valores dummy de desenvolvimento):**
  - `postgres://fitbrother:secret@...` em `.env.example`, `docker/docker-compose.yml`, `scripts/migrate.go`, `.github/workflows/ci.yml` e histórico `3362d15`/`3684bad`.
  - Secrets dummy explícitos de CI em `.github/workflows/ci.yml:113-122` (`AES_KEY`, `CPF_HMAC_SECRET`, `PASSWORD_PEPPER`, `R2_*`).
  - Senha padrão local hardcoded em `scripts/seed_trainer.go:40` e `scripts/debug_crypto.go:30` (`"Senha123!"`).
- **Ponto crítico de segurança em script:** `scripts/debug_crypto.go:25-26` imprime `PASSWORD_PEPPER` em hexadecimal e string pura no `stdout` quando executado.

---

## 4. Inconsistências Entre Documentação, Configuração e Código

| Item | Evidência 1 | Evidência 2 | Divergência Identificada |
| :--- | :--- | :--- | :--- |
| **Destino de Deploy Backend** | `docs/architecture/SAD.md:40`, `README.md:36` citam **Google Cloud Run** | `docker/docker-compose.yml:4` e `internal/infrastructure/config/config_test.go:5` citam **Railway** | Documentação e comentários de código divergem sobre provedor de produção. |
| **Link de Roadmap no README** | `README.md:24,186` aponta para `docs/plano-de-ataque.md` | Diretório `docs/` no repositório contém apenas `docs/architecture/SAD.md` | Link quebrado; `docs/plano-de-ataque.md` não existe no Git. |
| **Versão do Go** | `go.mod:3` declara `go 1.25.0` | `.github/workflows/ci.yml:10`, `docker/Dockerfile:2`, `README.md:31` usam `1.26.6` | Divergência menor de versão de toolchain vs `go.mod`. |
| **Redis obrigatório vs não utilizado** | `config.go:77` (`requireEnv("REDIS_URL")`), `docker-compose.yml:71`, `ci.yml:85`, `SAD.md:37` | `go.mod` e todo o código em `internal/` e `cmd/` | **Não existe nenhum cliente Redis importado ou usado em runtime**; contudo, a API entra em falha fatal no boot se `REDIS_URL` estiver vazio. |
| **Contrato de Upload de Vídeos (API vs Flutter)** | Backend `video_handler.go`: `POST /videos/upload-url` (`{exercise_id, filename, mime_type}`) e `POST /videos` (`{exercise_id, r2_key, ...}`) | Flutter `create_exercise_screen.dart:80-108`: chama `POST /videos/request-upload` e `POST /videos/confirm` com outros campos | **Integração de upload de vídeo no Flutter está quebrada por incompatibilidade de contrato.** |
| **`.gitignore` bloqueando documentação** | `.gitignore:50-51` lista `AGENTS.md` e `docs/` | `SAD.md` já era rastreado antes da regra no `.gitignore` | Novos arquivos em `docs/` e `AGENTS.md` ficam ignorados pelo Git até ajuste explícito do `.gitignore`. |
| **Testes de Integração** | `README.md:159` e `ci.yml:123` executam `./tests/integration/...` | `tests/integration/health_test.go:1-15` | O arquivo contém apenas `t.Log("Ambiente de Integração Configurado e Operacional")`; não há testes reais contra PostgreSQL/Redis. |
| **Metadados Web do Flutter** | `SAD.md:14` define foco Web/PWA com estética moderna | `fit_brother_app/web/index.html:26-29` e `web/manifest.json:2-8` | Boilerplate padrão (`"A new Flutter project."`, `#0175C2`, `fit_brother_app`). |

---

## 5. Inventário de Arquivos Estruturais Inspecionados

- **Backend Go (`51` arquivos principais de código/teste/migration/script):**
  - Entrypoint e config: `cmd/api/main.go`, `internal/infrastructure/config/config.go`, `logger/logger.go`, `jwt/service.go`, `crypto/{aes,hash,password}.go`.
  - Domínio: `internal/domain/{user,student,workout}/{entity,repository}.go`. *(Observação: `video` não possui pacote próprio em `domain`; sua interface/entidade foi colocada em `workout` ou acoplada em `application/video`).*
  - Aplicação: `internal/application/{auth,student,workout,video}/*`.
  - Adaptadores: `internal/adapters/{http/{router,middleware/auth_middleware,handler/*},repository/postgres/*,storage/r2_storage}.go`.
  - Migrations: `migrations/001` a `005` (`.up.sql` e `.down.sql`).
  - Scripts: `scripts/{migrate,seed_trainer,gen_hash,verify_hash,debug_crypto}.go`, `scripts/start_backend.ps1`.
- **Frontend Flutter (`fit_brother_app/`):**
  - `pubspec.yaml`, `analysis_options.yaml`, `web/{index.html,manifest.json}`.
  - `lib/main.dart`, `lib/core/{api/dio_client.dart,router/app_router.dart}`.
  - `lib/features/auth/{auth_provider,login_screen}.dart`.
  - `lib/features/dashboard/dashboard_screen.dart`.
  - `lib/features/students/{students_provider,create_student_screen}.dart`.
  - `lib/features/workouts/{workouts_provider,workout_plans_screen,plan_detail_screen,create_plan_screen,create_session_screen,create_exercise_screen}.dart`.
  - Testes: `test/widget_test.dart`, `test/core/api/dio_client_test.dart`.
