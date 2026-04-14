CREATE TABLE students (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    cpf_encrypted TEXT NOT NULL,                -- AES-256-GCM (Nunca nulo, para podermos decriptografar ao visualizar)
    cpf_hash      VARCHAR(64) NOT NULL UNIQUE,  -- HMAC do CPF original (Para buscas seguras ex: logar)
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexamos apenas o Hash de busca para velocidade absurdamente alta sem expor informações
CREATE INDEX idx_students_cpf_hash ON students (cpf_hash);

CREATE TABLE biometric_data (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id          UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    weight_kg           NUMERIC(5,2) NOT NULL,
    height_cm           NUMERIC(5,2) NOT NULL,
    bmi                 NUMERIC(4,2), -- Pode ser calculado direto no software, mas armazenamos como cache/historico
    body_fat_percentage NUMERIC(4,2),
    muscle_mass_kg      NUMERIC(5,2),
    notes_encrypted     TEXT,         -- GCM com anotações medicas sigilosas
    recorded_by         UUID NOT NULL REFERENCES users(id), 
    recorded_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Auditorias
CREATE TABLE audit_logs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID REFERENCES users(id) ON DELETE SET NULL, -- Quem fez a ação
    action        VARCHAR(50) NOT NULL,                         -- Ex: CREATE_STUDENT, DELETE_STUDENT_LGPD
    resource_type VARCHAR(50) NOT NULL,                         -- Ex: student, biometric
    resource_id   UUID,                                         
    changes       JSONB,                                        -- Diferencial do que mudou
    ip_address    INET,
    user_agent    TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_resource ON audit_logs (resource_type, resource_id);
CREATE INDEX idx_audit_logs_user ON audit_logs (user_id, created_at DESC);
