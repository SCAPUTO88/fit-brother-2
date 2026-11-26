-- migrations/007_organizations_invites_and_lgpd.up.sql
-- Fase 3: Multi-tenancy (organizations, memberships, trainer_student_relationships),
-- convites, reset de senha, MFA TOTP e registros LGPD (consent_records, privacy_requests, audit_events)
-- além da descontinuação de CPF obrigatório (ADR-006).

CREATE TABLE IF NOT EXISTS organizations (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    slug       VARCHAR(120) NOT NULL UNIQUE,
    type       VARCHAR(20)  NOT NULL DEFAULT 'SOLO' CHECK (type IN ('SOLO', 'STUDIO')),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS memberships (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL CHECK (role IN ('OWNER', 'TRAINER', 'ADMIN')),
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'SUSPENDED')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_memberships_user_status
    ON memberships (user_id, status);

CREATE TABLE IF NOT EXISTS trainer_student_relationships (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    trainer_user_id UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id      UUID        NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'PAUSED', 'ENDED')),
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at        TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_rel_active_trainer_student
    ON trainer_student_relationships (trainer_user_id, student_id)
    WHERE status = 'ACTIVE';

CREATE INDEX IF NOT EXISTS idx_rel_org_trainer_status
    ON trainer_student_relationships (organization_id, trainer_user_id, status);

CREATE INDEX IF NOT EXISTS idx_rel_student_status
    ON trainer_student_relationships (student_id, status);

CREATE TABLE IF NOT EXISTS invitations (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID         NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    trainer_user_id UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email           VARCHAR(255) NOT NULL,
    full_name       VARCHAR(255) NOT NULL,
    token_hash      VARCHAR(64)  NOT NULL UNIQUE,
    status          VARCHAR(20)  NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ACCEPTED', 'EXPIRED', 'REVOKED')),
    expires_at      TIMESTAMPTZ  NOT NULL,
    accepted_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invitations_trainer_status
    ON invitations (trainer_user_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_invitations_email_status
    ON invitations (email, status);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_user
    ON password_reset_tokens (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS user_mfa (
    user_id          UUID        PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    secret_encrypted TEXT        NOT NULL,
    enabled          BOOLEAN     NOT NULL DEFAULT FALSE,
    last_used_step   BIGINT      NOT NULL DEFAULT 0,
    verified_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS consent_records (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    purpose        VARCHAR(50)  NOT NULL CHECK (purpose IN ('HEALTH_DATA', 'MEDIA_UPLOAD', 'TERMS')),
    policy_version VARCHAR(50)  NOT NULL,
    granted        BOOLEAN      NOT NULL,
    ip_address     INET,
    user_agent     TEXT,
    recorded_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_consent_records_user
    ON consent_records (user_id, purpose, recorded_at DESC);

CREATE TABLE IF NOT EXISTS privacy_requests (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    request_type VARCHAR(30) NOT NULL CHECK (request_type IN ('EXPORT', 'ANONYMIZE', 'RECTIFY')),
    status       VARCHAR(30) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'COMPLETED', 'REJECTED')),
    details      TEXT,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_privacy_requests_user
    ON privacy_requests (user_id, requested_at DESC);

CREATE TABLE IF NOT EXISTS audit_events (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id   UUID        REFERENCES users(id) ON DELETE SET NULL,
    organization_id UUID        REFERENCES organizations(id) ON DELETE SET NULL,
    action          VARCHAR(80) NOT NULL,
    target_type     VARCHAR(50) NOT NULL,
    target_id       UUID,
    metadata        JSONB       NOT NULL DEFAULT '{}'::jsonb,
    ip_address      INET,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_events_actor
    ON audit_events (actor_user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_events_org
    ON audit_events (organization_id, created_at DESC);

-- ADR-006: Descontinuação de CPF obrigatório (colunas passam a aceitar NULL)
ALTER TABLE students
    ALTER COLUMN cpf_encrypted DROP NOT NULL,
    ALTER COLUMN cpf_hash DROP NOT NULL;

-- Backfill idempotente: organiza treinadores existentes em organizações SOLO
INSERT INTO organizations (name, slug, type)
SELECT
    COALESCE(NULLIF(TRIM(u.full_name), ''), 'Personal') || ' Studio',
    'org-' || replace(u.id::text, '-', ''),
    'SOLO'
FROM users u
WHERE u.role = 'TRAINER'
  AND u.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM memberships m WHERE m.user_id = u.id
  )
ON CONFLICT (slug) DO NOTHING;

INSERT INTO memberships (organization_id, user_id, role, status)
SELECT o.id, u.id, 'OWNER', 'ACTIVE'
FROM users u
INNER JOIN organizations o ON o.slug = 'org-' || replace(u.id::text, '-', '')
WHERE u.role = 'TRAINER'
  AND u.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM memberships m WHERE m.organization_id = o.id AND m.user_id = u.id
  )
ON CONFLICT (organization_id, user_id) DO NOTHING;

-- Backfill idempotente: migra students.trainer_id existente para trainer_student_relationships
INSERT INTO trainer_student_relationships (organization_id, trainer_user_id, student_id, status)
SELECT DISTINCT ON (s.trainer_id, s.id)
    m.organization_id,
    s.trainer_id,
    s.id,
    'ACTIVE'
FROM students s
INNER JOIN memberships m ON m.user_id = s.trainer_id AND m.status = 'ACTIVE'
WHERE s.trainer_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM trainer_student_relationships r
      WHERE r.trainer_user_id = s.trainer_id
        AND r.student_id = s.id
        AND r.status = 'ACTIVE'
  );
