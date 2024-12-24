-- Arrows table
--
-- Represents arrows that archers own and use during tournaments.
-- Each arrow is unique to an archer and can be used for multiple shots.

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
-- Relationships:
-- - M:1 with archer (each arrow belongs to one archer)
-- - 1:M with shot (each arrow can be used for multiple shots)
COMMENT ON TABLE archer IS 'Each arrow will be owned by only one archer.
An archer will have many arrows.';

COMMENT ON COLUMN arrow.id IS 'this will be extracted from a NFC tag.';
COMMENT ON COLUMN arrow.weight IS 'The weight of the arrow in grains and it is optional because
not everyone will have a scale.';
COMMENT ON COLUMN arrow.diameter IS 'The diameter of the arrow in millimeters and it is optional
because not everyone will have a caliper.';
COMMENT ON COLUMN arrow.length IS 'The length of the arrow in millimeters and it is required
because it will be used to do other calculations.';
COMMENT ON COLUMN arrow.human_readable_name IS 'This is a label to identify the arrow easily.';
