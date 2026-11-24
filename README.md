# Fit Brother

Fit Brother is a personal fitness platform for managing students, workout
plans, sessions, exercises, and demonstration videos. The product direction is
Flutter Web/PWA first, Android/iOS later, with a Go backend hosted outside GCP.

## Architecture

The backend follows Hexagonal Architecture (Ports & Adapters):

```text
cmd/api                 Go API entrypoint
internal/domain         Entities and repository/storage ports
internal/application    Use cases and DTOs
internal/adapters       HTTP handlers, Postgres repositories, Local/R2/Disabled storage
internal/infrastructure Config, crypto, JWT, logger
migrations              SQL migrations for PostgreSQL
fit_brother_app         Flutter Web/PWA and future mobile app
docs                    Architecture, ADRs, and execution plans
```

See [docs/02-architecture/target-architecture.md](docs/02-architecture/target-architecture.md)
for the current target architecture and [docs/07-project/roadmap.md](docs/07-project/roadmap.md)
for the implementation roadmap.

## Stack

| Layer | Technology |
| --- | --- |
| Backend | Go 1.26.11 in CI/Docker, module Go 1.26.0, Gin, sqlx, pgx/v5 |
| Database | Supabase PostgreSQL in production; PostgreSQL 16 locally |
| Rate Limiting | In-memory per-IP limiter (`ADR-004`, no mandatory Redis) |
| Video storage | Plug-in `StoragePort` (`STORAGE_DRIVER=local \| r2 \| disabled`) |
| CDN/DNS | Cloudflare Free |
| Backend deploy | Container PaaS outside GCP (`Render` / `Fly.io`) |
| App deploy | Cloudflare Pages |
| App | Flutter stable / Dart 3.11+, Riverpod, GoRouter, Dio |
| Web Auth | `HttpOnly; SameSite=Lax` cookies + CSRF protection (`ADR-008`) |
| CI/CD | GitHub Actions (`golangci-lint`, `govulncheck`, `Gitleaks`, `Semgrep`, `OSV-Scanner`, `Trivy`, SBOM) |

## Current Capabilities

- JWT RS256 login, refresh rotation with family revocation, and logout (Bearer for mobile + `HttpOnly` cookies & CSRF protection for Web).
- In-memory rate limiting on authentication endpoints (`/v1/auth/login`, `/v1/auth/refresh`).
- Student creation, listing, detail lookup, and LGPD anonymization with trainer ownership isolation.
- Workout plans, sessions, and exercises with RBAC and ownership validation.
- Pluggable video storage (`local` signed disk endpoints for dev/tests, `r2` for production, or `disabled`).
- Flutter client with serialized `401` refresh queue, Web cookie session mode, students, workouts, and video upload flows.
- SQL migrations `001`–`006` for users, refresh tokens, students, workouts, ownership, and video upload tracking.

## Local Setup

### Requirements

- Go 1.26.x
- Flutter stable / Dart 3.11+
- Docker Desktop (for local PostgreSQL 16)
- Git

### Environment

Create a local env file (API starts with only `DATABASE_URL` + Auth/Crypto keys; Redis and R2 are optional):

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
- Adminer: `http://localhost:8081`
- API: `http://localhost:8080`

### Run Migrations

```bash
go run ./scripts/migrate.go up
```

### Run Backend Outside Docker

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
| POST | `/auth/login` | Login (supports `X-Session-Mode: cookie`) |
| POST | `/auth/refresh` | Refresh access token |
| POST | `/auth/logout` | Logout and revoke token family |
| POST | `/students` | Create student (`TRAINER`) |
| GET | `/students` | List trainer's students (`TRAINER`) |
| GET | `/students/:id` | Get student detail |
| DELETE | `/students/:id` | Anonymize student (LGPD) |
| POST | `/workouts` | Create workout plan |
| GET | `/workouts` | List plans for current user |
| GET | `/workouts/:id` | Get plan details |
| POST | `/workouts/:id/sessions` | Create workout session |
| POST | `/workouts/sessions/:session_id/exercises` | Create exercise |
| POST | `/videos/upload-url` | Create signed upload URL |
| POST | `/videos` | Confirm uploaded video metadata |
| GET | `/videos/exercises/:exercise_id/signed-url` | Create signed playback URL |
| PUT / GET | `/storage/local` | Signed local object endpoints (when `STORAGE_DRIVER=local`) |

## Testing and Checks

Backend:

```bash
go test -race ./...
go test -v -tags integration ./tests/integration/...
go build ./cmd/api
```

Flutter:

```bash
cd fit_brother_app
flutter analyze --no-fatal-infos
flutter test
flutter build web --release
```

## Security Notes

- CPF is encrypted with AES-256-GCM and searchable through HMAC-SHA256.
- Passwords use bcrypt plus `PASSWORD_PEPPER`.
- JWT uses RS256 with local key files or provider-managed PEM secrets.
- Web sessions use `HttpOnly; SameSite=Lax` cookies plus `Origin` and `X-CSRF-Token` validation instead of storing tokens in browser `localStorage`.
- Videos use short-lived signed URLs backed by Cloudflare R2 or HMAC-signed local storage.
- Never commit `.env`, JWT keys, peppers, R2 credentials, Supabase credentials, or production secrets.
