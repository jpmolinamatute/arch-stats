/*

Most query will be done using archer_id and/or tournament_id

Flow of data:
1. N/A
2. A tournament will be created if none exists.
3. Each archer will register in the database if they are not already registered.
   When registering the archer will register all their arrow.
4. Each archer will select a tournament and a lane.
5. After all the archer participating in the tournament have selected a lane, the tournament will
start.
6. The archers will shoot at the target.
7. The shot will be recorded in the database.
8. The tournament ends when all archers have finished shooting.
*/

-- CREATE DATABASE archery;
-- COMMENT ON DATABASE archery IS 'the database will track shot by one or more archers within a
-- tournament.';
-- \c archery;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------
--                                   Functions
---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------


-- function calculate_points()
CREATE OR REPLACE FUNCTION calculate_points(
    shot_x REAL,
    shot_y REAL,
    target_x REAL,
    target_y REAL,
    radius REAL [],
    points INT []
) RETURNS INT AS $$
DECLARE
    distance REAL;
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
    FROM tournament
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

---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------
--                                   Tables
---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------

-- Tournaments table
CREATE TABLE IF NOT EXISTS tournament (
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
CREATE TABLE IF NOT EXISTS archer (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    genre ARCHER_GENRE NOT NULL,
    type_of_archer ARCHER_TYPE NOT NULL,
    bow_poundage REAL NOT NULL,
    created_at TIMESTAMP DEFAULT current_timestamp
);

-- Arrows table
CREATE TABLE IF NOT EXISTS arrow (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    weight REAL DEFAULT 0.0,
    diameter REAL DEFAULT 0.0,
    spine REAL DEFAULT 0.0,
    length REAL NOT NULL,
    human_readable_name VARCHAR(255) NOT NULL,
    archer_id UUID REFERENCES archer (id) ON DELETE CASCADE,
    label_position REAL NOT NULL,
    UNIQUE (archer_id, human_readable_name)
);

-- Lanes table
CREATE TABLE IF NOT EXISTS lane (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID REFERENCES tournament (id) ON DELETE CASCADE,
    lane_number INT NOT NULL,
    max_x_coordinate REAL,
    max_y_coordinate REAL,
    number_of_archers INT NOT NULL
);

-- Targets table
CREATE TABLE IF NOT EXISTS target (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    x_coordinate REAL,
    y_coordinate REAL,
    radius REAL [],
    points INT [],
    height REAL,
    human_readable_name VARCHAR(255),
    lane_id UUID REFERENCES lane (id) ON DELETE CASCADE,
    archer_id UUID REFERENCES archer (id) ON DELETE CASCADE,
    CHECK (array_length(radius, 1) = array_length(points, 1)),
    UNIQUE (lane_id, human_readable_name)
);


-- Tournament registration table
CREATE TABLE IF NOT EXISTS registration (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    archer_id UUID REFERENCES archer (id) ON DELETE CASCADE,
    lane_id UUID REFERENCES lane (id) ON DELETE CASCADE,
    tournament_id UUID REFERENCES tournament (id) ON DELETE CASCADE,
    UNIQUE (archer_id, tournament_id),
    UNIQUE (archer_id, lane_id, tournament_id)
);

-- Shots table
CREATE TABLE IF NOT EXISTS shot (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    arrow_engage_time BIGINT,
    arrow_disengage_time BIGINT,
    arrow_landing_time BIGINT,
    x_coordinate REAL,
    y_coordinate REAL,
    pull_length REAL,
    distance REAL,
    lane_id UUID REFERENCES lane (id) ON DELETE CASCADE,
    arrow_id UUID REFERENCES arrow (id) ON DELETE CASCADE
);

---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------
--                                   Triggers
---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------


-- Create a trigger that calls the trigger function before inserting or updating rows in the lanes
-- table
CREATE TRIGGER check_lane_number_trigger
BEFORE INSERT OR UPDATE ON lane
FOR EACH ROW
EXECUTE FUNCTION check_lane_number();

-- Create a trigger that calls the trigger function before inserting or updating rows in the shot
-- table
CREATE TRIGGER check_shot_coordinates_trigger
BEFORE INSERT OR UPDATE ON shot
FOR EACH ROW
EXECUTE FUNCTION check_shot_coordinates();


---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------
--                                   Views
---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------

-- View for shot with calculated points
-- CREATE OR REPLACE VIEW shots_view AS
-- SELECT
--     s.distance,
--     (s.arrow_landing_time - s.arrow_disengage_time) AS flight_time,
--     s.distance / (s.arrow_landing_time - s.arrow_disengage_time) AS speed,
--     calculate_points(
--         s.x_coordinate,
--         s.y_coordinate,
--         t.x_coordinate,
--         t.y_coordinate,
--         t.radius,
--         t.points
--     ) AS shot_points
-- FROM shot AS s
-- INNER JOIN
--     target AS t ON s.target_id = t.id;

-- CREATE OR REPLACE VIEW raw_data AS
-- SELECT
--     s.arrow_engage_time,
--     s.arrow_disengage_time,
--     s.arrow_landing_time,
--     s.x_coordinate AS shot_x_coordinate,
--     s.y_coordinate AS shot_y_coordinate,
--     s.pull_length,
--     s.distance,
--     s.target_id,
--     t.x_coordinate AS target_x_coordinate,
--     t.y_coordinate AS target_y_coordinate,
--     t.radius,
--     t.points,
--     t.height AS target_height,
--     t.human_readable_name AS target_name
-- FROM shot AS s
-- INNER JOIN target AS t ON s.target_id = t.id;

---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------
--                                   Comments
---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------


COMMENT ON TABLE tournament IS 'Each tournament must have a start time, an end time, and a number
of lanes.
Each tournament can have multiple archers.';
COMMENT ON TABLE archer IS 'Each archer will have many arrows.
Each archer will have one or more target placed in a lane.
Each archer must be registered in only one tournament.
Each archer will shoot in only one lane per tournament.
Each archer can participate in only one tournament at the time.
Each archer will fire many shot per tournament.
Each archer can exists in the system without any arrows but an arrow must have an archer.
Each archer can exists in the system without any tournament but a tournament must have at least one
archer.
Multiple archers can shoot in the same lane but at different times.';
COMMENT ON TABLE lane IS 'Each lane can be used by only one archer at a time.
Each lane will have a limit of number of archers.
Each lane will be available only for one tournament at a time.';
COMMENT ON TABLE target IS 'Each target will be set to a specific lane.';
COMMENT ON COLUMN lane.lane_number IS 'This will be a number from 1 to the number of lanes set in
the tournament.';
COMMENT ON COLUMN lane.max_x_coordinate IS 'the maximum x coordinate that the
sensor will read. Coordinates are in centimeters.';
COMMENT ON COLUMN lane.max_y_coordinate IS 'the maximum y coordinate that the
sensor will read. Coordinates are in centimeters.';
COMMENT ON COLUMN arrow.id IS 'this will be extracted from a NFC tag.';
COMMENT ON COLUMN arrow.weight IS 'The weight of the arrow in grains and it is optional because
not everyone will have a scale.';
COMMENT ON COLUMN arrow.diameter IS 'The diameter of the arrow in millimeters and it is optional
because not everyone will have a caliper.';
COMMENT ON COLUMN arrow.length IS 'The length of the arrow in millimeters and it is required
because it will be used to do other calculations.';
COMMENT ON COLUMN arrow.human_readable_name IS 'This is a label to identify the arrow easily.';
COMMENT ON COLUMN target.id IS 'human readable name is used to generate the uuid.';
COMMENT ON COLUMN target.human_readable_name IS 'This is a label to identify the target easily.';
COMMENT ON COLUMN target.height IS 'This is the height from the ground to the center of the
target. Height is in centimeters.';
COMMENT ON COLUMN target.radius IS 'This is an array of x axis values that represent the radius 
of the different rings in the target. Coordinates are in centimeters.';
COMMENT ON COLUMN target.points IS 'This is an array of points each ring in the target.';
COMMENT ON COLUMN target.x_coordinate IS 'x coordinate of the center of the target. Coordinates
are in centimeters.';
COMMENT ON COLUMN target.y_coordinate IS 'y coordinate of the center of the target. Coordinates
are in centimeters.';
COMMENT ON COLUMN shot.arrow_engage_time IS 'Time in seconds when the arrow is engaged in the bow.
';
COMMENT ON COLUMN shot.arrow_disengage_time IS 'Time in seconds when the arrow leave the bow.';
COMMENT ON COLUMN shot.arrow_landing_time IS 'Time in seconds when the arrow hit the target.';
COMMENT ON COLUMN shot.pull_length IS 'Distance from the knocking point to the sensor in the bow
at full draw. The measurement is in centimeters.';
COMMENT ON COLUMN shot.distance IS 'The distance represented in meters from the archer to the
target.';

---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------
--                                   Dummy Data
---------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------


-- Insert into tournament table
INSERT INTO tournament (
    id, start_time, end_time, number_of_lanes, is_opened
) VALUES (
    '550e8400-e29b-41d4-a716-446655440000', '2023-01-01 09:00:00', NULL, 10, TRUE
);

-- Insert into archer table
INSERT INTO archer (
    id, first_name, last_name, password, email, genre, type_of_archer, bow_poundage, created_at
) VALUES (
    '550e8400-e29b-41d4-a716-446655440001',
    'John',
    'Doe',
    'password123',
    'john.doe@example.com',
    'male',
    'olympic',
    40.5,
    current_timestamp
);

-- Insert into lane table
INSERT INTO lane (
    id, tournament_id, lane_number, number_of_archers,max_x_coordinate, max_y_coordinate
) VALUES (
    '550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440000', 1, 4, 100.0, 200.0
);

-- Insert into target table
INSERT INTO target (
    id,
    x_coordinate,
    y_coordinate,
    radius,
    points,
    height,
    human_readable_name,
    lane_id,
    archer_id
) VALUES (
    '550e8400-e29b-41d4-a716-446655440003',
    1.23,
    4.56,
    ARRAY[10.0, 20.0, 30.0],
    ARRAY[10, 9, 8],
    1.5,
    'Target 1',
    '550e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440001'
), (
    '550e8400-e29b-41d4-a716-446655440005',
    2.34,
    5.67,
    ARRAY[15.0, 25.0, 35.0],
    ARRAY[10, 9, 8],
    1.6,
    'Target 2',
    '550e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440001'
), (
    '550e8400-e29b-41d4-a716-446655440007',
    3.45,
    6.78,
    ARRAY[20.0, 30.0, 40.0],
    ARRAY[10, 9, 8],
    1.7,
    'Target 3',
    '550e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440001'
);



-- Insert into arrow table
INSERT INTO arrow (
    id, weight, diameter, spine, length, human_readable_name, archer_id, label_position
) VALUES (
    '4c19ead8-9b49-4876-978c-f22b5ec5edbf',
    30.0,
    5.0,
    400.0,
    28.0,
    'Arrow 1',
    '550e8400-e29b-41d4-a716-446655440001',
    10.0
), (
    '854eef0f-badb-414d-a580-2b4aa5df79eb',
    32.0,
    5.2,
    420.0,
    28.5,
    'Arrow 2',
    '550e8400-e29b-41d4-a716-446655440001',
    10.5
), (
    '7122a4ad-eece-46eb-9424-f680b7b12d4e',
    34.0,
    5.4,
    440.0,
    29.0,
    'Arrow 3',
    '550e8400-e29b-41d4-a716-446655440001',
    11.0
), (
    '1cd1d238-cf78-40ed-a32b-5f8208473183',
    36.0,
    5.6,
    460.0,
    29.5,
    'Arrow 4',
    '550e8400-e29b-41d4-a716-446655440001',
    11.5
), (
    '55af4551-bc88-4602-8a43-4837c16cdb81',
    38.0,
    5.8,
    480.0,
    30.0,
    'Arrow 5',
    '550e8400-e29b-41d4-a716-446655440001',
    12.0
), (
    'e1388bf0-df48-4a2a-867e-e882348c6c3d',
    40.0,
    6.0,
    500.0,
    30.5,
    'Arrow 6',
    '550e8400-e29b-41d4-a716-446655440001',
    12.5
), (
    'ed8f8586-1137-4192-bcc7-c0607d7c2df4',
    42.0,
    6.2,
    520.0,
    31.0,
    'Arrow 7',
    '550e8400-e29b-41d4-a716-446655440001',
    13.0
), (
    '1f91e8e0-77dc-4621-8108-f000b0e690a5',
    44.0,
    6.4,
    540.0,
    31.5,
    'Arrow 8',
    '550e8400-e29b-41d4-a716-446655440001',
    13.5
), (
    'edd2e424-9d62-4457-ad75-08b3de4f4b72',
    46.0,
    6.6,
    560.0,
    32.0,
    'Arrow 9',
    '550e8400-e29b-41d4-a716-446655440001',
    14.0
), (
    'd5e554dc-face-47b3-adca-d6354fbe6b1e',
    48.0,
    6.8,
    580.0,
    32.5,
    'Arrow 10',
    '550e8400-e29b-41d4-a716-446655440001',
    14.5
), (
    '6c34d6b9-c778-4e78-9ad8-9ade92890659',
    50.0,
    7.0,
    600.0,
    33.0,
    'Arrow 11',
    '550e8400-e29b-41d4-a716-446655440001',
    15.0
), (
    '6a1a0923-1bc2-40e0-87cb-5d342ea83643',
    52.0,
    7.2,
    620.0,
    33.5,
    'Arrow 12',
    '550e8400-e29b-41d4-a716-446655440001',
    15.5
);

-- Insert into registration table
INSERT INTO registration (
    id, archer_id, lane_id, tournament_id
) VALUES (
    uuid_generate_v4(),
    '550e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440000'
);
