\c "arch-stats"

CREATE TABLE IF NOT EXISTS lane (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL,
    target_track_id UUID NOT NULL,
    lane_number INT NOT NULL,
    max_x_coordinate REAL,
    max_y_coordinate REAL,
    UNIQUE (tournament_id, target_track_id),
    FOREIGN KEY (target_track_id) REFERENCES target_track (id) ON DELETE CASCADE,
    FOREIGN KEY (tournament_id) REFERENCES tournament (id) ON DELETE CASCADE
);

COMMENT ON COLUMN lane.target_track_id IS 'it will be provided by the Raspberry Pi';
COMMENT ON COLUMN lane.max_x_coordinate IS 'it will be read by the target sensor';
COMMENT ON COLUMN lane.max_y_coordinate IS 'it will be read by the target sensor';
COMMENT ON COLUMN lane.lane_number IS 'it will be provided by Tournament Organizer';
COMMENT ON COLUMN lane.tournament_id IS 'it will be provided by the WebUI';


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
