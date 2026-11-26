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

---

## Fase 3 — Identidade, Multi-Tenancy & LGPD (`feat/fase-3-identidade-multitenancy-lgpd`, 2026-11-26)

- **Ambiente de execução real:** GitHub Actions (`ubuntu-latest`) no repositório `SCAPUTO88/fit-brother-2` (PR `#3`, commit `e41c7de`, workflow run [`45735271644`](https://github.com/SCAPUTO88/fit-brother-2/actions/runs/45735271644)).
- **Resultados verificados na esteira CI (`6/6` jobs aprovados):**
  1. **`Lint (Go)` (`success`):** `golangci-lint` (`v1.64.8`) sem issues (`0`) e `gofmt -l .` limpo.
  2. **`Security Scan (Go)` (`success`):** `govulncheck` (`v1.2.0`) com Go `1.26.11` sem vulnerabilidades reportadas.
  3. **`DevSecOps (Secrets, SAST, SCA & SBOM)` (`success`):**
     - `Gitleaks` (`8.30.1`): `no leaks found`.
     - `Semgrep` (`1.186.0`): `0` findings em `221` arquivos (`80` regras executadas).
     - `OSV-Scanner` (`2.7.1`): `70` pacotes Go (`go.mod`) e `118` pacotes Dart/Flutter (`fit_brother_app/pubspec.lock`) analisados sem vulnerabilidades conhecidas.
     - `Trivy`: SBOM CycloneDX gerado e publicado como artefato `fit-brother-sbom-cdx` (`9638 bytes`); Trivy FS (`HIGH`/`CRITICAL`) limpo em `go.mod` e `fit_brother_app/pubspec.lock`.
  4. **`Test (Go + Postgres)` (`success`):**
     - `go test -v -race -count=1 ./internal/...` aprovado em todos os `19` pacotes testáveis (incluindo `application/{auth,invitation,organization,privacy,student,video,workout}`, `infrastructure/totp` e `adapters/mailer`).
     - `go test -v -race -count=1 -timeout 3m -tags integration ./tests/integration/...` aprovado com PostgreSQL 16 (`52.377s` pacote): `TestSecurityBaselinePostgresIntegration` (`51.36s`), cobrindo `Fluxo_Web_Cookie_+_CSRF_e_LocalStorage_sem_Redis/R2` (`2.52s`) e `Fase_3_-_Identidade_Multi-Tenancy_Convites_MFA_Reset_e_LGPD` (`33.28s`, aplicando migrations `001`–`007` e validando criação sem CPF, organizações/membros, encerramento de vínculo, convite/revogação/aceite, reset de senha com revogação de sessões, consentimentos/exportação/auto-anonimização LGPD e fluxo completo de MFA TOTP com anti-replay).
     - **Cobertura unitária (`go tool cover`):** `36.9%` total de statements (`domain/student` `93.8%`, `infrastructure/totp` `91.9%`, `domain/workout` `86.4%`, `infrastructure/jwt` `79.3%`, `infrastructure/crypto` `79.2%`, `adapters/http/middleware` `73.4%`, `application/video` `73.0%`, `infrastructure/config` `67.6%`, `application/privacy` `67.5%`, `application/organization` `61.4%`, `adapters/storage` `59.6%`, `application/student` `58.7%`, `application/workout` `56.7%`, `application/invitation` `55.6%`, `application/auth` `55.3%`, `adapters/mailer` `18.2%`, `adapters/http` `7.2%`, `adapters/http/handler` `6.6%`, `adapters/repository/postgres` `3.6%`).
  5. **`Flutter (Analyze, Test & Web Build)` (`success`):** `flutter analyze --no-fatal-infos` (sem warnings/errors; apenas `14` `info`s), `flutter test` (`8` testes aprovados em `test/core/api/dio_client_test.dart`, `test/dio_client_test.dart` e `test/widget_test.dart`) e `flutter build web --release` (`✓ Built build/web`).
  6. **`Build (Go & Docker)` (`success`):** compilação estática de `./cmd/api` e build da imagem `docker/Dockerfile`.

---

## Fase 4 — Biblioteca de Exercícios (`feat/fase-4-biblioteca-exercicios`, 2026-11-26)

- **Ambiente de execução real:** GitHub Actions (`ubuntu-latest`) no repositório `SCAPUTO88/fit-brother-2` (PR `#4`, commit `fed5545`, workflow run [`45840454506`](https://github.com/SCAPUTO88/fit-brother-2/actions/runs/45840454506)).
- **Resultados verificados na esteira CI (`6/6` jobs aprovados):**
  1. **`Lint (Go)` (`success`):** `golangci-lint` (`v1.64.8`) sem issues (`0`) e `gofmt -l .` limpo.
  2. **`Security Scan (Go)` (`success`):** `govulncheck` (`v1.2.0`) com Go `1.26.11` sem vulnerabilidades reportadas.
  3. **`DevSecOps (Secrets, SAST, SCA & SBOM)` (`success`):**
     - `Gitleaks` (`8.30.1`): `no leaks found` (~`1.07 MB` analisados).
     - `Semgrep` (`1.186.0`): `0` findings em `236` arquivos (`80` regras executadas).
     - `OSV-Scanner` (`2.7.1`): `70` pacotes Go (`go.mod`) e `118` pacotes Dart/Flutter (`fit_brother_app/pubspec.lock`) analisados sem vulnerabilidades conhecidas.
     - `Trivy`: SBOM CycloneDX gerado e publicado como artefato `fit-brother-sbom-cdx` (`9631` bytes); Trivy FS (`CRITICAL,HIGH`) limpo em `go.mod` e `fit_brother_app/pubspec.lock`.
  4. **`Test (Go + Postgres)` (`success`):**
     - `go test -v -race -count=1 ./internal/...` aprovado em todos os `20` pacotes testáveis (incluindo `internal/application/catalog`, com `importer_test.go` e `usecases_test.go`).
     - `go test -v -race -count=1 -timeout 3m -tags integration ./tests/integration/...` aprovado com PostgreSQL 16 (`64.877s` pacote): `TestSecurityBaselinePostgresIntegration` (`63.86s`), cobrindo `Fluxo_Web_Cookie_+_CSRF_e_LocalStorage_sem_Redis/R2` (`2.55s`), `Fase_3_-_Identidade_Multi-Tenancy_Convites_MFA_Reset_e_LGPD` (`32.83s`) e `Fase_4_-_Biblioteca_de_Exercicios_Catalogo_PT-BR_Importador_Idempotente_e_Midia_Segura` (`12.66s`, aplicando migrations `001`–`008` e validando filtros PT-BR, seed `The Unlicense`, importador idempotente em 2 rodadas contra PostgreSQL, criação/bloqueio multi-tenant por organização e vínculo ativo vs. `ENDED`, e upload/confirmação/URL assinada de mídia).
     - **Cobertura unitária (`go tool cover`):** `36.1%` total de statements (`internal/domain/student` `93.8%`, `internal/infrastructure/totp` `91.9%`, `internal/domain/workout` `86.4%`, `internal/infrastructure/jwt` `79.3%`, `internal/infrastructure/crypto` `79.2%`, `internal/adapters/http/middleware` `73.4%`, `internal/application/video` `73.0%`, `internal/infrastructure/config` `67.6%`, `internal/application/privacy` `67.5%`, `internal/application/catalog` `67.4%`, `internal/application/organization` `61.4%`, `internal/adapters/storage` `59.6%`, `internal/application/student` `58.7%`, `internal/application/workout` `56.7%`, `internal/application/invitation` `55.6%`, `internal/application/auth` `55.3%`, `internal/adapters/mailer` `18.2%`, `internal/adapters/http` `7.0%`, `internal/adapters/http/handler` `5.3%`, `internal/adapters/repository/postgres` `2.4%`).
  5. **`Flutter (Analyze, Test & Web Build)` (`success`):** `flutter analyze --no-fatal-infos` (sem warnings/errors; apenas `17` `info`s do SDK sobre `material.dart`), `flutter test` (`10 tests passed`, incluindo os widget tests da `ExerciseCatalogScreen` para Treinador e Aluno) e `flutter build web --release` (`✓ Built build/web`).
  6. **`Build (Go & Docker)` (`success`):** compilação estática de `./cmd/api` e build da imagem `docker/Dockerfile`.

