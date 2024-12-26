CREATE TABLE IF NOT EXISTS shooting (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_track_id UUID NOT NULL,
    arrow_id UUID NOT NULL,
    arrow_engage_time BIGINT,
    pull_length REAL,
    arrow_disengage_time BIGINT,
    arrow_landing_time BIGINT,
    x_coordinate REAL,
    y_coordinate REAL,
    FOREIGN KEY (arrow_id) REFERENCES arrow(id) ON DELETE CASCADE,
    FOREIGN KEY (target_track_id) REFERENCES target_track(id) ON DELETE CASCADE
);

COMMENT ON COLUMN shooting.target_track_id IS 'it will be provided by the Raspberry Pi';
COMMENT ON COLUMN shooting.arrow_id IS 'it will be read by the bow sensor and the target sensor';
COMMENT ON COLUMN shooting.arrow_engage_time IS 'it will be read by the bow sensor';
COMMENT ON COLUMN shooting.pull_length IS 'it will be read by the bow sensor';
COMMENT ON COLUMN shooting.arrow_disengage_time IS 'it will be read by the bow sensor';
COMMENT ON COLUMN shooting.arrow_landing_time IS 'it will be read by the target sensor';
COMMENT ON COLUMN shooting.x_coordinate IS 'it will be read by the target sensor';
COMMENT ON COLUMN shooting.y_coordinate IS 'it will be read by the target sensor';
