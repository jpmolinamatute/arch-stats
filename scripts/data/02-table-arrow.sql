\c "arch-stats"

CREATE TABLE IF NOT EXISTS arrow (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    archer_id UUID NOT NULL,
    weight REAL DEFAULT 0.0,
    diameter REAL DEFAULT 0.0,
    spine REAL DEFAULT 0.0,
    length REAL NOT NULL,
    name VARCHAR(255) NOT NULL,
    label_position REAL NOT NULL,
    FOREIGN KEY (archer_id) REFERENCES archer (id) ON DELETE CASCADE,
    UNIQUE (archer_id, name)
);
COMMENT ON COLUMN arrow.archer_id IS 'it will be provided by the WebUI';
