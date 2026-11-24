# Diagramas de Sequência e Swimlanes (Fluxos Críticos TO-BE)

## 1. Fluxo de Convite e Ativação do Aluno (Fase 3 — Sem CPF)

```mermaid
sequenceDiagram
    autonumber
    actor Trainer as Personal Trainer
    participant AppT as Flutter Web (Trainer)
    participant API as Fit Brother API (Go)
    participant DB as PostgreSQL
    participant Mail as SMTP (Resend/Brevo)
    actor Student as Aluno

    Trainer->>AppT: Informa nome e e-mail do aluno
    AppT->>API: POST /v1/invitations (Sessão Trainer + Org)
    API->>API: Gera token aleatório (256-bit) e calcula SHA-256(token)
    API->>DB: INSERT invitations (token_hash, expires_at=72h, PENDING)
    API->>Mail: Envia link https://app.../activate?token=...
    API-->>AppT: 201 Created (sem expor token em log)

    Student->>AppT: Abre link de ativação, cria senha e aceita Termos/LGPD
    AppT->>API: POST /v1/auth/activate {token, password, consent_health_data}
    API->>DB: BEGIN TX
    API->>DB: UPDATE invitations SET status='ACCEPTED' WHERE token_hash=? AND status='PENDING' AND expires_at>NOW()
    alt Convite inválido/expirado/usado
        API-->>AppT: 400/410 Convite inválido ou expirado
    else Convite válido
        API->>DB: INSERT users (STUDENT) + students (sem CPF) + trainer_student_relationships + consent_records
        API->>DB: COMMIT TX
        API-->>AppT: 200 OK (Sessão iniciada com cookies HttpOnly)
    end
```

---

## 2. Upload e Playback Autorizado de Vídeo (Fases 1–4 — Anti-BOLA)

```mermaid
sequenceDiagram
    autonumber
    actor User as Usuário (Trainer ou Aluno)
    participant App as Flutter App
    participant API as Fit Brother API
    participant DB as PostgreSQL
    participant R2 as StoragePort (R2 / Local)

    rect rgb(240, 245, 255)
        note right of User: Fluxo A — Upload pelo Trainer
        User->>App: Seleciona vídeo (MP4/WebM <= 50MB) para exercício X
        App->>API: POST /v1/videos/upload-url {exercise_id, mime_type, size_bytes}
        API->>DB: Verifica posse do exercício X pela organização/trainer
        API->>API: Gera storage_key controlado (org_id/exercise_id/uuid.ext)
        API->>DB: INSERT exercise_media (status='PENDING')
        API->>R2: Gera Presigned PutObject (TTL 5m, ContentType/Length)
        API-->>App: 200 {media_id, upload_url, expires_in}
        App->>R2: PUT upload_url (binário direto)
        App->>API: POST /v1/videos {media_id}
        API->>R2: HeadObject(storage_key) — valida existência, tamanho e MIME real
        API->>DB: UPDATE exercise_media SET status='CONFIRMED', size_bytes=...
        API-->>App: 201 Created
    end

    rect rgb(240, 255, 245)
        note right of User: Fluxo B — Playback pelo Aluno ou Trainer
        User->>App: Toca em "Ver demonstração" do exercício X
        App->>API: GET /v1/videos/exercises/:exercise_id/signed-url
        API->>DB: Verifica se trainer é dono da org OU aluno possui treino publicado/ativo com exercício X
        alt Sem vínculo autorizado
            API-->>App: 403 Forbidden
        else Autorizado
            API->>R2: Gera Presigned GetObject (TTL 5m)
            API-->>App: 200 {signed_url}
        end
    end
```

---

## 3. Execução de Treino pelo Aluno com Idempotência (Fase 6)

```mermaid
sequenceDiagram
    autonumber
    actor Aluno
    participant PWA as Flutter Mobile/PWA
    participant API as Fit Brother API
    participant DB as PostgreSQL

    Aluno->>PWA: Inicia "Treino A" e registra séries/cargas (usa cronômetro local)
    PWA->>PWA: Gera client_idempotency_key (UUID) por sessão/lote
    Aluno->>PWA: Clica em "Concluir Treino" (ou reconecta após oscilação de rede)
    PWA->>API: POST /v1/student/executions {assignment_id, idempotency_key, sets[], feedback}
    API->>DB: Verifica assignment pertence ao student_id do usuário logado
    API->>DB: INSERT ... ON CONFLICT (client_idempotency_key) DO NOTHING (em TX)
    alt Já processado anteriormente
        API-->>PWA: 200 OK (retorna execução existente sem duplicar)
    else Novo registro
        API->>DB: Insere set_logs + workout_feedback + COMMIT
        API-->>PWA: 201 Created
    end
```
