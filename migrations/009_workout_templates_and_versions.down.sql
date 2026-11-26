-- migrations/009_workout_templates_and_versions.down.sql
DROP TABLE IF EXISTS workout_assignments CASCADE;
DROP TRIGGER IF EXISTS trg_workout_versions_immutable ON workout_versions;
DROP FUNCTION IF EXISTS prevent_workout_version_update();
DROP TABLE IF EXISTS workout_versions CASCADE;
DROP TABLE IF EXISTS workout_templates CASCADE;
