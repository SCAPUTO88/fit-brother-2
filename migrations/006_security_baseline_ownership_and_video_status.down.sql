-- migrations/006_security_baseline_ownership_and_video_status.down.sql
DROP INDEX IF EXISTS idx_videos_exercise_status;
ALTER TABLE videos DROP COLUMN IF EXISTS status;

DROP INDEX IF EXISTS idx_workout_plans_trainer_student;
DROP INDEX IF EXISTS idx_students_trainer;
ALTER TABLE students DROP COLUMN IF EXISTS trainer_id;
