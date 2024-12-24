-- Tournaments table
--
-- Represents tournaments that archers can register for.
-- A tournament can host multiple lanes and registrations.

CREATE TABLE IF NOT EXISTS tournament (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    number_of_lanes INT NOT NULL,
    is_open BOOLEAN DEFAULT FALSE,
    place VARCHAR(255) DEFAULT 'Unknown',
    CONSTRAINT open_tournament_no_end_time CHECK (
        (is_open = TRUE AND end_time IS NULL) OR (is_open = FALSE)
    ),
    CONSTRAINT closed_tournament_with_end_time CHECK (
        (is_open = FALSE AND end_time IS NOT NULL) OR (is_open = TRUE)
    )
);


-- Add a unique index to ensure only one tournament can be open at a time
CREATE UNIQUE INDEX only_one_open_tournament ON tournament (is_open) WHERE is_open = TRUE;

COMMENT ON TABLE tournament IS 'An open tournament has is_open set to true and has no end_time.
A closed tournament has is_open set to false and has an end_time.
There can be only one open tournament at a time.';
-- Relationships:
-- - 1:M with lane (each tournament has multiple lanes)
-- - M:M with archer through registration
