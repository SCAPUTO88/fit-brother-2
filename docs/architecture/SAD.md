# Software Architecture Document (SAD)

## Fit Brother - Personal Fitness Platform

**Versao:** 2.0  
**Data:** 2026-09-13  
**Status:** Arquitetura alvo para MVP pessoal em free-tier

---

## 1. Objetivo

Fit Brother e uma plataforma pessoal para gestao de alunos, planos de treino,
sessoes, exercicios e videos demonstrativos. O produto deve comecar como
Flutter Web/PWA, evoluir para app Android/iOS com a mesma base Flutter e manter
backend proprio em Go.

Os objetivos arquiteturais sao:

- operar com custo zero ou proximo de zero usando free-tier;
- preservar dados sensiveis de alunos conforme LGPD;
- manter backend testavel via arquitetura hexagonal;
- servir videos com URL assinada e TTL curto;
- permitir evolucao visual inspirada em apps fitness modernos, sem copiar assets
  de terceiros.

---

## 2. Stack Alvo

| Camada | Decisao | Justificativa |
| --- | --- | --- |
| App | Flutter Web/PWA primeiro; Android/iOS depois | Reuso de UI, rotas e estado; ja existe no repositorio |
| Backend | Go 1.26.6 em CI/Docker, modulo Go 1.25.0 + Gin | Baixo custo, binario simples, codigo atual |
| Banco | Supabase PostgreSQL | PostgreSQL gerenciado free-tier usando `DATABASE_URL` |
| Banco local | PostgreSQL 16 via Docker Compose | Replica local simples da Supabase |
| Cache | Upstash Redis | Rate limit, sessoes auxiliares e cache curto |
| Storage | Cloudflare R2 | Storage S3-compatible, bom para videos |
| CDN/DNS | Cloudflare Free | Dominio, cache, protecao basica e proximidade com R2 |
| Backend deploy | Railway | Ja existe workflow de deploy no CI |
| App deploy | Cloudflare Pages recomendado; Vercel aceitavel | Free-tier e bom encaixe com Cloudflare |
| CI/CD | GitHub Actions | Lint, security scan, testes, build e deploy |

## 3. Modulos do Repositorio

```text
cmd/api                 API Go
internal/domain         Entidades e ports
internal/application    Use cases e DTOs
internal/adapters       HTTP, Postgres e R2
internal/infrastructure Config, crypto, JWT e logger
migrations              SQL versionado por golang-migrate
docker                  Dockerfile e Compose local
fit_brother_app         Flutter Web/mobile
docs                    Arquitetura e planos
```

O backend segue arquitetura hexagonal. O dominio nao depende de Gin, sqlx,
Supabase, Redis, R2 ou Railway. Integracoes externas ficam em adapters ou
infrastructure.

```mermaid
flowchart LR
    App[Flutter Web/PWA e Mobile] --> API[Go API /v1]
    API --> UseCases[Application Use Cases]
    UseCases --> Domain[Domain: User, Student, Workout, Video]
    UseCases --> Ports[Repository/Storage Ports]
    Ports --> PG[(Supabase Postgres)]
    Ports --> R2[(Cloudflare R2)]
    API --> Redis[(Upstash Redis)]
    Cloudflare[Cloudflare DNS/CDN] --> App
    Cloudflare --> R2
    GitHub[GitHub Actions] --> Railway[Railway Backend]
    GitHub --> Pages[Cloudflare Pages/Vercel App]
```

---

## 4. Requisitos Funcionais Cobertos

| RF | Requisito | Componentes |
| --- | --- | --- |
| RF-01 | Login, refresh e logout | `auth`, JWT RS256, refresh tokens |
| RF-02 | Cadastro e listagem de alunos | `student`, Postgres |
| RF-03 | Consulta e anonimizacao LGPD de aluno | `student`, crypto, Postgres |
| RF-04 | Criacao e consulta de planos de treino | `workout`, Postgres |
| RF-05 | Criacao de sessoes e exercicios | `workout`, Postgres |
| RF-06 | Upload de video demonstrativo | `video`, R2 signed upload URL |
| RF-07 | Playback seguro de video | `video`, R2 signed read URL |
| RF-08 | App responsivo com experiencia fitness | Flutter, Riverpod, GoRouter, Dio |
| RF-09 | Deploy free-tier | Railway, Supabase, Upstash, Cloudflare |

