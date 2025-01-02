CREATE TABLE IF NOT EXISTS target (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lane_id UUID NOT NULL,
    x_coordinate REAL,
    y_coordinate REAL,
    radius REAL [],
    points INT [],
    height REAL,
    name VARCHAR(255),
    lane_id UUID REFERENCES lane (id) ON DELETE CASCADE,
    CHECK (array_length(radius, 1) = array_length(points, 1)),
    UNIQUE (lane_id, name)
    FOREIGN KEY (lane_id) REFERENCES lane(id) ON DELETE CASCADE
);
