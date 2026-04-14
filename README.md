# Fit Brother 🏋️

Backend em **Go** para o sistema de gestão de treinos Fit Brother — seguro, LGPD-compliant e 100% baseado em serviços gratuitos.

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Backend | Go 1.23 + Gin |
| Banco de Dados | PostgreSQL 16 (Supabase free tier) |
| Cache | Redis (Upstash free tier) |
| Storage de Vídeos | Cloudflare R2 (10 GB grátis, egress zero) |
| Hosting | Railway.app (free tier) |
| Admin Web | Next.js + Vercel (free tier) |
| App Mobile | Flutter (Android + iOS) |
| CI/CD | GitHub Actions |

## Arquitetura

Hexagonal Architecture (Ports & Adapters) — domínio completamente desacoplado de frameworks e infraestrutura.

```
internal/
├── domain/          # Entidades e regras de negócio puras
├── application/     # Use Cases (orquestração)
├── adapters/        # HTTP handlers, repositórios, storage
└── infrastructure/  # Config, crypto, JWT, logger
```

Documentação completa: [docs/architecture/SAD.md](docs/architecture/SAD.md)

## Segurança

- **CPF**: criptografado com AES-256-GCM + AAD; hash HMAC-SHA256 para buscas
- **Senhas**: bcrypt (custo 12) + pepper externo ao banco
- **JWT**: RS256, access token 15min + refresh token 7 dias com Token Family Rotation
- **Vídeos**: Signed URLs com TTL de 5 minutos (Cloudflare R2)
- **OWASP Top 10**: mitigações implementadas em todos os endpoints

## Setup Local

### Pré-requisitos

- [Go 1.23+](https://go.dev/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Git](https://git-scm.com/)

### 1. Clonar o repositório

```bash
git clone https://github.com/SCAPUTO88/fit-brother.git
cd fit-brother
```

### 2. Configurar variáveis de ambiente

```bash
cp .env.example .env
# Edite o .env com suas credenciais
```

### 3. Gerar chaves JWT (se ainda não existirem)

```bash
# Linux/Mac ou Git Bash no Windows:
openssl genrsa -out secrets/jwt_rsa.pem 2048
openssl rsa -in secrets/jwt_rsa.pem -pubout -out secrets/jwt_rsa.pub.pem

# Gerar secrets de criptografia:
openssl rand -base64 32   # para AES_KEY
openssl rand -base64 32   # para CPF_HMAC_SECRET
openssl rand -base64 32   # para PASSWORD_PEPPER
```

### 4. Subir infraestrutura local

```bash
cd docker
docker-compose up -d postgres redis adminer
```

- **PostgreSQL**: `localhost:5432`
- **Redis**: `localhost:6379`
- **Adminer (UI do DB)**: http://localhost:8081

### 5. Rodar as migrações

```bash
# Em breve — Sprint 1
```

### 6. Rodar a API

```bash
go run ./cmd/api
```

A API estará disponível em **http://localhost:8080**

### Health check

```bash
curl http://localhost:8080/health
# {"status":"ok","env":"development","version":"1.0.0"}
```

## Testes

```bash
# Testes unitários
go test -v -race ./internal/...

# Testes de integração
go test -v -tags integration ./tests/integration/...

# Cobertura
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out
```

## Endpoints (em construção)

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/health` | Health check |
| — | — | Sprint 1: Auth (login, refresh, logout) |
| — | — | Sprint 2: Alunos + Biometria |
| — | — | Sprint 3: Treinos + Exercícios |
| — | — | Sprint 4: Vídeos (Signed URLs) |

## Conformidade LGPD

- Dados pessoais (CPF) criptografados em repouso
- Soft delete com anonimização para direito ao esquecimento (Art. 18)
- Trilha de auditoria completa (`audit_logs`)
- Consentimento registrado no cadastro

## Roadmap

- [x] Sprint 0 — Fundação (estrutura, config, CI/CD)
- [ ] Sprint 1 — Autenticação JWT
- [ ] Sprint 2 — Módulo de Alunos
- [ ] Sprint 3 — Módulo de Treinos
- [ ] Sprint 4 — Storage de Vídeos
- [ ] Sprint 5 — App Flutter
- [ ] Sprint 6 — Painel Admin
- [ ] Sprint 7 — Launch 🚀

---

Desenvolvido com ❤️ usando Go + Hexagonal Architecture