---

## 5. Persistencia e Migrations

O banco de producao deve ser Supabase PostgreSQL. O backend usa `DATABASE_URL`,
`pgx/v5`, `sqlx` e migrations SQL, entao nao depende de APIs proprietarias da
Supabase. O Compose local sobe PostgreSQL 16 e Redis 7 para desenvolvimento.

Migrations atuais:

- `001_create_users`
- `002_create_refresh_tokens`
- `003_create_students_and_biometrics`
- `004_create_workouts`
- `005_create_videos`

Comando de migracao:

```bash
go run ./scripts/migrate.go up
go run ./scripts/migrate.go down
```

### Modelo de Dados Atual

```mermaid
erDiagram
    USERS ||--o| STUDENTS : owns
    USERS ||--o{ REFRESH_TOKENS : has
    USERS ||--o{ WORKOUT_PLANS : trainer
    STUDENTS ||--o{ BIOMETRIC_DATA : has
    STUDENTS ||--o{ WORKOUT_PLANS : receives
    WORKOUT_PLANS ||--o{ WORKOUT_SESSIONS : contains
    WORKOUT_SESSIONS ||--o{ EXERCISES : contains
    EXERCISES ||--o| VIDEOS : demonstrates

    USERS {
        uuid id PK
        text email
        text full_name
        text password_hash
        text role
        bool is_active
        timestamptz created_at
        timestamptz updated_at
    }

    STUDENTS {
        uuid id PK
        uuid user_id FK
        text cpf_encrypted
        text cpf_hash
        timestamptz created_at
    }

    WORKOUT_PLANS {
        uuid id PK
        uuid trainer_id FK
        uuid student_id FK
        text title
        text description
        bool is_active
    }

    VIDEOS {
        uuid id PK
        uuid exercise_id FK
        uuid uploader_id FK
        text r2_key
        text filename
        bigint size_bytes
        text mime_type
    }
```

---

## 6. API HTTP

Base local: `http://localhost:8080/v1`

| Metodo | Rota | Acesso | Uso |
| --- | --- | --- | --- |
| GET | `/health` | Publico | Health check |
| POST | `/auth/login` | Publico | Login |
| POST | `/auth/refresh` | Publico | Refresh token |
| POST | `/auth/logout` | Publico hoje; deve exigir token/RT valido | Logout |
| POST | `/students` | Publico hoje; alvo TRAINER | Criar aluno |
| GET | `/students` | Autenticado | Listar alunos |
| GET | `/students/:id` | Autenticado | Detalhar aluno |
| DELETE | `/students/:id` | Autenticado | Anonimizacao |
| POST | `/workouts` | TRAINER | Criar plano |
| GET | `/workouts` | TRAINER/STUDENT | Listar planos por papel |
| GET | `/workouts/:id` | TRAINER/STUDENT dono | Detalhar plano |
| POST | `/workouts/:id/sessions` | TRAINER | Criar sessao |
| POST | `/workouts/sessions/:session_id/exercises` | TRAINER | Criar exercicio |
| POST | `/videos/upload-url` | TRAINER | Gerar URL assinada upload |
| POST | `/videos` | TRAINER | Confirmar upload |
| GET | `/videos/exercises/:exercise_id/signed-url` | Autenticado | Gerar URL assinada leitura |

Divida a evolucao de API em duas etapas:

1. estabilizar e documentar o contrato atual;
2. restringir `POST /students` e `POST /auth/logout` quando o fluxo do app ja
   tiver token e papel de treinador funcionando sem atrito.

---

## 7. Sequencias Principais

### 7.1 Login e Emissao de Tokens

