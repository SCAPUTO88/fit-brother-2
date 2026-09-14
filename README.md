# Fit Brother

Fit Brother is a personal fitness platform for managing students, workout
plans, sessions, exercises, and demonstration videos. The product direction is
Flutter Web/PWA first, Android/iOS later, with a Go backend and a free-tier
production stack.

## Architecture

The backend follows Hexagonal Architecture (Ports & Adapters):

```text
cmd/api                 Go API entrypoint
internal/domain         Entities and repository/storage ports
internal/application    Use cases and DTOs
internal/adapters       HTTP handlers, Postgres repositories, R2 storage
internal/infrastructure Config, crypto, JWT, logger
migrations              SQL migrations for PostgreSQL
fit_brother_app         Flutter Web/PWA and future mobile app
docs                    Architecture and execution plans
```

See [docs/architecture/SAD.md](docs/architecture/SAD.md) for the full system
architecture and [docs/plano-de-ataque.md](docs/plano-de-ataque.md) for the
implementation roadmap.

## Stack

| Layer | Technology |
| --- | --- |
| Backend | Go 1.26.6 in CI/Docker, module Go 1.25.0, Gin, sqlx, pgx/v5 |
| Database | Supabase PostgreSQL in production; PostgreSQL 16 locally |
| Cache | Upstash Redis in production; Redis 7 locally |
| Video storage | Cloudflare R2 with signed URLs |
| CDN/DNS | Cloudflare Free |
| Backend deploy | Google Cloud Run |
| App deploy | Cloudflare Pages recommended; Vercel acceptable |
| App | Flutter 3.41.x, Dart 3.11.4, Riverpod, GoRouter, Dio |
| CI/CD | GitHub Actions |

## Current Capabilities

- JWT login, refresh, and logout flow.
- Student creation, listing, detail lookup, and anonymization.
- Workout plans, sessions, and exercises.
- Cloudflare R2 signed URL flow for video upload and playback.
- Flutter client with authentication, students, workouts, and plan detail flows.
- SQL migrations for users, refresh tokens, students, workouts, and videos.

## Local Setup

### Requirements

- Go 1.26.6 recommended for local parity with CI and Docker
- Flutter 3.41.x / Dart 3.11.x
- Docker Desktop
- Git

### Environment

Create a local env file:

```bash
cp .env.example .env
```

Generate local JWT keys:

```bash
openssl genrsa -out secrets/jwt_rsa.pem 2048
openssl rsa -in secrets/jwt_rsa.pem -pubout -out secrets/jwt_rsa.pub.pem
```

Generate crypto secrets:

```bash
openssl rand -base64 32   # AES_KEY
openssl rand -base64 32   # CPF_HMAC_SECRET
openssl rand -base64 32   # PASSWORD_PEPPER
```

### Start Local Services

```bash
docker compose -f docker/docker-compose.yml up -d
```

Services:

- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- Adminer: `http://localhost:8081`
- API: `http://localhost:8080`

### Run Migrations

```bash
go run ./scripts/migrate.go up
```

Rollback all migrations:

```bash
go run ./scripts/migrate.go down
```

### Run Backend Outside Docker

The Compose command already starts the API. For local debugging without the
API container:

```bash
go run ./cmd/api
```

Health check:

```bash
curl http://localhost:8080/v1/health
```

### Run Flutter Web

```bash
cd fit_brother_app
flutter pub get
flutter run -d chrome
```

## API Overview

Base path: `/v1`

| Method | Route | Purpose |
| --- | --- | --- |
| GET | `/health` | API health check |
| POST | `/auth/login` | Login |
| POST | `/auth/refresh` | Refresh access token |
| POST | `/auth/logout` | Logout |
| POST | `/students` | Create student |
| GET | `/students` | List students |
| GET | `/students/:id` | Get student detail |
| DELETE | `/students/:id` | Anonymize student |
| POST | `/workouts` | Create workout plan |
| GET | `/workouts` | List plans for current user |
| GET | `/workouts/:id` | Get plan details |
| POST | `/workouts/:id/sessions` | Create workout session |
| POST | `/workouts/sessions/:session_id/exercises` | Create exercise |
| POST | `/videos/upload-url` | Create signed upload URL |
| POST | `/videos` | Confirm uploaded video metadata |
| GET | `/videos/exercises/:exercise_id/signed-url` | Create signed playback URL |

## Testing and Checks

Backend:

```bash
go test -v ./...
go test -v -tags integration ./tests/integration/...
go build ./cmd/api
```

Flutter:

```bash
cd fit_brother_app
flutter analyze
flutter test
```

CI additionally runs `golangci-lint`, `gofmt -l .`, `govulncheck`, race-enabled
Go tests, Linux binary build, and Docker image build. Cloud Run deployment is
configured separately from the current workflow.

## Security Notes

- CPF is encrypted with AES-256-GCM and searchable through HMAC-SHA256.
- Passwords use bcrypt plus `PASSWORD_PEPPER`.
- JWT uses RS256 with local key files and provider-managed secrets in production.
- Videos are private in R2 and accessed only through short-lived signed URLs.
- Do not commit `.env`, JWT keys, peppers, R2 credentials, Supabase credentials,
  Upstash credentials, or production secrets.

## Roadmap

The current roadmap is tracked in [docs/plano-de-ataque.md](docs/plano-de-ataque.md):

1. Baseline and security hygiene.
2. Supabase production database and migrations.
3. Google Cloud Run backend deploy.
4. Upstash Redis rate limiting and short cache.
5. API contract and authorization tightening.
6. R2 video flow hardening.
7. Flutter Web/PWA visual refresh.
8. Frontend deploy.
9. Operations and observability.
10. Mobile app packaging.
