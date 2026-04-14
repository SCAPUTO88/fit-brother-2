CREATE TABLE workout_plans (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    trainer_id  UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id  UUID        NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workout_plans_trainer ON workout_plans (trainer_id);
CREATE INDEX idx_workout_plans_student ON workout_plans (student_id);

CREATE TABLE workout_sessions (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_plan_id UUID        NOT NULL REFERENCES workout_plans(id) ON DELETE CASCADE,
    title           VARCHAR(255) NOT NULL,
    day_of_week     SMALLINT,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workout_sessions_plan ON workout_sessions (workout_plan_id);

CREATE TABLE exercises (
    id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_session_id UUID        NOT NULL REFERENCES workout_sessions(id) ON DELETE CASCADE,
    name               VARCHAR(255) NOT NULL,
    sets               SMALLINT,
    reps               VARCHAR(50),
    rest_seconds       SMALLINT,
    notes              TEXT,
    video_url          TEXT,
    position           SMALLINT    NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exercises_session ON exercises (workout_session_id, position);
