/*

Most query will be done using archer_id and tournament_id

Flow of data:
1. N/A
2. A tournament will be created if none exists.
3. Each archer will register in the database if they are not already registered.
   When registering the archer will register all their arrows.
4. Each archer will select a tournament and a lane.
5. After all the archer participating in the tournament have selected a lane, the tournament will
start.
6. The archers will shoot at the targets.
7. The shots will be recorded in the database.
8. The tournament ends when all archers have finished shooting.
*/

CREATE DATABASE archery;
COMMENT ON DATABASE archery IS 'the database will track shots by one or more archers within a
tournament.';
\c archery;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------
--                                   Functions
---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------


-- function calculate_points()
CREATE OR REPLACE FUNCTION calculate_points(
    shot_x FLOAT,
    shot_y FLOAT,
    target_x FLOAT,
    target_y FLOAT,
    radius FLOAT [],
    points INT []
) RETURNS INT AS $$
DECLARE
    distance FLOAT;
    ring_index INT;
    shot_points INT;
BEGIN
    -- Calculate the distance from the shot to the center of the target
    distance := SQRT(POWER(shot_x - target_x, 2) + POWER(shot_y - target_y, 2));
    
    -- Find the appropriate ring
    FOR ring_index IN 1 .. array_length(radius, 1) LOOP
        IF distance <= radius[ring_index] THEN
            shot_points := points[ring_index];
            RETURN shot_points;
        END IF;
    END LOOP;
    
    -- If the shot is outside all rings, return 0 points
    RETURN 0;
END;
$$ LANGUAGE plpgsql;

-- function check_lane_number()
CREATE OR REPLACE FUNCTION check_lane_number() RETURNS TRIGGER AS $$
DECLARE
    max_lanes INT;
BEGIN
    SELECT number_of_lanes
    INTO max_lanes
    FROM tournaments
    WHERE id = NEW.tournament_id;

    IF NEW.lane_number < 1 OR NEW.lane_number > max_lanes THEN
        RAISE EXCEPTION 'lane_number out of range';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- function check_shot_coordinates()
CREATE OR REPLACE FUNCTION check_shot_coordinates() RETURNS TRIGGER AS $$
DECLARE
    max_x FLOAT;
    max_y FLOAT;
BEGIN
    SELECT l.max_x_coordinate, l.max_y_coordinate
    INTO max_x, max_y
    FROM lanes AS l
    INNER JOIN targets AS t ON l.id = t.lane_id
    WHERE t.id = NEW.target_id;

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

---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------
--                                   Tables
---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------

-- Tournaments table
CREATE TABLE IF NOT EXISTS tournaments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    number_of_lanes INT NOT NULL,
    is_opened BOOLEAN DEFAULT FALSE,
    CONSTRAINT open_tournament_no_end_time CHECK (
        (is_opened = TRUE AND end_time IS NULL) OR (is_opened = FALSE)
    ),
    CONSTRAINT closed_tournament_with_end_time CHECK (
        (is_opened = FALSE AND end_time IS NOT NULL) OR (is_opened = TRUE)
    )
);

-- Archers table
CREATE TABLE IF NOT EXISTS archers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    genre ARCHER_GENRE NOT NULL,
    type_of_archer ARCHER_TYPE NOT NULL,
    bow_poundage FLOAT NOT NULL,
    created_at TIMESTAMP DEFAULT current_timestamp
);

-- Arrows table
CREATE TABLE IF NOT EXISTS arrows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    weight FLOAT DEFAULT 0.0,
    diameter FLOAT DEFAULT 0.0,
    spine FLOAT DEFAULT 0.0,
    length FLOAT NOT NULL,
    human_readable_name VARCHAR(255) NOT NULL,
    archer_id UUID REFERENCES archers (id) ON DELETE CASCADE,
    label_position FLOAT NOT NULL,
    UNIQUE (archer_id, human_readable_name)
);

-- Lanes table
CREATE TABLE IF NOT EXISTS lanes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID REFERENCES tournaments (id) ON DELETE CASCADE,
    lane_number INT NOT NULL,
    number_of_archers INT NOT NULL
);

-- Targets table
CREATE TABLE IF NOT EXISTS targets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    x_coordinate FLOAT,
    y_coordinate FLOAT,
    radius FLOAT [],
    points INT [],
    height FLOAT,
    lane_sensor_id UUID NOT NULL,
    max_x_coordinate FLOAT,
    max_y_coordinate FLOAT,
    human_readable_name VARCHAR(255),
    lane_id UUID REFERENCES lanes (id) ON DELETE CASCADE,
    archer_id UUID REFERENCES archers (id) ON DELETE CASCADE,
    CHECK (array_length(radius, 1) = array_length(points, 1)),
    UNIQUE (lane_id, human_readable_name)
);


-- Tournament registration table
CREATE TABLE IF NOT EXISTS registration (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    archer_id UUID REFERENCES archers (id) ON DELETE CASCADE,
    lane_id UUID REFERENCES lanes (id) ON DELETE CASCADE,
    tournament_id UUID REFERENCES tournaments (id) ON DELETE CASCADE,
    UNIQUE (archer_id, tournament_id),
    UNIQUE (archer_id, lane_id, tournament_id)
);

