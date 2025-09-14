--
-- Sessions and Targets
--

-- A session represents an archery practice, and only one can be open at a time.
-- Each session must have exactly one associated target.
-- Target faces are optional, but if present, they must be contained entirely within the 
-- target's boundaries.

--
-- Shots and Scoring
--

-- An arrow shot can be recorded with or without landing coordinates and time.
-- If a shot has landing coordinates, it means the arrow hit the target butt.
-- A shot is only scored if there are target faces on the target.
-- If a shot with landing coordinates hits a target face, it gets a score based on the face's rings.
-- If a shot with landing coordinates misses all target faces but hits the target butt, 
-- it receives a score of 0.
-- If a shot misses the entire target, it receives a NULL score.

--
-- Arrows
--

-- An arrow's human_identifier must be unique among all active arrows.
-- An arrow is either active (not voided) or inactive (voided), but not both.
-- The is_active and voided_date fields must be consistent.

-- 
-- UNITS
--

-- All units must be in metric system

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE OR REPLACE FUNCTION validate_face_row(
    target_id UUID,
    x REAL,
    y REAL,
    wa_faces_id UUID
) RETURNS BOOLEAN AS $$
DECLARE
    max_radius REAL;
    tgt_max_x REAL;
    tgt_max_y REAL;
BEGIN
    SELECT max_x, max_y
    INTO tgt_max_x, tgt_max_y
    FROM targets
    WHERE id = target_id;
    
    -- Compute max radius
    SELECT MAX(wa_face_rings.diameter)/2
    INTO max_radius
    FROM wa_face_rings
    WHERE wa_face_rings.wa_faces_id = validate_face_row.wa_faces_id
    GROUP BY wa_face_rings.wa_faces_id;

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

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    is_opened BOOLEAN NOT NULL DEFAULT FALSE,
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
    target_sensor_id UUID NOT NULL,
    session_id UUID NOT NULL,
    lane INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT targets_one_per_session UNIQUE (session_id, target_sensor_id),
    FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS wa_faces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS wa_face_rings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    diameter REAL NOT NULL,
    point INTEGER NOT NULL,
    color TEXT NOT NULL,
    wa_faces_id UUID NOT NULL,
    FOREIGN KEY (wa_faces_id) REFERENCES wa_faces (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS faces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    x REAL NOT NULL,
    y REAL NOT NULL,
    human_identifier TEXT NOT NULL,
    is_reduced BOOLEAN NOT NULL DEFAULT FALSE,
    wa_faces_id UUID NOT NULL,
    target_id UUID NOT NULL,
    FOREIGN KEY (target_id) REFERENCES targets (id) ON DELETE CASCADE,
    FOREIGN KEY (wa_faces_id) REFERENCES wa_faces (id) ON DELETE CASCADE,
    CONSTRAINT faces_unique_target_human UNIQUE (target_id, human_identifier),
    CHECK (validate_face_row(target_id, x, y, wa_faces_id))
);

CREATE OR REPLACE FUNCTION get_shot_score(
    shot_x REAL,
    shot_y REAL,
    target_id UUID,
    target_max_x REAL,
    target_max_y REAL
) RETURNS INTEGER AS $$
    SELECT
      CASE
        WHEN $1 IS NULL OR $2 IS NULL THEN NULL
        WHEN $1 < 0 OR $2 < 0 OR $1 > $4 OR $2 > $5 THEN NULL
        WHEN NOT EXISTS (SELECT 1 FROM faces f WHERE f.target_id = $3) THEN NULL
        ELSE COALESCE(
          (
            SELECT MAX(r.point)::int
            FROM faces f
            JOIN wa_face_rings r ON r.wa_faces_id = f.wa_faces_id
            WHERE f.target_id = $3
              AND (($1 - f.x) * ($1 - f.x) + ($2 - f.y) * ($2 - f.y))
                    <= (r.diameter * r.diameter) / 4.0
          ),
          0
        )
      END