```mermaid
sequenceDiagram
    participant App as Flutter App
    participant API as Go API
    participant Auth as Auth UseCase
    participant DB as Supabase Postgres

    App->>API: POST /v1/auth/login
    API->>Auth: Execute(email, password, client_ip)
    Auth->>DB: Get user by email
    DB-->>Auth: user + password_hash
    Auth->>Auth: bcrypt+pepper check
    Auth->>Auth: sign JWT RS256
    Auth->>DB: Save refresh token hash + family_id
    Auth-->>API: access_token, refresh_token, role
    API-->>App: 200 OK
```

### 7.2 Criacao de Aluno

```mermaid
sequenceDiagram
    participant Trainer as Trainer UI
    participant API as Go API
    participant Student as Student UseCase
    participant Crypto as Crypto Services
    participant DB as Supabase Postgres

    Trainer->>API: POST /v1/students
    API->>Student: CreateStudentInput
    Student->>DB: Check email exists
    Student->>Crypto: Clean CPF, HMAC hash, AES-GCM encrypt
    Student->>Crypto: bcrypt password + pepper
    Student->>DB: Insert users
    Student->>DB: Insert students
    Student-->>API: StudentOutput without CPF
    API-->>Trainer: 200 OK
```

### 7.3 Criacao de Plano, Sessao e Exercicio

```mermaid
sequenceDiagram
    participant Trainer as Trainer UI
    participant API as Go API
    participant JWT as Auth Middleware
    participant Workout as Workout UseCases
    participant DB as Supabase Postgres

    Trainer->>API: POST /v1/workouts
    API->>JWT: Validate Bearer token
    JWT-->>API: role TRAINER + user_id
    API->>Workout: CreatePlan(trainer_id, student_id, title)
    Workout->>DB: Insert workout_plans
    DB-->>Workout: ok
    Workout-->>API: PlanOutput
    API-->>Trainer: 201 Created

    Trainer->>API: POST /v1/workouts/:id/sessions
    API->>Workout: CreateSession(plan_id, title, day)
    Workout->>DB: Insert workout_sessions
    API-->>Trainer: 201 Created

    Trainer->>API: POST /v1/workouts/sessions/:id/exercises
    API->>Workout: CreateExercise(session_id, sets, reps)
    Workout->>DB: Insert exercises
    API-->>Trainer: 201 Created
```

### 7.4 Upload de Video

```mermaid
sequenceDiagram
    participant Trainer as Trainer UI
    participant API as Go API
    participant Video as Video UseCase
    participant R2 as Cloudflare R2
    participant DB as Supabase Postgres

    Trainer->>API: POST /v1/videos/upload-url
    API->>Video: RequestUploadURL(exercise_id, filename, mime)
    Video->>R2: Presign PUT with TTL
    R2-->>Video: presigned_url, r2_key
    API-->>Trainer: presigned_url, r2_key
    Trainer->>R2: PUT video bytes
    Trainer->>API: POST /v1/videos confirm metadata
    API->>DB: Insert videos record
    API-->>Trainer: 201 Created
```

### 7.5 Playback de Video

```mermaid
sequenceDiagram
    participant Student as Student UI
    participant API as Go API
    participant JWT as Auth Middleware
    participant Video as Video UseCase
    participant DB as Supabase Postgres
    participant R2 as Cloudflare R2

    Student->>API: GET /v1/videos/exercises/:id/signed-url
    API->>JWT: Validate Bearer token
    API->>Video: GetSignedURL(exercise_id)
    Video->>DB: Find video by exercise_id
    Video->>R2: Presign GET with short TTL
    API-->>Student: signed_url, expires_in_seconds
    Student->>R2: GET video using signed_url
```

---

## 8. Cache e Redis

Redis deve ser introduzido de forma incremental. Usos recomendados:

1. rate limit por IP/email em login e refresh;
2. bloqueio temporario de familias de refresh token comprometidas;
3. cache curto para listas nao sensiveis, como planos do usuario logado;
4. invalidacao por write: criar/editar plano, sessao, exercicio ou aluno.

Nao cachear CPF descriptografado, access tokens ou payloads com dados sensiveis.

---

## 9. Seguranca e LGPD

