-- migrations/009_workout_templates_and_versions.up.sql
-- Fase 5: Workout Builder (BL-14)
-- Templates reutilizáveis por organização, snapshots imutáveis (workout_versions)
-- e atribuições versionadas para alunos (workout_assignments), com backfill idempotente do legado.

CREATE TABLE IF NOT EXISTS workout_templates (
    id                       UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id          UUID         NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    trainer_user_id          UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title                    VARCHAR(255) NOT NULL,
    description              TEXT         NOT NULL DEFAULT '',
    goal                     VARCHAR(120) NOT NULL DEFAULT '',
    draft_structure          JSONB        NOT NULL DEFAULT '{"sessions":[]}'::jsonb,
    cloned_from_template_id  UUID         REFERENCES workout_templates(id) ON DELETE SET NULL,
    is_archived              BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workout_templates_org_archived
    ON workout_templates (organization_id, is_archived, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_workout_templates_trainer
    ON workout_templates (trainer_user_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS workout_versions (
    id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id          UUID         NOT NULL REFERENCES workout_templates(id) ON DELETE CASCADE,
    version_number       INT          NOT NULL CHECK (version_number >= 1),
    title                VARCHAR(255) NOT NULL,
    description          TEXT         NOT NULL DEFAULT '',
    goal                 VARCHAR(120) NOT NULL DEFAULT '',
    snapshot_structure   JSONB        NOT NULL,
    published_by_user_id UUID         REFERENCES users(id) ON DELETE SET NULL,
    published_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (template_id, version_number)
);

CREATE INDEX IF NOT EXISTS idx_workout_versions_template
    ON workout_versions (template_id, version_number DESC);

CREATE OR REPLACE FUNCTION prevent_workout_version_update()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'workout_versions are immutable and cannot be updated';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_workout_versions_immutable ON workout_versions;
CREATE TRIGGER trg_workout_versions_immutable
    BEFORE UPDATE ON workout_versions
    FOR EACH ROW
    EXECUTE FUNCTION prevent_workout_version_update();

CREATE TABLE IF NOT EXISTS workout_assignments (
    id                     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id        UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    template_id            UUID        NOT NULL REFERENCES workout_templates(id) ON DELETE CASCADE,
    workout_version_id     UUID        NOT NULL REFERENCES workout_versions(id) ON DELETE RESTRICT,
    student_id             UUID        NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    trainer_user_id        UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    legacy_workout_plan_id UUID        REFERENCES workout_plans(id) ON DELETE SET NULL,
    valid_from             DATE        NOT NULL DEFAULT CURRENT_DATE,
    valid_until            DATE,
    status                 VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'ARCHIVED')),
    notes                  TEXT        NOT NULL DEFAULT '',
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workout_assignments_student_status
    ON workout_assignments (student_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_workout_assignments_org_trainer
    ON workout_assignments (organization_id, trainer_user_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_workout_assignments_version
    ON workout_assignments (workout_version_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_workout_assignments_legacy_plan
    ON workout_assignments (legacy_workout_plan_id)
    WHERE legacy_workout_plan_id IS NOT NULL;

-- Backfill idempotente: converte planos legados (workout_plans) em templates + v1 + assignments
DO $$
DECLARE
    wp RECORD;
    v_org_id UUID;
    v_template_id UUID;
    v_version_id UUID;
    v_structure JSONB;
BEGIN
    FOR wp IN
        SELECT p.*
        FROM workout_plans p
        WHERE NOT EXISTS (
            SELECT 1 FROM workout_assignments wa WHERE wa.legacy_workout_plan_id = p.id
        )
    LOOP
        SELECT COALESCE(
            (
                SELECT r.organization_id
                FROM trainer_student_relationships r
                WHERE r.trainer_user_id = wp.trainer_id
                  AND r.student_id = wp.student_id
                ORDER BY r.started_at DESC
                LIMIT 1
            ),
            (
                SELECT m.organization_id
                FROM memberships m
                WHERE m.user_id = wp.trainer_id
                  AND m.status = 'ACTIVE'
                ORDER BY m.created_at ASC
                LIMIT 1
            )
        ) INTO v_org_id;

        IF v_org_id IS NULL THEN
            CONTINUE;
        END IF;

        SELECT jsonb_build_object(
            'sessions',
            COALESCE(
                (
                    SELECT jsonb_agg(
                        jsonb_build_object(
                            'id', ws.id,
                            'title', ws.title,
                            'day_of_week', COALESCE(ws.day_of_week, 0),
                            'notes', COALESCE(ws.notes, ''),
                            'position', ws.pos,
                            'exercises', COALESCE(
                                (
                                    SELECT jsonb_agg(
                                        jsonb_strip_nulls(
                                            jsonb_build_object(
                                                'id', e.id,
                                                'legacy_exercise_id', e.id,
                                                'catalog_exercise_id', e.catalog_exercise_id,
                                                'name', e.name,
                                                'sets', COALESCE(e.sets, 0),
                                                'reps', COALESCE(e.reps, ''),
                                                'rest_seconds', COALESCE(e.rest_seconds, 0),
                                                'notes', COALESCE(e.notes, ''),
                                                'position', COALESCE(e.position, 0)
                                            )
                                        )
                                        ORDER BY e.position, e.created_at
                                    )
                                    FROM exercises e
                                    WHERE e.workout_session_id = ws.id
                                ),
                                '[]'::jsonb
                            )
                        )
                        ORDER BY ws.pos
                    )
                    FROM (
                        SELECT
                            ws_inner.id,
                            ws_inner.title,
                            ws_inner.day_of_week,
                            ws_inner.notes,
                            ROW_NUMBER() OVER (ORDER BY ws_inner.day_of_week, ws_inner.created_at) AS pos
                        FROM workout_sessions ws_inner
                        WHERE ws_inner.workout_plan_id = wp.id
                    ) ws
                ),
                '[]'::jsonb
            )
        ) INTO v_structure;

        v_template_id := gen_random_uuid();
        v_version_id := gen_random_uuid();

        INSERT INTO workout_templates (
            id, organization_id, trainer_user_id, title, description, goal, draft_structure, created_at, updated_at
        ) VALUES (
            v_template_id,
            v_org_id,
            wp.trainer_id,
            wp.title,
            COALESCE(wp.description, ''),
            '',
            v_structure,
            wp.created_at,
            wp.updated_at
        );

        INSERT INTO workout_versions (
            id, template_id, version_number, title, description, goal, snapshot_structure, published_by_user_id, published_at
        ) VALUES (
            v_version_id,
            v_template_id,
            1,
            wp.title,
            COALESCE(wp.description, ''),
            '',
            v_structure,
            wp.trainer_id,
            wp.updated_at
        );

        INSERT INTO workout_assignments (
            id, organization_id, template_id, workout_version_id, student_id, trainer_user_id,
            legacy_workout_plan_id, valid_from, status, created_at, updated_at
        ) VALUES (
            gen_random_uuid(),
            v_org_id,
            v_template_id,
            v_version_id,
            wp.student_id,
            wp.trainer_id,
            wp.id,
            wp.created_at::date,
            CASE WHEN wp.is_active THEN 'ACTIVE' ELSE 'ARCHIVED' END,
            wp.created_at,
            wp.updated_at
        );
    END LOOP;
END $$;
