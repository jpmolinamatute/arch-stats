CREATE TABLE IF NOT EXISTS tournament (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