- CPF: AES-256-GCM com AAD baseado no `student_id`.
- Busca de CPF: HMAC-SHA256 com `CPF_HMAC_SECRET`.
- Senhas: bcrypt com `PASSWORD_PEPPER`.
- JWT: RS256 com chaves em `secrets` localmente e secrets no provider.
- Videos: R2 keys privadas; acesso via signed URLs com TTL curto.
- Anonimizacao: `DELETE /students/:id` deve remover dados identificaveis.
- Logs: nunca registrar CPF, senha, refresh token ou URL assinada completa.

Debito tecnico atual: remover logs verbosos de falha de login que exponham
hash de senha antes de qualquer ambiente compartilhado ou deploy publico.

---

## 10. Deploy e Ambientes

### Desenvolvimento local

- API: `go run ./cmd/api`
- Banco/Redis: `docker/docker-compose.yml`
- Migrations: `go run ./scripts/migrate.go up`
- App: `cd fit_brother_app && flutter run -d chrome`

### Producao alvo

```mermaid
flowchart TB
    Dev[GitHub main] --> CI[GitHub Actions]
    CI --> Tests[Go tests, lint, govulncheck, Docker build]
    Tests --> Railway[Railway Go API]
    Tests --> Pages[Cloudflare Pages Flutter Web]
    Railway --> Supabase[(Supabase Postgres)]
    Railway --> Upstash[(Upstash Redis)]
    Railway --> R2[(Cloudflare R2)]
    Pages --> Railway
    Cloudflare[Cloudflare DNS/CDN] --> Pages
    Cloudflare --> R2
```

Railway continua sendo o deploy do backend porque o workflow ja contem
`railway up --service fit-brother-api`. Cloudflare Pages e recomendado para o
Flutter Web por consolidar DNS/CDN/R2; Vercel segue viavel se for mais rapido
para o primeiro publish.

---

## 11. Direcao Visual do App

A referencia Dribbble escolhida aponta para:

- tema escuro, alto contraste e verde de acao;
- cards compactos para metricas, treinos e progresso;
- linguagem mobile-first;
- tracking de treino, metas e saude;
- espaco futuro para gamificacao e recomendacoes.

O app nao deve copiar layout, assets ou marca de terceiros. A direcao visual
sera reinterpretada em um design system proprio:

- cores: preto/carvao, cinza frio, verde energetico, branco;
- componentes: cards de treino, chips de status, barras de progresso, video
  preview, bottom navigation;
- telas prioritarias: login, dashboard, alunos, planos, detalhe do plano,
  criacao de sessao/exercicio e player de video.

---

## 12. Observabilidade

MVP:

- logs estruturados no backend com zerolog;
- request id por requisicao;
- logs de erro sem dados sensiveis;
- health check `/v1/health`;
- metricas simples no provider de deploy.

Evolucao:

- dashboard de erros por endpoint;
- alerta de taxa alta de login falho;
- alerta de erro em signed URLs;
- tracking de tamanho total no R2.

---

## 13. Riscos e Mitigacoes

| Risco | Impacto | Mitigacao |
| --- | --- | --- |
| Free-tier suspender ou hibernar servicos | App lento ou indisponivel | Aceitar para uso pessoal; documentar limites |
| Flutter Web gerar bundle grande | UX inicial ruim | Lazy loading, assets otimizados, PWA cache |
| URL assinada vazar | Acesso temporario ao video | TTL curto e R2 key privada |
| Cache servir dado sensivel | LGPD | Nao cachear CPF, tokens ou URLs completas |
| Contrato API mudar sem app acompanhar | Regressao | Testes de handler/usecase e plano de API |

---

## 14. Criterios de Aceite da Arquitetura

- Supabase e a base oficial de producao; Postgres local replica o contrato.
- Railway faz deploy do backend com Dockerfile atual.
- Flutter Web/PWA e o primeiro cliente publicado.
- R2 armazena videos; backend nunca faz proxy de arquivo grande.
- Redis e usado primeiro para rate limit, depois cache curto.
- Diagramas deste SAD representam fluxos implementaveis no codigo atual.
- Nenhuma decisao exige servico pago para o MVP pessoal.