$$ LANGUAGE sql STABLE;


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
    target_sensor_id UUID NOT NULL,
    arrow_engage_time TIMESTAMP WITH TIME ZONE NOT NULL,
    arrow_disengage_time TIMESTAMP WITH TIME ZONE NOT NULL,
    arrow_landing_time TIMESTAMP WITH TIME ZONE,
    x REAL,
    y REAL,
    FOREIGN KEY (arrow_id) REFERENCES arrows (id) ON DELETE CASCADE,
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

CREATE INDEX IF NOT EXISTS idx_faces_target_id ON faces (target_id);
CREATE INDEX IF NOT EXISTS idx_targets_session_id ON targets (session_id);
CREATE INDEX IF NOT EXISTS idx_shots_arrow_id ON shots (arrow_id);
-- New: speed up filtering shots by current target sensor
CREATE INDEX IF NOT EXISTS idx_shots_target_sensor_id ON shots (target_sensor_id);
CREATE UNIQUE INDEX IF NOT EXISTS arrows_uniq_hid
ON arrows (human_identifier)
WHERE is_active IS TRUE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_one_open_id
ON sessions (is_opened) WHERE is_opened IS TRUE;


CREATE TRIGGER shots_notify_trigger_archy
AFTER INSERT ON shots
FOR EACH ROW EXECUTE FUNCTION notify_new_shot_archy();

-- for now this view will contain only one target because we are currently supporting one target per
-- session.
CREATE MATERIALIZED VIEW current_target AS
SELECT
    id,
    max_x,
    max_y,
    distance,
    target_sensor_id,
    session_id,
    lane
FROM targets
WHERE
    session_id = (
        SELECT sessions.id
        FROM sessions
        WHERE sessions.is_opened = TRUE
        ORDER BY sessions.start_time DESC
        LIMIT 1
    )
WITH DATA;

CREATE OR REPLACE VIEW session_performance AS
SELECT
    shots.id,
    shots.arrow_engage_time,
    shots.arrow_disengage_time,
    shots.arrow_landing_time,
    shots.x,
    shots.y,
    arrows.id AS arrow_id,
    arrows.human_identifier,
    current_target.lane,
    extract(
        EPOCH FROM (shots.arrow_landing_time - shots.arrow_disengage_time)
    ) AS time_of_flight_seconds,
    current_target.distance::FLOAT / nullif(
        extract(EPOCH FROM (shots.arrow_landing_time - shots.arrow_disengage_time)),
        0
    ) AS arrow_speed,
    get_shot_score(
        shots.x, shots.y, current_target.id, current_target.max_x, current_target.max_y
    ) AS score
FROM
    shots
INNER JOIN
    arrows ON shots.arrow_id = arrows.id
INNER JOIN
    current_target ON shots.target_sensor_id = current_target.target_sensor_id;


INSERT INTO wa_faces (id, name) VALUES
('bcfda882-5d4b-4c94-8911-468cc5d7d60f', '122 cm'),
('308dc394-e0c0-4dbd-835d-8afe7d59ce4e', '80 cm'),
('8ba08ff8-364b-4a46-914c-17f1311f3f1a', '60 cm'),
('417d902f-3a4b-4262-a530-6d36f3dd47e7', '40 cm');

-- source https://archerygeekery.co.uk/2023/05/05/accurate-svg-archery-target-images/
INSERT INTO wa_face_rings (wa_faces_id, diameter, point, color) VALUES
('bcfda882-5d4b-4c94-8911-468cc5d7d60f', 6.10, 11, '#FFE552'),
('bcfda882-5d4b-4c94-8911-468cc5d7d60f', 12.20, 10, '#FFE552'),
('bcfda882-5d4b-4c94-8911-468cc5d7d60f', 24.40, 9, '#FFE552'),
('bcfda882-5d4b-4c94-8911-468cc5d7d60f', 36.60, 8, '#F65058'),
('bcfda882-5d4b-4c94-8911-468cc5d7d60f', 48.80, 7, '#F65058'),
('bcfda882-5d4b-4c94-8911-468cc5d7d60f', 61.00, 6, '#00B4E4'),
('bcfda882-5d4b-4c94-8911-468cc5d7d60f', 73.20, 5, '#00B4E4'),
('bcfda882-5d4b-4c94-8911-468cc5d7d60f', 85.40, 4, '#000000'),
('bcfda882-5d4b-4c94-8911-468cc5d7d60f', 97.60, 3, '#000000'),
('bcfda882-5d4b-4c94-8911-468cc5d7d60f', 109.80, 2, '#FFFFFF'),
('bcfda882-5d4b-4c94-8911-468cc5d7d60f', 122.00, 1, '#FFFFFF'),

