-- Lanes table
--
-- Represents lanes available during tournaments.
-- Each lane is part of a tournament and can have multiple targets.

CREATE TABLE IF NOT EXISTS lane (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID REFERENCES tournament (id) ON DELETE CASCADE,
    target_id UUID REFERENCES target (id) ON DELETE CASCADE,
    lane_number INT NOT NULL,
    distance REAL,
    max_x_coordinate REAL,
    max_y_coordinate REAL,
    number_of_archers INT NOT NULL
);

-- Relationships:
-- * M:1 with tournament (each lane belongs to one tournament)
-- * 1:M with target (each lane can have multiple targets)

COMMENT ON TABLE lane IS 'Each lane can be used by only one archer at a time.
Multiple archers can shoot in the same lane but at different times.
Each lane will have a limit of number of archers.
Each lane will be available only for one tournament at a time.';
COMMENT ON COLUMN lane.lane_number IS 'This will be a number from 1 to the number of lanes set in
the tournament.';
COMMENT ON COLUMN lane.max_x_coordinate IS 'the maximum x coordinate that the
sensor will read. Coordinates are in centimeters.';
COMMENT ON COLUMN lane.max_y_coordinate IS 'the maximum y coordinate that the
sensor will read. Coordinates are in centimeters.';

-- function check_lane_number()
CREATE OR REPLACE FUNCTION check_lane_number() RETURNS TRIGGER AS $$
DECLARE
    max_lanes INT;
BEGIN
    SELECT number_of_lanes
    INTO max_lanes
    FROM tournament
    WHERE id = NEW.tournament_id;

    IF NEW.lane_number < 1 OR NEW.lane_number > max_lanes THEN
        RAISE EXCEPTION 'lane_number out of range';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger that calls the trigger function before inserting or updating rows in the lanes
-- table
CREATE TRIGGER check_lane_number_trigger
BEFORE INSERT OR UPDATE ON lane
FOR EACH ROW
EXECUTE FUNCTION check_lane_number();