-- Shots table
CREATE TABLE IF NOT EXISTS shots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lane_sensor_id UUID NOT NULL,
    bow_sensor_id UUID NOT NULL,
    arrow_engage_time INT,
    arrow_disengage_time INT,
    arrow_landing_time INT,
    x_coordinate FLOAT,
    y_coordinate FLOAT,
    pull_length FLOAT,
    distance FLOAT,
    arrow_id UUID REFERENCES arrows (id) ON DELETE CASCADE,
);

---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------
--                                   Triggers
---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------


-- Create a trigger that calls the trigger function before inserting or updating rows in the lanes
-- table
CREATE TRIGGER check_lane_number_trigger
BEFORE INSERT OR UPDATE ON lanes
FOR EACH ROW
EXECUTE FUNCTION check_lane_number();

-- Create a trigger that calls the trigger function before inserting or updating rows in the shots
-- table
CREATE TRIGGER check_shot_coordinates_trigger
BEFORE INSERT OR UPDATE ON shots
FOR EACH ROW
EXECUTE FUNCTION check_shot_coordinates();


---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------
--                                   Views
---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------

-- View for shots with calculated points
CREATE OR REPLACE VIEW shots_view AS
SELECT
    s.distance,
    (s.arrow_landing_time - s.arrow_disengage_time) AS flight_time,
    s.distance / (s.arrow_landing_time - s.arrow_disengage_time) AS speed,
    calculate_points(
        s.x_coordinate,
        s.y_coordinate,
        t.x_coordinate,
        t.y_coordinate,
        t.radius,
        t.points
    ) AS shot_points
FROM shots AS s
INNER JOIN
    targets AS t ON s.target_id = t.id;

CREATE OR REPLACE VIEW raw_data AS
SELECT
    s.arrow_engage_time,
    s.arrow_disengage_time,
    s.arrow_landing_time,
    s.x_coordinate AS shot_x_coordinate,
    s.y_coordinate AS shot_y_coordinate,
    s.pull_length,
    s.distance,
    s.target_id,
    t.x_coordinate AS target_x_coordinate,
    t.y_coordinate AS target_y_coordinate,
    t.radius,
    t.points,
    t.height AS target_height,
    t.human_readable_name AS target_name
FROM shots AS s
INNER JOIN targets AS t ON s.target_id = t.id;

---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------
--                                   Comments
---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------


COMMENT ON TABLE tournaments IS 'Each tournament must have a start time, an end time, and a number
of lanes.
Each tournament can have multiple archers.';
COMMENT ON TABLE archers IS 'Each archer will have many arrows.
Each archer will have one or more targets placed in a lane.
Each archer must be registered in only one tournament.
Each archer will shoot in only one lane per tournament.
Each archer can participate in only one tournament at the time.
Each archer will fire many shots per tournament.
Each archer can exists in the system without any arrows but an arrow must have an archer.
Each archer can exists in the system without any tournament but a tournament must have at least one
archer.
Multiple archers can shoot in the same lane but at different times.';
COMMENT ON TABLE lanes IS 'Each lane can be used by only one archer at a time.
Each lane will have a limit of number of archers.
Each lane will be available only for one tournament at a time.';
COMMENT ON TABLE targets IS 'Each target will be set to a specific lane.';
COMMENT ON COLUMN lanes.lane_number IS 'This will be a number from 1 to the number of lanes set in
the tournament.';
COMMENT ON COLUMN arrows.id IS 'this will be extracted from a NFC tag.';
COMMENT ON COLUMN arrows.weight IS 'The weight of the arrow in grains and it is optional because
not everyone will have a scale.';
COMMENT ON COLUMN arrows.diameter IS 'The diameter of the arrow in millimeters and it is optional
because not everyone will have a caliper.';
COMMENT ON COLUMN arrows.length IS 'The length of the arrow in millimeters and it is required
because it will be used to do other calculations.';
COMMENT ON COLUMN arrows.human_readable_name IS 'This is a label to identify the arrow easily.';
COMMENT ON COLUMN targets.id IS 'human readable name is used to generate the uuid.';
COMMENT ON COLUMN targets.human_readable_name IS 'This is a label to identify the target easily.';
COMMENT ON COLUMN targets.height IS 'This is the height from the ground to the center of the
target. Height is in centimeters.';
COMMENT ON COLUMN targets.radius IS 'This is an array of x axis values that represent the radius 
of the different rings in the target. Coordinates are in centimeters.';
COMMENT ON COLUMN targets.points IS 'This is an array of points each ring in the target.';
COMMENT ON COLUMN targets.max_x_coordinate IS 'the maximum x coordinate that the
sensor will read. Coordinates are in centimeters.';
COMMENT ON COLUMN targets.max_y_coordinate IS 'the maximum y coordinate that the
sensor will read. Coordinates are in centimeters.';
COMMENT ON COLUMN targets.x_coordinate IS 'x coordinate of the center of the target. Coordinates
are in centimeters.';
COMMENT ON COLUMN targets.y_coordinate IS 'y coordinate of the center of the target. Coordinates
are in centimeters.';
COMMENT ON COLUMN shots.arrow_engage_time IS 'Time in seconds when the arrow is engaged in the bow.
';
COMMENT ON COLUMN shots.arrow_disengage_time IS 'Time in seconds when the arrow leave the bow.';
COMMENT ON COLUMN shots.arrow_landing_time IS 'Time in seconds when the arrow hit the target.';
COMMENT ON COLUMN shots.pull_length IS 'Distance from the knocking point to the sensor in the bow
at full draw. The measurement is in centimeters.';
COMMENT ON COLUMN shots.distance IS 'The distance represented in meters from the archer to the
target.';
