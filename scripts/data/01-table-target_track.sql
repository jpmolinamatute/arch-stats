CREATE TABLE IF NOT EXISTS target_track (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
);

COMMENT ON TABLE target_track.id IS 'This is a UUID stored in a file that identifies a Raspberry Pi';
