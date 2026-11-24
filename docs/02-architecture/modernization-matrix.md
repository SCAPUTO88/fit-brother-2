# Matriz de Modernização (KEEP / UPDATE / REFACTOR / REPLACE / REMOVE / ADD / DEFER)

| Item / Componente | Decisão | Justificativa Técnica e de Negócio | Fase |
| :--- | :--- | :--- | :--- |
| **Backend Go + Gin + sqlx/pgx** | **KEEP** | Base limpa, rápida, compilada em binário único leve e ideal para free-tier. | — |
| **PostgreSQL + golang-migrate (`001`–`005`)** | **KEEP** | Padrão SQL portátil (evita lock-in Supabase JS SDK); preservar histórico `001`–`005` intacto. | — |
| **Flutter Web/PWA (Riverpod + GoRouter + Dio)** | **KEEP** | Já existe no repo, entrega Web/PWA e preserva caminho para Android/iOS futuros com 1 codebase. | — |
| **Branch `develop` (`16a8e69`) como base** | **KEEP** | Único ponto íntegro com Sprints 0–5. (`main` está parada na Sprint 2). | Fase 0/1 |
| **Crypto AES-256-GCM + Bcrypt+Pepper + HMAC** | **KEEP / UPDATE** | Boa base criptográfica; adicionar suporte futuro opcional a Argon2id e `kid` no JWT. | Fase 1–3 |
| **Dependências Go (`go.mod`) e Flutter (`pubspec.yaml`)** | **UPDATE** | Alinhar `go.mod` com versão alvo do CI e revisar `pubspec.yaml` após diagnóstico local. | Fase 2 |
| **GitHub Actions (`.github/workflows/ci.yml`)** | **UPDATE** | Fixar versões (sem `latest` solto), restringir `permissions`, incluir Flutter + Gitleaks + Semgrep + OSV + Trivy + SBOM. | Fase 2 |
| **Autenticação / Sessão Web (`auth`, `DioClient`, `flutter_secure_storage`)** | **REFACTOR** | Transações no refresh, `logout` no Flutter, refresh automático, `is_active` check e cookies `HttpOnly; SameSite` via mesmo domínio/BFF para Web. | Fase 1–2 |
| **Módulo Students & Workouts (`handler`, `usecases`, `repo`)** | **REFACTOR** | Corrigir `POST /students` sem auth, `users.id != students.id`, ownership checks, transações e separar templates/versões/execuções. | Fase 1, 3, 5 |
| **Módulo Videos (`video/usecases.go` + `r2_storage.go` + Flutter)** | **REFACTOR** | Extrair `StoragePort` para `domain`, tornar R2 opcional no boot, adicionar `HeadObject` + ownership + alinhar contrato Flutter. | Fase 1–2 |
| **Testes de Integração (`tests/integration/health_test.go`)** | **REPLACE** | Substituir stub vazio por testes reais contra PostgreSQL descartável / Testcontainers. | Fase 1 |
| **Google Cloud Run + Refs GCP (`SAD.md`, `README.md`)** | **REPLACE** | Sair integralmente do GCP para PaaS fora do ecossistema Google (Render Free / Fly / Railway) mantendo Dockerfile padrão. | Fase 2 / 9 |
| **`REDIS_URL` obrigatório no boot (`config.go:77`)** | **REMOVE** | Redis não é usado em nenhuma linha de código; rate limit inicial funcionará em memória/Postgres. | Fase 2 |
| **CPF obrigatório em `students` (`migrations/003` + `create_student`)** | **REMOVE** | Viola minimização LGPD para app de treino e eleva risco de vazamento. | Fase 3 |
| **`scripts/debug_crypto.go` (print de pepper) e senha fixa em `seed_trainer.go`** | **REMOVE / REFACTOR** | Eliminar exposição de `PASSWORD_PEPPER` em `stdout` e exigir leitura segura ou geração aleatória em seed. | Fase 1 |
| **Multi-tenancy (`organizations`, `memberships`, `relationships`)** | **ADD** | Base estrutural de isolamento entre treinadores, estúdios e alunos. | Fase 3 |
| **Convites, Ativação, Reset de Senha + SMTP (Resend/Brevo)** | **ADD** | Fluxo seguro de onboarding do aluno sem treinador definir senha manualmente. | Fase 3 / 9 |
| **Catálogo de Exercícios (`free-exercise-db` Unlicense)** | **ADD** | Biblioteca independente em PT-BR com grupos musculares, equipamentos e mídias licenciadas. | Fase 4 |
| **Experiência Mobile do Aluno, Avaliações e Gráficos** | **ADD** | Telas e endpoints de execução, cronômetro, séries, histórico, medidas e dashboard operacional. | Fase 6–8 |
| **Apps Nativos Lojas, Fotos Evolução, Chat, Push, Wearables, IA, Pagamentos** | **DEFER** | Documentados em `mvp-scope.md`; bloqueados até estabilização do MVP. | Pós-MVP |
