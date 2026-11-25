# Estado Atual do Projeto (Checkpoint de Handoff)

**Última atualização:** 2026-11-25  
**Fase atual:** **Fase 2 — Simplificação & Modernização (Implementada, validada com 6/6 jobs verdes no GitHub Actions e aberta em PR `#2` para `develop`)**  
**Repositório conectado:** `https://github.com/SCAPUTO88/fit-brother-2`  
**Branch ativa:** `feat/fase-2-simplificacao-modernizacao` → **PR `#2`** (`https://github.com/SCAPUTO88/fit-brother-2/pull/2`) para `develop`  
**Branch `origin/main`:** `565e121` (Sprint 2 — aguardando promoção controlada após validação em `develop`)

---

## 1. Resumo das Fases Concluídas

- **Fase 1 — Security Baseline:** Merged em `develop` via PR `#1` (isolamento trainer/aluno com `migrations/006`, transações atômicas em Auth/Student/Anonymize, `HeadObject` em vídeos, `SecurityHeaders`, suíte de integração PostgreSQL e correções iniciais no Flutter).
- **Fase 2 — Simplificação & Modernização (`BL-07`, `BL-08`, `BL-09`):**
  - **BL-07 — `StoragePort`, Configuração & Rate Limiter em Memória:**
    - Extraído `domain/storage.Port` (Hexagonal), desacoplando `application/video` de `adapters/storage` e `R2Config`.
    - Implementados `LocalStorage` (URLs assinadas via HMAC-SHA256, proteção contra path traversal, sidecar `.meta.json`, endpoints `/v1/storage/local`), `DisabledStorage` (`503 Service Unavailable`) e `R2Storage`.
    - `config.Load()` simplificado: `REDIS_URL` opcional/não fatal, `STORAGE_DRIVER=local|r2|disabled` (default `local` em dev/test, `disabled` em produção sem R2, `r2` automático quando credenciais presentes) e retorno explícito de erro em vez de `panic`.
    - Middleware `RateLimiter` em memória (`sync.Mutex`, janela fixa configurável por IP com limpeza automática de entradas expiradas e `Retry-After`) aplicado em `/v1/auth/login` e `/v1/auth/refresh`.
  - **BL-08 — Sessão Web via Cookies `HttpOnly` + CSRF e Fila de Refresh no Flutter:**
    - `AuthHandler` suporta `X-Session-Mode: cookie` emitindo `fitbrother_access_token` (`HttpOnly`), `fitbrother_refresh_token` (`HttpOnly`) e `fitbrother_csrf` sem expor tokens no JSON em modo Web.
    - `RequireAuth` aceita `Authorization: Bearer` ou cookie `HttpOnly`; middleware `CSRFProtection` valida `Origin`/`Referer` + Double-Submit Cookie com comparação em tempo constante.
    - CORS endurecido rejeitando `*` quando `AllowCredentials=true` e liberando `X-CSRF-Token` e `X-Session-Mode`.
    - No Flutter (`fit_brother_app`): abstração `TokenStorage` (`NoOpTokenStorage` na Web sem `localStorage`, `SecureTokenStorage` no mobile), adaptador `dart:js_interop` puro (`web_cookie_web.dart`) e fila serializada com `Completer<bool>` no `DioClient` para requisições `401` concorrentes, validada com testes unitários (`dio_client_test.dart`).
  - **BL-09 — Saída do GCP & Esteira CI/DevSecOps:**
    - Removidos `redis` e `REDIS_URL` de `docker/docker-compose.yml`, `.env.example`, `Makefile`, `README.md` e `.github/workflows/ci.yml`; `adminer` fixado em `4.8.1`; banner de substituição adicionado em `docs/architecture/SAD.md` e seção 1 marcada em `docs/05-infrastructure/gcp-decommission-checklist.md`.
    - CI `.github/workflows/ci.yml` com `permissions: contents: read`, versões fixadas (`golangci-lint v1.64.8`, `govulncheck v1.2.0`, `gitleaks 8.30.1`, `osv-scanner 2.7.1`, `semgrep 1.186.0`, `trivy-action v0.36.0`) e publicação do artefato SBOM CycloneDX (`fit-brother-sbom-cdx`).
    - Atualizado `golang.org/x/net` para `v0.62.0` (e dependências `golang.org/x/*` alinhadas via MVS) corrigindo `CVE-2026-78661`.

---

## 2. Estado do Pipeline e Próximos Passos

1. **Esteira CI (`.github/workflows/ci.yml`):**
   - Todos os **6 jobs** (`Lint (Go)`, `Security Scan (Go)`, `DevSecOps (Secrets, SAST, SCA & SBOM)`, `Test (Go + Postgres)`, `Flutter (Analyze, Test & Web Build)` e `Build (Go & Docker)`) aprovados na run [`45715307808`](https://github.com/SCAPUTO88/fit-brother-2/actions/runs/45715307808) (detalhes em [`../06-quality/test-results.md`](../06-quality/test-results.md)).
2. **Revisão e merge:**
   - PR [`#2`](https://github.com/SCAPUTO88/fit-brother-2/pull/2) (`feat/fase-2-simplificacao-modernizacao` → `develop`) pronto para revisão e aprovação de merge pelo usuário antes de avançar para a **Fase 3 — Identidade, Multi-Tenancy & LGPD**.

