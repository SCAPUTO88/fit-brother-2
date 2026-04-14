CREATE TABLE refresh_tokens (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(64) NOT NULL UNIQUE,  -- SHA-256 do token raw
    family_id   UUID        NOT NULL,          -- agrupa tokens de uma mesma sessão
    is_revoked  BOOLEAN     NOT NULL DEFAULT FALSE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_ip  INET                           -- auditoria de segurança
);

-- Busca rápida por hash do token (usado em cada refresh)
CREATE INDEX idx_refresh_tokens_hash   ON refresh_tokens (token_hash);
-- Revogação de família inteira (detecção de roubo de token)
CREATE INDEX idx_refresh_tokens_family ON refresh_tokens (family_id);
-- Limpeza de tokens expirados (job periódico)
CREATE INDEX idx_refresh_tokens_expiry ON refresh_tokens (expires_at) WHERE is_revoked = FALSE;
