-- Shots table
--
-- Tracks each shot made by an archer during a tournament.
-- A shot records the target, arrow, and precise coordinates.

CREATE TABLE IF NOT EXISTS shot (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    arrow_engage_time BIGINT,
    arrow_disengage_time BIGINT,
    arrow_landing_time BIGINT,
    x_coordinate REAL,
    y_coordinate REAL,
    pull_length REAL,
    target_id UUID REFERENCES target (id) ON DELETE CASCADE,
    arrow_id UUID REFERENCES arrow (id) ON DELETE CASCADE,
);
-- Through views, we can calculate score, aiming time, grouping size, speed

-- Relationships:
-- - M:1 with target (each shot hits one target)
-- - M:1 with arrow (each shot uses one arrow)
-- - Archer can be inferred from arrow -> archer relationship
COMMENT ON COLUMN shot.arrow_engage_time IS 'Time in seconds when the arrow is engaged in the bow.
';
COMMENT ON COLUMN shot.arrow_disengage_time IS 'Time in seconds when the arrow leave the bow.';
COMMENT ON COLUMN shot.arrow_landing_time IS 'Time in seconds when the arrow hit the target.';
COMMENT ON COLUMN shot.pull_length IS 'Distance from the knocking point to the sensor in the bow
at full draw. The measurement is in centimeters.';
COMMENT ON COLUMN shot.distance IS 'The distance represented in meters from the archer to the
target.';

-- function check_shot_coordinates()
CREATE OR REPLACE FUNCTION check_shot_coordinates() RETURNS TRIGGER AS $$
DECLARE
    max_x REAL;
    max_y REAL;
BEGIN
    SELECT max_x_coordinate, max_y_coordinate
    INTO max_x, max_y
    FROM lane
    WHERE id = NEW.lane_id;

    IF NEW.x_coordinate < 0.0 OR NEW.x_coordinate > max_x THEN
        RAISE EXCEPTION 'x_coordinate out of range';
    END IF;

    IF NEW.y_coordinate < 0.0 OR NEW.y_coordinate > max_y THEN
        RAISE EXCEPTION 'y_coordinate out of range';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TYPE archer_type AS ENUM ('compound', 'traditional', 'barebow', 'olympic');
CREATE TYPE archer_genre AS ENUM ('male', 'female', 'no-answered');


-- Create a trigger that calls the trigger function before inserting or updating rows in the shot
-- table
CREATE TRIGGER check_shot_coordinates_trigger
BEFORE INSERT OR UPDATE ON shot
FOR EACH ROW
EXECUTE FUNCTION check_shot_coordinates();
