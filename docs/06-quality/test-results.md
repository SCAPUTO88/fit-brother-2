# Registro Contínuo de Resultados de Testes

Este arquivo registra os resultados executados ao longo das fases.

---

## Fase 0 — Baseline Inicial (2026-11-23)

- **Ambiente:** Contêiner isolado de auditoria (sem `go`, `flutter`, `dart`, `docker`, `psql`).
- **Status de execução:** Nenhum comando Go/Flutter/Docker pôde ser executado localmente neste contêiner. Detalhes completos em [`../00-discovery/baseline-test-results.md`](../00-discovery/baseline-test-results.md).

---

## Fase 1 — Security Baseline (`fix/fase-1-security-baseline`, 2026-11-24)

- **Ambiente de execução real:** GitHub Actions (`ubuntu-latest`) no repositório `SCAPUTO88/fit-brother-2` (PR `#1`, commit `785e7f9`, workflow run [`45461877292`](https://github.com/SCAPUTO88/fit-brother-2/actions/runs/45461877292)).
- **Resultados verificados na esteira CI (`5/5` jobs aprovados):**
  1. **`Lint (Go)` (`success`):** `golangci-lint` sem issues (`0`) e `gofmt -l .` limpo.
  2. **`Security Scan (Go)` (`success`):** `govulncheck ./...` com Go `1.26.11` e módulos `golang.org/x/{crypto@v0.57.0, net@v0.60.0, sync@v0.23.0, sys@v0.48.0, text@v0.42.0}` sem vulnerabilidades reportadas.
  3. **`Test (Go + Postgres)` (`success`):**
     - `go test -v -race -count=1 ./internal/...` aprovado em todos os pacotes testáveis (`application/{auth,student,video,workout}`, `domain/{student,workout}`, `adapters/http/{middleware,handler}`, `adapters/repository/postgres`, `infrastructure/{config,crypto,jwt}`).
     - `go test -v -race -count=1 -timeout 3m -tags integration ./tests/integration/...` aprovado com PostgreSQL 16 (`TestIntegrationSetup` e `TestSecurityBaselinePostgresIntegration` em `15.52s`, aplicando migrations `001`–`006` e validando isolamento de dados, treinos de aluno, upload/playback de vídeo e anonimização LGPD).
     - **Cobertura unitária (`go tool cover`):** `39.7%` total de statements (`domain/student` `93.8%`, `domain/workout` `86.4%`, `infrastructure/jwt` `79.3%`, `infrastructure/crypto` `79.2%`, `adapters/http/middleware` `75.7%`, `application/video` `74.2%`, `application/student` `72.0%`, `application/workout` `64.0%`, `application/auth` `51.9%`).
  4. **`Flutter (Analyze, Test & Web Build)` (`success`):** `flutter analyze --no-fatal-infos` (sem warnings/errors; apenas 10 `info` do SDK Flutter 3.50.1 sobre `material.dart`), `flutter test` (`widget_test.dart`) aprovado e `flutter build web --release` (`✓ Built build/web`).
  5. **`Build (Go & Docker)` (`success`):** compilação estática de `./cmd/api` e build da imagem `docker/Dockerfile` (`golang:1.26.11-alpine` → `distroless`).
- **Observação de correção capturada pelo CI:** O teste de integração com PostgreSQL detectou e permitiu corrigir um deadlock de lock `FOR UPDATE` em `RefreshUseCase.Execute` (a revogação da família comprometida passou a ocorrer após o rollback da transação).

---

## Fase 2 — Simplificação & Modernização (`feat/fase-2-simplificacao-modernizacao`, 2026-11-25)

- **Ambiente de execução real:** GitHub Actions (`ubuntu-latest`) no repositório `SCAPUTO88/fit-brother-2` (PR `#2`, commit `fc0f6db`, workflow run [`45715307808`](https://github.com/SCAPUTO88/fit-brother-2/actions/runs/45715307808)).
- **Resultados verificados na esteira CI (`6/6` jobs aprovados):**
  1. **`Lint (Go)` (`success`):** `golangci-lint` (`v1.64.8`) sem issues (`0`) e `gofmt -l .` limpo.
  2. **`Security Scan (Go)` (`success`):** `govulncheck` (`v1.2.0`) com Go `1.26.11` sem vulnerabilidades reportadas após atualização de `golang.org/x/net` para `v0.62.0` (e alinhamento MVS de `crypto@v0.58.0`, `sync@v0.24.0`, `sys@v0.49.0`, `text@v0.43.0`).
  3. **`DevSecOps (Secrets, SAST, SCA & SBOM)` (`success`):**
     - `Gitleaks` (`8.30.1`): `no leaks found`.
     - `Semgrep` (`1.186.0`): `0` findings (`0` blocking).
     - `OSV-Scanner` (`2.7.1`): `70` pacotes Go (`go.mod`) e `118` pacotes Dart/Flutter (`fit_brother_app/pubspec.lock`) analisados sem vulnerabilidades.
     - `Trivy`: SBOM CycloneDX gerado e publicado como artefato `fit-brother-sbom-cdx` (`9643 bytes`); Trivy FS (`HIGH`/`CRITICAL`) com `0` vulnerabilidades em `go.mod` e `fit_brother_app/pubspec.lock`.
  4. **`Test (Go + Postgres)` (`success`):**
     - `go test -v -race -count=1 ./internal/...` aprovado em todos os 14 pacotes testáveis (incluindo `adapters/storage` com `LocalStorage`, `adapters/http/middleware` com `RateLimiter` e `CSRFProtection`, `adapters/http` CORS e `infrastructure/config`).
     - `go test -v -race -count=1 -timeout 3m -tags integration ./tests/integration/...` aprovado com PostgreSQL 16 sem Redis nem R2 (`STORAGE_DRIVER=local`): `TestSecurityBaselinePostgresIntegration` em `17.85s` (`18.871s` pacote), incluindo o subteste `Fluxo_Web_Cookie_+_CSRF_e_LocalStorage_sem_Redis/R2` (`2.51s`).
     - **Cobertura unitária (`go tool cover`):** `42.4%` total de statements (`domain/student` `93.8%`, `domain/workout` `86.4%`, `infrastructure/jwt` `79.3%`, `infrastructure/crypto` `79.2%`, `adapters/http/middleware` `73.2%`, `application/video` `73.0%`, `application/student` `72.0%`, `infrastructure/config` `70.0%`, `application/workout` `64.0%`, `adapters/storage` `59.6%`, `application/auth` `51.9%`, `adapters/http/handler` `12.2%`, `adapters/http` `8.8%`, `adapters/repository/postgres` `6.8%`).
  5. **`Flutter (Analyze, Test & Web Build)` (`success`):** `flutter analyze --no-fatal-infos` (sem warnings/errors; apenas 10 `info` do SDK sobre `material.dart`), `flutter test` (`6` testes aprovados em `test/dio_client_test.dart`, `test/core/api/dio_client_test.dart` e `test/widget_test.dart`) e `flutter build web --release` (`✓ Built build/web`).
  6. **`Build (Go & Docker)` (`success`):** compilação estática de `./cmd/api` e build da imagem `docker/Dockerfile`.

