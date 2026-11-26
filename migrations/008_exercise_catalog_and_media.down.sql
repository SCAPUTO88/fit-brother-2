-- migrations/008_exercise_catalog_and_media.down.sql

ALTER TABLE exercises
    DROP COLUMN IF EXISTS catalog_exercise_id;

DROP TABLE IF EXISTS exercise_media;
DROP TABLE IF EXISTS exercise_catalog;
DROP TABLE IF EXISTS equipments;
DROP TABLE IF EXISTS muscle_groups;
