# Modelagem Atual de Dados (AS-IS)

Schema extraído integralmente das migrations `001` a `005` em `migrations/`.

---

## 1. Diagrama Entidade-Relacionamento Atual (DER AS-IS)

```mermaid
erDiagram
    users ||--o{ refresh_tokens : "possui"
    users ||--o| students : "perfil aluno (user_id UNIQUE)"
    users ||--o{ workout_plans : "trainer_id (prescreve)"
    users ||--o{ videos : "uploader_id"
    students ||--o{ biometric_data : "possui"
    students ||--o{ workout_plans : "student_id"
    workout_plans ||--o{ workout_sessions : "contem"
    workout_sessions ||--o{ exercises : "contem"
    exercises ||--o{ videos : "possui"

    users {
        uuid id PK
        text email UK
        text full_name
        text password_hash
        text role "TRAINER | STUDENT"
        boolean is_active
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    refresh_tokens {
        uuid id PK
        uuid user_id FK
        text token_hash UK
        timestamptz expires_at
        timestamptz revoked_at
        timestamptz created_at
    }

    students {
        uuid id PK
        uuid user_id FK "UNIQUE"
        text cpf_encrypted "NOT NULL (AES-256-GCM)"
        text cpf_hash UK "NOT NULL (HMAC-SHA256)"
        date birth_date
        text phone
        text emergency_contact
        timestamptz created_at
        timestamptz updated_at
    }

    biometric_data {
        uuid id PK
        uuid student_id FK
        numeric weight_kg
        numeric height_cm
        numeric body_fat_pct
        numeric muscle_mass_kg "ausente na entity Go"
        text notes_encrypted "ausente na entity Go"
        timestamptz measured_at
        timestamptz created_at
    }

    workout_plans {
        uuid id PK
        uuid trainer_id FK "-> users.id"
        uuid student_id FK "-> students.id"
        text title
        text description
        boolean is_active
        timestamptz created_at
        timestamptz updated_at
    }

    workout_sessions {
        uuid id PK
        uuid workout_plan_id FK
        text title
        smallint day_of_week
        text notes
        timestamptz created_at
    }

    exercises {
        uuid id PK
        uuid workout_session_id FK
        text name
        smallint sets
        text reps
        numeric load_kg
        smallint rest_seconds
        text notes
        smallint order_index
        timestamptz created_at
    }

    videos {
        uuid id PK
        uuid exercise_id FK
        uuid uploader_id FK "-> users.id"
        text r2_key UK
        text filename
        bigint size_bytes
        text mime_type
        timestamptz created_at
    }
```

---

## 2. Problemas Identificados nas Migrations Atuais

1. **Ausência de relacionamento Trainer ↔ Student:** `students` não referencia `trainer_id` nem há tabela de vínculo; qualquer trainer acessa qualquer aluno.
2. **Inconsistência `users.id` vs `students.id`:** `workout_plans.student_id` aponta para `students.id`, mas o código atual (`workout_handler.go` e `workout/usecases.go`) passa `claims.UserID` (`users.id`).
3. **CPF obrigatório (`NOT NULL`) em `students`:** Contraria a minimização da LGPD para gestão de treinos e cria superfície crítica de exposição (`cpf_encrypted` retornada por `GET /students/:id`).
4. **Tabela `biometric_data` ociosa e dessincronizada:** As colunas `muscle_mass_kg` e `notes_encrypted` existem no SQL (`003_create_students_and_biometrics.up.sql`), mas não constam em `domain/student/entity.go`, nem há repositórios/endpoints.
5. **Ausência de catálogo e execução:** `exercises` mistura definição do exercício (`name`), prescrição (`sets`, `reps`, `load_kg`) e não possui histórico de execução pelo aluno nem versionamento de fichas.
6. **Risco no `migrate down`:** `scripts/migrate.go down` executa rollback total de todas as tabelas (`DROP TABLE`), incompatível com ambientes persistentes.
