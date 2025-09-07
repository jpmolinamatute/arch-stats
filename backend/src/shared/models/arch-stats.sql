-- This is a reference file and describe the tables, views, and other database objects in the 
-- arch-stats database.

-- An archer opens a session, register arrows and calibrate targets with optional target faces.
-- Only one session can be is_opened = TRUE at a time.
-- Afterwards, the archer can record arrow shots.
-- Scoring ONLY happens when there are target faces associated with a target.
-- If the system record shots.arrow_landing_time, shots.x, and shots.y, it means that the arrow
-- landed somewhere in the target butt. If target faces are present, they can be used for scoring.
-- When scoring, the system checks:
-- if the shot landed on a target face: the system scores the shot.
-- if the shot misses the target face BUT landed on the target butt: The system scores 0.
-- if the shot misses the target entirely: The system does not score the shot (aka null).

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    is_opened BOOLEAN NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    location VARCHAR(255) NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    is_indoor BOOLEAN DEFAULT FALSE,
    CHECK (
        (is_opened AND end_time IS NULL)
        OR (NOT is_opened AND end_time IS NOT NULL)
    )
);

CREATE TABLE IF NOT EXISTS targets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    max_x REAL NOT NULL,
    max_y REAL NOT NULL,
    distance INTEGER NOT NULL,
    session_id UUID NOT NULL,
    CONSTRAINT targets_one_per_session UNIQUE (session_id),
    FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS faces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    x REAL NOT NULL,
    y REAL NOT NULL,
    human_identifier TEXT NOT NULL,
    radii REAL [] NOT NULL,
    points INTEGER [] NOT NULL,
    target_id UUID NOT NULL,
    FOREIGN KEY (target_id) REFERENCES targets (id) ON DELETE CASCADE,
    UNIQUE (target_id, human_identifier),
    CHECK (
        array_length(radii, 1) > 0
        AND array_length(points, 1) > 0
        AND array_length(points, 1) = array_length(radii, 1)
        AND validate_face_row(target_id, x, y, radii)
    )
);

CREATE TABLE IF NOT EXISTS arrows (
    id UUID PRIMARY KEY,
    length REAL NOT NULL,
    human_identifier VARCHAR(10) NOT NULL,
    is_programmed BOOLEAN NOT NULL DEFAULT FALSE,
    registration_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    voided_date TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    label_position REAL,
    weight REAL,
    diameter REAL,
    spine REAL,
    CONSTRAINT arrows_active_voided_consistency CHECK (
        (is_active = FALSE AND voided_date IS NOT NULL)
        OR (is_active = TRUE AND voided_date IS NULL)
    )
);

CREATE TABLE IF NOT EXISTS shots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    arrow_id UUID NOT NULL,
    session_id UUID NOT NULL,
    arrow_engage_time TIMESTAMP WITH TIME ZONE NOT NULL,
    arrow_disengage_time TIMESTAMP WITH TIME ZONE NOT NULL,
    arrow_landing_time TIMESTAMP WITH TIME ZONE,
    x REAL,
    y REAL,
    FOREIGN KEY (arrow_id) REFERENCES arrows (id) ON DELETE CASCADE,
    FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE,
    CHECK (
        (
            arrow_landing_time IS NOT NULL
            AND x IS NOT NULL
            AND y IS NOT NULL
        )
        OR
        (
            arrow_landing_time IS NULL
            AND x IS NULL
            AND y IS NULL
        )
    )
);

CREATE OR REPLACE VIEW performance AS
SELECT
    shots.id,
    shots.arrow_engage_time,
    shots.arrow_disengage_time,
    shots.arrow_landing_time,
    shots.x,
    shots.y,
    arrows.human_identifier AS arrow_human_identifier,
    arrows.id AS arrow_id,
    extract(
        EPOCH FROM (shots.arrow_landing_time - shots.arrow_disengage_time)
    ) AS time_of_flight_seconds,
    targets.distance::FLOAT / nullif(
        extract(EPOCH FROM (shots.arrow_landing_time - shots.arrow_disengage_time)),
        0
    ) AS arrow_speed,
    get_shot_score(shots.x, shots.y, targets.id, targets.max_x, targets.max_y) AS score
