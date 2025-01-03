\c "arch-stats"

CREATE TABLE IF NOT EXISTS target_track (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

COMMENT ON COLUMN target_track.id IS 'This is a UUID stored in a file that identifies a Raspberry Pi';
