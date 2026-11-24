# Infraestrutura Atual (AS-IS)

Diagnóstico do desenho de infraestrutura observado nos arquivos atuais (`docker/docker-compose.yml`, `docker/Dockerfile`, `.github/workflows/ci.yml`, `internal/infrastructure/config/config.go`, `docs/architecture/SAD.md`, `README.md`).

---

## 1. Topologia Referenciada Hoje

| Camada | Configurado / Documentado no AS-IS | Observação Crítica |
| :--- | :--- | :--- |
| **Desenvolvimento Local** | Docker Compose (`postgres:16-alpine`, `redis:7-alpine`, `adminer`, `api`) | Redis sobe no Compose, mas a API Go nunca abre conexão com Redis. |
| **Backend Alvo (Docs)** | **Google Cloud Run** (`SAD.md`, `README.md`, commit `4a28ac9`) | Não há Terraform/workflow de deploy GCP no repo, apenas documentação e leitura de `PORT`. |
| **Backend Legado (Comentários)** | **Railway** (`docker-compose.yml:4`, `config_test.go:5`) | Inconsistência documental com `SAD.md`. |
| **Banco Produção (Docs)** | Supabase PostgreSQL (`DATABASE_URL`) | Portável via SQL padrão (`pgx/v5`), porém sem migrations executadas nem pipeline de backup. |
| **Storage** | Cloudflare R2 (`R2_*` obrigatórios em `config.go`) | Sem abstração para ambiente sem R2; erro fatal se qualquer uma das 5 variáveis faltar. |
| **Frontend Deploy** | Cloudflare Pages / Vercel (`SAD.md`) | Sem build Flutter no CI (`ci.yml`), sem configuração de headers/CSP nem proxy mesmo-site. |
| **Cache / Rate Limit** | Upstash Redis (`REDIS_URL` obrigatório em `config.go`) | **Não implementado em nenhum pacote Go**; bloqueia boot desnecessariamente. |
| **CI/CD** | GitHub Actions apenas para Go | Faltam jobs Flutter, Gitleaks, Semgrep, OSV-Scanner, Trivy, SBOM e permissões mínimas. |
