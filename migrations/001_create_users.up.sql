-- migrations/001_create_users.up.sql
-- Habilita UUID nativo do PostgreSQL
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL,
    full_name     VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,   -- bcrypt + pepper, NUNCA texto plano
    role          VARCHAR(20)  NOT NULL CHECK (role IN ('TRAINER', 'STUDENT')),
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ                          -- soft delete (LGPD)
);

-- Índice único apenas para usuários não deletados (permite reusar email após exclusão)
CREATE UNIQUE INDEX idx_users_email_active
    ON users (email)
    WHERE deleted_at IS NULL;

-- índice para buscas por role
CREATE INDEX idx_users_role ON users (role) WHERE deleted_at IS NULL;
