# Diagramas de Componentes (C4 Nível 3 — Backend Go e Frontend Flutter)

## 1. Backend Go — Arquitetura Hexagonal Alvo

```mermaid
flowchart TB
    subgraph AdaptersIn["Adapters de Entrada (HTTP / CLI)"]
        Router["Gin Router + RateLimiter + SecurityHeaders + CSRF"]
        AuthMW["AuthMiddleware (Sessão/JWT + Check is_active)"]
        OwnershipMW["Ownership / Relationship Guard"]
        Handlers["Handlers: Auth, Invite, Student, Catalog, Workout, Execution, Assessment, Video, Privacy"]
        Router --> AuthMW --> OwnershipMW --> Handlers
    end

    subgraph Application["Application (Use Cases & Transações)"]
        AuthUC["Auth & Session UC (Login, Refresh Tx, Logout, Reset, MFA)"]
        OrgUC["Org & Relationship UC (Invite, Accept, Revoke)"]
        CatalogUC["Exercise Catalog UC (Search, Custom, Import)"]
        WorkoutUC["Workout Builder UC (Templates, Clone, Publish Version)"]
        ExecUC["Execution UC (Start Session, Idempotent SetLog, Feedback)"]
        AssessUC["Assessment & Metrics UC (Measurements, Progress Charts, Dashboard)"]
        VideoUC["Video UC (Via StoragePort + HeadObject Verification)"]
        PrivacyUC["LGPD UC (Consent, Export, Anonymize Tx, Audit)"]
    end

    subgraph Domain["Domain (Puro — Zero Deps Externas)"]
        Entities["Entities & Rules"]
        Ports["Ports: Repos, TxManager, StoragePort, MailerPort, AuditPort"]
    end

    subgraph AdaptersOut["Adapters de Saída & Infra"]
        PgRepo["Postgres Repos + TxManager (sqlx/pgx)"]
        StorageAdapters["R2Storage / LocalStorage (StoragePort)"]
        MailerAdapter["Resend / Brevo / ConsoleMailer (MailerPort)"]
        CryptoJWT["Crypto (AES-256-GCM, Bcrypt/Argon2, HMAC) + JWT (kid)"]
    end

    Handlers --> Application
    Application --> Entities
    Application --> Ports
    PgRepo -. implements .-> Ports
    StorageAdapters -. implements .-> Ports
    MailerAdapter -. implements .-> Ports
    Application --> CryptoJWT
```

### Correção Explícita Frente ao Estado Atual
- Hoje `internal/application/video/usecases.go` referencia `*storage.R2Storage` diretamente. Na Fase 2, `video/usecases.go` passa a depender exclusivamente de `domain.StoragePort`, permitindo mock em testes, storage local em desenvolvimento offline e R2/Supabase em nuvem.

---

## 2. Frontend Flutter — Organização Modular Alvo

```mermaid
flowchart LR
    subgraph Core["fit_brother_app/lib/core"]
        Theme["Design System (Grafite/Preto/Branco + Verde-Lima, Dark/Light)"]
        Router["GoRouter (Auth Guard + Role Redirect: Trainer vs Student)"]
        Api["DioClient (Cookie/Credentials + Refresh Queue Mutex + Error Mapper)"]
        Storage["SessionState (Sem PII/Tokens sensíveis persistidos em claro na Web)"]
    end

    subgraph TrainerFeatures["Shell Personal Trainer (Sidebar Desktop/Tablet)"]
        TDash["Dashboard Operacional"]
        TStudents["Alunos & Convites"]
        TCatalog["Biblioteca de Exercícios"]
        TBuilder["Workout Builder & Templates"]
        TAssess["Avaliações & Evolução"]
    end

    subgraph StudentFeatures["Shell Aluno (Bottom Nav Mobile-First)"]
        SHome["Início / Treino do Dia"]
        SWorkouts["Minhas Fichas"]
        SExec["Execução (Séries, Cargas, Cronômetro, Vídeo)"]
        SHist["Histórico & Feedback"]
        SEvol["Minha Evolução"]
        SProf["Perfil & Privacidade LGPD"]
    end

    Core --> TrainerFeatures
    Core --> StudentFeatures
```
