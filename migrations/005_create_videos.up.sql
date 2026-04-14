CREATE TABLE videos (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    exercise_id UUID        NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    uploader_id UUID        NOT NULL REFERENCES users(id),
    r2_key      TEXT        NOT NULL UNIQUE,
    filename    TEXT        NOT NULL,
    size_bytes  BIGINT,
    mime_type   VARCHAR(100),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_videos_exercise ON videos (exercise_id);
