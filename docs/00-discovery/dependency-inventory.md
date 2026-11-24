# Inventário de Dependências Externas e Serviços

Este inventário cobre pacotes de código (`go.mod`, `pubspec.yaml`), serviços externos referenciados na infraestrutura e ferramentas de CI/CD.

---

## 1. Dependências Backend Go (`go.mod` — `go 1.25.0`)

| Pacote | Versão Atual | Papel no Sistema | Risco / Observação |
| :--- | :--- | :--- | :--- |
| `github.com/gin-gonic/gin` | `v1.12.0` | Framework HTTP `/v1` | Ativo, estável; manter e revisar middleware de erros/trusted proxies. |
| `github.com/gin-contrib/cors` | `v1.7.7` | Middleware CORS | `config.go` converte origens; precisa endurecimento por ambiente e suporte eventual a `AllowCredentials` se usar cookies. |
| `github.com/golang-jwt/jwt/v5` | `v5.3.1` | Assinatura e validação JWT RS256 | Não utiliza `kid` no header (dificulta rotação de chaves) nem `aud` explícito. |
| `github.com/jackc/pgx/v5` | `v5.9.2` | Driver PostgreSQL | Ativo e moderno; suporta pool `pgxpool` e prepared statements. |
| `github.com/jmoiron/sqlx` | `v1.4.0` | Acesso SQL e scan de structs | Funciona bem; requer atenção ao uso de transações (`BeginTxx`) onde hoje falta. |
| `github.com/golang-migrate/migrate/v4` | `v4.19.1` | Migrations SQL (via `scripts/migrate.go`) | Está como `// indirect` no `go.mod` pois só aparece em `scripts/migrate.go` (`//go:build ignore`). |
| `github.com/aws/aws-sdk-go-v2` (+ `config`, `credentials`, `service/s3`) | `v1.41.5` / `s3 v1.99.0` | Cliente S3 compatível com Cloudflare R2 | Deve ficar atrás de `StoragePort` (`domain`) para desacoplar da camada de aplicação. |
| `github.com/google/uuid` | `v1.6.0` | Geração/parse de UUIDs | Estável. |
| `github.com/joho/godotenv` | `v1.5.1` | Carga de `.env` local | Apenas dev; ignorado se `.env` ausente. |
| `github.com/rs/zerolog` | `v1.35.0` | Logging estruturado JSON | Bom desempenho; aplicar redação rigorosa de PII/tokens. |
| `golang.org/x/crypto` | `v0.51.0` | `bcrypt` para hashing de senhas | Atualizado; avaliar Argon2id ou manter bcrypt+pepper em novas senhas/rehash. |
| `github.com/DATA-DOG/go-sqlmock` | `v1.5.2` | Mocks SQL em testes | Útil, mas não substitui testes reais com PostgreSQL/Testcontainers. |
| `github.com/stretchr/testify` | `v1.11.1` | Assertions e mocks de teste | Estável. |

---

## 2. Dependências Frontend Flutter (`fit_brother_app/pubspec.yaml` — Dart `^3.11.4`)

| Pacote | Versão Atual | Papel no Sistema | Risco / Observação |
| :--- | :--- | :--- | :--- |
| `flutter_riverpod` | `^3.3.1` | Gerência de estado | Usado sem codegen efetivo apesar de `riverpod_generator` estar em `dev_dependencies`. |
| `go_router` | `^17.2.1` | Roteamento declarativo | Sem guards de autenticação ou redirecionamento por perfil (`TRAINER`/`STUDENT`). |
| `dio` | `^5.9.2` | Cliente HTTP REST | Interceptor `onError` para `401` é apenas um stub sem refresh automático. |
| `flutter_secure_storage` | `^10.0.0` | Armazenamento de `access_token` e `refresh_token` | **Risco Web:** No navegador usa WebCrypto + `localStorage`, acessível via XSS (ver [ADR-008](../adr/ADR-008-web-session-security-bff-cookies.md)). |
| `file_picker` | `^11.0.2` | Seleção de arquivos de vídeo | Usado em `create_exercise_screen.dart`; requer validação de tipo e tamanho antes de upload. |
| `cupertino_icons` | `^1.0.8` | Ícones iOS | Padrão do template Flutter. |
| `flutter_lints` / `build_runner` / `riverpod_generator` | `^6.0.0` / `^2.13.1` / `^4.0.3` | Qualidade e codegen (`dev_dependencies`) | `riverpod_generator` e `build_runner` não são usados nos fontes atuais. |

---

## 3. Serviços Externos Referenciados no Código, Config e Documentação

| Serviço / Provedor | Onde Aparece | Uso Real no Código | Veredito na Retomada |
| :--- | :--- | :--- | :--- |
| **PostgreSQL / Supabase Postgres** | `config.go`, `migrations/`, `docker-compose.yml`, `SAD.md` | **Em uso real** via `sqlx` + `pgx/v5` | **KEEP / EVOLVE** — Manter PostgreSQL padrão e usar Supabase Postgres gerenciado em homologação/piloto. |
| **Cloudflare R2** | `r2_storage.go`, `config.go:93-97`, `SAD.md:38`, `create_exercise_screen.dart:92` | **Em uso parcial** (backend gera presigned URLs; Flutter quebra no contrato) | **REFACTOR** — Desacoplar atrás de `StoragePort` e tornar **opcional** no startup (local/filesystem ou R2/Supabase Storage). |
| **Redis / Upstash** | `config.go:77` (`REDIS_URL`), `docker-compose.yml:71`, `ci.yml:85`, `SAD.md:37` | **Não utilizado** em nenhuma linha de código funcional | **REMOVE (como obrigatório)** — Eliminar `REDIS_URL` obrigatório e dependência em runtime até existir necessidade real (ver [ADR-004](../adr/ADR-004-redis-removal-and-rate-limiting.md)). |
| **Google Cloud Run (GCP)** | `SAD.md:40,73,377`, `README.md:36,172`, `cmd/api` commit `4a28ac9` | Apenas ajuste de leitura de `PORT` e documentos | **REPLACE / EXIT GCP** — Migrar alvo de hospedagem para fora do GCP (Render/Fly/Koyeb — ver [ADR-007](../adr/ADR-007-target-infrastructure-gcp-exit.md)). |
| **Railway** | `docker/docker-compose.yml:4`, `config_test.go:5` | Referência residual de comentário e teste de `PORT` | **REMOVE / EVALUATE** — Manter resolução padrão de `PORT`, removendo acoplamentos semânticos ao Railway (ou reavaliar como alternativa paga barata). |
| **Cloudflare Pages / Vercel / CDN** | `SAD.md:41,71`, `README.md:37` | Planejado para deploy do Flutter Web | **KEEP (Cloudflare Pages)** — Free-tier com banda estática ilimitada (sujeito a revisão se usar BFF no mesmo domínio). |
| **GitHub Actions** | `.github/workflows/ci.yml` | Em uso apenas para Go (lint, `govulncheck`, unit tests, docker build) | **UPDATE** — Fixar versões, adicionar Flutter, Gitleaks, Semgrep, OSV-Scanner, Trivy, SBOM e permissões mínimas. |
| **SMTP (Resend / Brevo)** | Ausente hoje | Necessário a partir da Fase 3 (convites, ativação, reset de senha) | **ADD (Fase 3 / 9)** — Adotar provedor gratuito transacional com baixa fricção (Resend ou Brevo). |