('308dc394-e0c0-4dbd-835d-8afe7d59ce4e', 4.00, 11, '#FFE552'),
('308dc394-e0c0-4dbd-835d-8afe7d59ce4e', 8.00, 10, '#FFE552'),
('308dc394-e0c0-4dbd-835d-8afe7d59ce4e', 16.00, 9, '#FFE552'),
('308dc394-e0c0-4dbd-835d-8afe7d59ce4e', 24.00, 8, '#F65058'),
('308dc394-e0c0-4dbd-835d-8afe7d59ce4e', 32.00, 7, '#F65058'),
('308dc394-e0c0-4dbd-835d-8afe7d59ce4e', 40.00, 6, '#00B4E4'),
('308dc394-e0c0-4dbd-835d-8afe7d59ce4e', 48.00, 5, '#00B4E4'),
('308dc394-e0c0-4dbd-835d-8afe7d59ce4e', 56.00, 4, '#000000'),
('308dc394-e0c0-4dbd-835d-8afe7d59ce4e', 64.00, 3, '#000000'),
('308dc394-e0c0-4dbd-835d-8afe7d59ce4e', 72.00, 2, '#FFFFFF'),
('308dc394-e0c0-4dbd-835d-8afe7d59ce4e', 80.00, 1, '#FFFFFF'),

('8ba08ff8-364b-4a46-914c-17f1311f3f1a', 3.00, 11, '#FFE552'),
('8ba08ff8-364b-4a46-914c-17f1311f3f1a', 6.00, 10, '#FFE552'),
('8ba08ff8-364b-4a46-914c-17f1311f3f1a', 12.00, 9, '#FFE552'),
('8ba08ff8-364b-4a46-914c-17f1311f3f1a', 18.00, 8, '#F65058'),
('8ba08ff8-364b-4a46-914c-17f1311f3f1a', 24.00, 7, '#F65058'),
('8ba08ff8-364b-4a46-914c-17f1311f3f1a', 30.00, 6, '#00B4E4'),
('8ba08ff8-364b-4a46-914c-17f1311f3f1a', 36.00, 5, '#00B4E4'),
('8ba08ff8-364b-4a46-914c-17f1311f3f1a', 42.00, 4, '#000000'),
('8ba08ff8-364b-4a46-914c-17f1311f3f1a', 48.00, 3, '#000000'),
('8ba08ff8-364b-4a46-914c-17f1311f3f1a', 54.00, 2, '#FFFFFF'),
('8ba08ff8-364b-4a46-914c-17f1311f3f1a', 60.00, 1, '#FFFFFF'),

('417d902f-3a4b-4262-a530-6d36f3dd47e7', 2.00, 11, '#FFE552'),
('417d902f-3a4b-4262-a530-6d36f3dd47e7', 4.00, 10, '#FFE552'),
('417d902f-3a4b-4262-a530-6d36f3dd47e7', 8.00, 9, '#FFE552'),
('417d902f-3a4b-4262-a530-6d36f3dd47e7', 12.00, 8, '#F65058'),
('417d902f-3a4b-4262-a530-6d36f3dd47e7', 16.00, 7, '#F65058'),
('417d902f-3a4b-4262-a530-6d36f3dd47e7', 20.00, 6, '#00B4E4'),
('417d902f-3a4b-4262-a530-6d36f3dd47e7', 24.00, 5, '#00B4E4'),
('417d902f-3a4b-4262-a530-6d36f3dd47e7', 28.00, 4, '#000000'),
('417d902f-3a4b-4262-a530-6d36f3dd47e7', 32.00, 3, '#000000'),
('417d902f-3a4b-4262-a530-6d36f3dd47e7', 36.00, 2, '#FFFFFF'),
('417d902f-3a4b-4262-a530-6d36f3dd47e7', 40.00, 1, '#FFFFFF');
