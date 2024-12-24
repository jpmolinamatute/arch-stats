-- Targets table
--
-- Represents targets placed within lanes.
-- Targets can be hit by multiple arrows during the tournament.

CREATE TABLE IF NOT EXISTS target (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    x_coordinate REAL,
    y_coordinate REAL,
    radius REAL [],
    points INT [],
    height REAL,
    human_readable_name VARCHAR(255),
    lane_id UUID REFERENCES lane (id) ON DELETE CASCADE,
    CHECK (array_length(radius, 1) = array_length(points, 1)),
    UNIQUE (lane_id, human_readable_name)
);

-- Relationships:
-- - M:1 with lane (each target belongs to one lane)
-- - 1:M with shot (each target can receive multiple shots)
COMMENT ON TABLE target IS 'Each target will be set to a specific lane.
Each target will be used by only one archer.
An archer will have one or more target.';
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
