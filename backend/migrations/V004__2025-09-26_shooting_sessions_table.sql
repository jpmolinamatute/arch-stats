CREATE TYPE face_type AS ENUM (
    '40cm_full',
    '60cm_full',
    '80cm_full',
    '122cm_full',
    '40cm_6rings',
    '60cm_6rings',
    '80cm_6rings',
    '122cm_6rings',
    '40cm_triple_vertical',
    '60cm_triple_triangular',
    'none'
);

-- Table to store shooting session information
CREATE TABLE session (
    session_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_archer_id UUID NOT NULL REFERENCES archer (archer_id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    closed_at TIMESTAMPTZ,
    session_location VARCHAR(255) NOT NULL,
    is_indoor BOOLEAN NOT NULL,
    is_opened BOOLEAN NOT NULL,
    CONSTRAINT sessions_time_check CHECK (
        closed_at IS NULL OR closed_at > created_at
    )
);

-- Table to store target butts and lane information
CREATE TABLE target (
    target_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES session (session_id) ON DELETE CASCADE,
    distance INTEGER NOT NULL,
    lane INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT targets_one_per_session UNIQUE (session_id, lane),
    CONSTRAINT targets_lane_bounds CHECK (lane BETWEEN 1 AND 100),
    CONSTRAINT targets_distance_bounds CHECK (distance BETWEEN 1 AND 100)
);

-- Table to stores slot assignments of participating archer in a target
-- butt or lane within a specific shooting session.
-- It also stores bowstyle, draw_weight and club_id used for a specific shooting session
CREATE TABLE slot (
    slot_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    target_id UUID NOT NULL REFERENCES target (target_id) ON DELETE CASCADE,
    archer_id UUID NOT NULL REFERENCES archer (archer_id) ON DELETE RESTRICT,
    session_id UUID NOT NULL REFERENCES session (session_id) ON DELETE CASCADE,
    face_type FACE_TYPE NOT NULL,
    slot_letter CHAR(1) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_shooting BOOLEAN NOT NULL,
    bowstyle BOWSTYLE_TYPE NOT NULL,
    draw_weight FLOAT NOT NULL CHECK (draw_weight > 0),
    club_id UUID,
    CONSTRAINT uq_archer_per_session UNIQUE (archer_id, session_id),
    CONSTRAINT slot_letter_valid CHECK (slot_letter IN ('A', 'B', 'C', 'D')),
    CONSTRAINT uq_slot_on_target UNIQUE (target_id, slot_letter, archer_id)
);

-- NOTE: bowstyle, draw_weight and current_club_id can change over time but NOT with in a session.
-- The information stored in the slot table will hold a historical data of what was used
-- at a given point in time.


CREATE UNIQUE INDEX unique_open_session_per_archer
ON session (owner_archer_id)
WHERE is_opened = TRUE;

-- read-only view to get all currently shooting archer in open session
CREATE VIEW slot_shooting_participants AS
SELECT
    slot.slot_id,
    slot.target_id,
    slot.archer_id,
    slot.session_id,
    slot.face_type,
    slot.slot_letter,
    slot.created_at,
    slot.bowstyle,
    slot.draw_weight,
    slot.is_shooting
FROM slot
WHERE slot.session_id IN (
    SELECT session.session_id
    FROM session
    WHERE session.is_opened IS TRUE
)
AND slot.is_shooting IS TRUE;

-- Function to get the next "empty" lane
CREATE FUNCTION get_next_lane(p_session_id UUID)
RETURNS INTEGER AS $$
    SELECT MAX(lane) + 1
    FROM target
    JOIN session
    ON target.session_id = session.session_id
    WHERE target.session_id = p_session_id
    AND session.is_opened IS TRUE;
$$ LANGUAGE sql STABLE;

-- Function to get all target with available slot with a specific distance within a session
CREATE FUNCTION get_available_targets(p_session_id UUID, p_distance INTEGER)
RETURNS TABLE (
    target_id UUID,
    session_id UUID,
    distance INTEGER,
    lane INTEGER,
    occupied INTEGER,
    created_at TIMESTAMPTZ
) AS
$$
SELECT
    target.target_id,
    target.session_id,
    target.distance,
    target.lane,
    COUNT(slot.slot_letter) AS occupied,
    target.created_at
FROM target
LEFT JOIN slot
    ON target.target_id = slot.target_id
WHERE target.session_id = p_session_id
    AND target.distance = p_distance
GROUP BY target.target_id, target.session_id, target.distance, target.lane
HAVING COUNT(slot.slot_letter) < 4
ORDER BY target.lane ASC
$$ LANGUAGE sql STABLE;

-- Function to get slot details with lane+slot_letter combination
CREATE FUNCTION get_slot_with_lane(p_archer_id UUID, p_session_id UUID)
RETURNS TABLE (
    slot_id UUID,
    target_id UUID,
    archer_id UUID,
    session_id UUID,
    face_type FACE_TYPE,
    slot_letter CHAR(1),
    created_at TIMESTAMPTZ,
    is_shooting BOOLEAN,
    bowstyle BOWSTYLE_TYPE,
    draw_weight FLOAT,
    club_id UUID,
    slot VARCHAR(4)
) AS
$$
SELECT
    slot.slot_id,
    slot.target_id,
    slot.archer_id,
    slot.session_id,
    slot.face_type,
    slot.slot_letter,
    slot.created_at,
    slot.is_shooting,
    slot.bowstyle,
    slot.draw_weight,
    slot.club_id,
    CONCAT(target.lane, slot.slot_letter) AS slot
FROM slot
INNER JOIN target
    ON slot.target_id = target.target_id
WHERE slot.archer_id = p_archer_id
    AND slot.session_id = p_session_id
LIMIT 1
$$ LANGUAGE sql STABLE;