FROM
    shots
INNER JOIN
    arrows ON shots.arrow_id = arrows.id
INNER JOIN
    targets ON shots.session_id = targets.session_id
WHERE shots.session_id = (
    SELECT sessions.id
    FROM sessions
    WHERE sessions.is_opened = TRUE
    ORDER BY sessions.start_time DESC
    LIMIT 1
);

CREATE INDEX IF NOT EXISTS idx_faces_target_id ON faces (target_id);
CREATE INDEX IF NOT EXISTS idx_targets_session_id ON targets (session_id);
CREATE INDEX IF NOT EXISTS idx_shots_arrow_id ON shots (arrow_id);
CREATE INDEX IF NOT EXISTS idx_shots_session_id ON shots (session_id);
CREATE UNIQUE INDEX IF NOT EXISTS arrows_uniq_active_human_identifier
ON arrows (human_identifier)
WHERE is_active IS TRUE;

CREATE OR REPLACE FUNCTION validate_face_row(
    target_id UUID,
    x REAL,
    y REAL,
    radii REAL []
) RETURNS BOOLEAN AS $$
DECLARE
    max_radius REAL;
    tgt_max_x REAL;
    tgt_max_y REAL;
BEGIN
    SELECT max_x, max_y INTO tgt_max_x, tgt_max_y FROM targets WHERE id = target_id;

    -- Compute max radius
    SELECT MAX(r) INTO max_radius FROM unnest(radii) AS r;

    -- Ensure face fully contained within target bounds (strictly inside)
    IF (x + max_radius) > tgt_max_x OR (y + max_radius) > tgt_max_y THEN
        RETURN FALSE;
    END IF;
    IF (x - max_radius) < 0 OR (y - max_radius) < 0 THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql STABLE;

CREATE OR REPLACE FUNCTION notify_new_shot_archy() RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('archy', row_to_json(NEW)::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'shots_notify_trigger_archy'
    ) THEN
        EXECUTE format(
            'CREATE TRIGGER shots_notify_trigger_%s AFTER INSERT ON shots
                FOR EACH ROW EXECUTE FUNCTION notify_new_shot_%s();',
            'archy', 'archy');
    END IF;
END$$;

CREATE OR REPLACE FUNCTION get_shot_score(
    shot_x REAL,
    shot_y REAL,
    target_id UUID,
    target_max_x REAL,
    target_max_y REAL
) RETURNS INTEGER AS $$
DECLARE
    face RECORD;
    distance REAL;
    score INTEGER := NULL;
    max_score INTEGER := 0;
    face_count INTEGER;
BEGIN
    -- If shot coordinates are null, it missed the target entirely
    IF shot_x IS NULL OR shot_y IS NULL THEN
        RETURN NULL;
    END IF;

    -- Check unrealistic shot coordinates
    IF shot_x < 0 OR shot_y < 0 OR shot_x > target_max_x OR shot_y > target_max_y THEN
        RETURN NULL;
    END IF;

    -- Check if any faces exist for the target
    SELECT COUNT(*) INTO face_count
    FROM faces
    WHERE target_id = get_shot_score.target_id
    LIMIT 3;

    -- If no faces exist, return NULL
    IF face_count = 0 THEN
        RETURN NULL;
    END IF;

    -- Check up to 3 faces for the target
    FOR face IN (
        SELECT x, y, radii, points
        FROM faces
        WHERE target_id = get_shot_score.target_id
    ) LOOP
        distance := SQRT(POWER(shot_x - face.x, 2) + POWER(shot_y - face.y, 2));
        FOR i IN 1..array_length(face.radii, 1) LOOP
            IF distance <= face.radii[i] THEN
                score := face.points[i];
                IF max_score IS NULL OR score > max_score THEN
                    max_score := score;
                END IF;
                EXIT;
            END IF;
        END LOOP;
    END LOOP;

    -- Return the highest score, or NULL if no face was hit
    RETURN max_score;
END;
$$ LANGUAGE plpgsql STABLE;
