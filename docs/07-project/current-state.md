# Estado Atual do Projeto (Checkpoint de Handoff)

**Última atualização:** 2026-11-26  
**Fase atual:** **Fase 4 — Biblioteca de Exercícios (Implementada, validada no GitHub Actions e aberta em PR `#4` para `develop`)**  
**Repositório conectado:** `https://github.com/SCAPUTO88/fit-brother-2`  
**Branch ativa:** `feat/fase-4-biblioteca-exercicios` → **PR `#4`** (`https://github.com/SCAPUTO88/fit-brother-2/pull/4`) para `develop`  
**Branch `origin/main`:** `565e121` (Sprint 2 — aguardando promoção controlada após validação em `develop`)

---

## 1. Resumo das Fases Concluídas

- **Fase 1 — Security Baseline:** Merged em `develop` via PR `#1` (isolamento trainer/aluno com `migrations/006`, transações atômicas em Auth/Student/Anonymize, `HeadObject` em vídeos, `SecurityHeaders`, suíte de integração PostgreSQL e correções iniciais no Flutter).
- **Fase 2 — Simplificação & Modernização (`BL-07`, `BL-08`, `BL-09`):** Merged em `develop` via PR `#2` (`StoragePort` + `LocalStorage` sem Redis/R2 obrigatórios, `RateLimiter` em memória, sessão Web com cookies `HttpOnly` + Double-Submit CSRF, fila de refresh no Flutter e esteira DevSecOps completa).
- **Fase 3 — Identidade, Multi-Tenancy & LGPD (`BL-10`, `BL-11`, `BL-12`, `ADR-006`):** Merged em `develop` (`e8d92bc`) via PR `#3` (`organizations`, `memberships`, `trainer_student_relationships`, convites via `MailerPort`, reset de senha anti-enumeração, MFA TOTP com anti-replay, CPF opcional e Central LGPD).
- **Fase 4 — Biblioteca de Exercícios (`BL-13`, `RF-05`, `ADR-005`):**
  - **Banco de Dados & Seed (`migrations/008_exercise_catalog_and_media`):**
    - Tabelas `muscle_groups` (19 grupos em PT-BR/EN), `equipments` (12 equipamentos em PT-BR/EN), `exercise_catalog` (índice único parcial para upsert global sem colidir com exercícios customizados por organização), `exercise_media` (`VIDEO`, `IMAGE`, `GIF` com `PENDING`/`CONFIRMED`) e FK opcional `exercises.catalog_exercise_id`.
    - Seed inicial de 32 exercícios clássicos PT-BR oriundos de `yuhonas/free-exercise-db` (`The Unlicense`).
  - **Domínio, Aplicação, CLI & HTTP (`catalog`):**
    - Tradutor determinístico PT-BR (`translator.go`), importador idempotente (`ImportFreeExerciseDB` + CLI `cmd/import-exercises`), busca/filtros multi-tenant e CRUD customizado por organização.
    - Endpoints `/v1/catalog/{filters,exercises,exercises/:id,exercises/:id/media/upload-url,exercises/:id/media/confirm,exercises/:id/media/:media_id/signed-url}` protegidos por JWT/cookie e validação de vínculo/organização.
    - Fluxo seguro de mídia (`50MB` vídeo, `10MB` imagem, `15MB` GIF) com chave gerada no servidor (`catalog/{org}/{exercise}/{media}{ext}`), confirmação via `HeadObject`, URL assinada temporária e eventos de auditoria.
  - **Flutter (`fit_brother_app`):**
    - Rotas `/catalog`, `/catalog/create` e `/catalog/:id` com telas `ExerciseCatalogScreen`, `ExerciseDetailScreen` e `CreateCatalogExerciseScreen`, integradas ao `DashboardScreen` (Treinador e Aluno) e validadas por widget tests.

---

## 2. Estado do Pipeline e Próximos Passos

1. **Esteira CI (`.github/workflows/ci.yml`):**
   - Pipeline `6/6` validada na run [`45840454506`](https://github.com/SCAPUTO88/fit-brother-2/actions/runs/45840454506) (`fed5545`, detalhes em [`../06-quality/test-results.md`](../06-quality/test-results.md)).
2. **Revisão e merge:**
   - PR [`#4`](https://github.com/SCAPUTO88/fit-brother-2/pull/4) (`feat/fase-4-biblioteca-exercicios` → `develop`) pronto para revisão e aprovação de merge pelo usuário antes de avançar para a **Fase 5 — Builder de Treinos & Versionamento**.
