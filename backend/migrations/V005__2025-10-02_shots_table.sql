CREATE TABLE shot (
    shot_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slot_id UUID NOT NULL REFERENCES slot (slot_id) ON DELETE CASCADE,
    x_mm DOUBLE PRECISION,
    y_mm DOUBLE PRECISION,
    score INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT shot_score_nonnegative CHECK (score >= 0),
    -- Either all three values are NULL (no shot recorded), or all are present
    CONSTRAINT shot_coords_score_all_or_none CHECK (
        (x_mm IS NULL AND y_mm IS NULL AND score IS NULL)
        OR (x_mm IS NOT NULL AND y_mm IS NOT NULL AND score IS NOT NULL)
    )
);

CREATE INDEX idx_shot_slot_id ON shot (slot_id);

-- Down below will create all custom functions related to live shooting session

CREATE OR REPLACE FUNCTION get_live_shots_by_archer_and_session(
    param_archer_id UUID,
    param_session_id UUID
)
RETURNS TABLE (
    shot_id UUID,
    slot_id UUID,
    x_mm DOUBLE PRECISION,
    y_mm DOUBLE PRECISION,
    score INTEGER,
    shot_created_at TIMESTAMPTZ,
    archer_id UUID,
    session_id UUID,
    face_type FACE_TYPE,
    bowstyle BOWSTYLE_TYPE,
    draw_weight FLOAT
)
LANGUAGE sql
STABLE
AS
$$
SELECT
    shot.shot_id,
    shot.slot_id,
    shot.x_mm,
    shot.y_mm,
    shot.score,
    shot.created_at AS shot_created_at,
    -- slot fields "duplicated" on every shot row
    slot.archer_id,
    slot.session_id,
    slot.face_type,
    slot.bowstyle,
    slot.draw_weight
FROM shot
INNER JOIN slot
    ON shot.slot_id = slot.slot_id
WHERE slot.archer_id = param_archer_id
    AND slot.session_id = param_session_id;
$$;
