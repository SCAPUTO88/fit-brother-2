-- migrations/006_security_baseline_ownership_and_video_status.up.sql
ALTER TABLE students
    ADD COLUMN IF NOT EXISTS trainer_id UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_students_trainer ON students (trainer_id);

-- Backfill determinístico para ambientes existentes com um único treinador por aluno
UPDATE students s
SET trainer_id = wp.trainer_id
FROM (
    SELECT student_id, MIN(trainer_id::text)::uuid AS trainer_id
    FROM workout_plans
    GROUP BY student_id
    HAVING COUNT(DISTINCT trainer_id) = 1
) wp
WHERE s.id = wp.student_id
  AND s.trainer_id IS NULL;

-- Em ambientes locais com exatamente um treinador ativo, associa alunos legados sem vínculo
UPDATE students
SET trainer_id = (
    SELECT id FROM users
    WHERE role = 'TRAINER' AND deleted_at IS NULL
    LIMIT 1
)
WHERE trainer_id IS NULL
  AND (SELECT COUNT(*) FROM users WHERE role = 'TRAINER' AND deleted_at IS NULL) = 1;

CREATE INDEX IF NOT EXISTS idx_workout_plans_trainer_student
    ON workout_plans (trainer_id, student_id);

ALTER TABLE videos
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'CONFIRMED'
    CHECK (status IN ('PENDING', 'CONFIRMED'));

CREATE INDEX IF NOT EXISTS idx_videos_exercise_status
    ON videos (exercise_id, status);
