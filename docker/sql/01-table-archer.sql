-- Archers table
--
-- Represents individual archers participating in tournaments.
-- Each archer can own multiple arrows and register for multiple tournaments.

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
-- Relationships:
-- * 1:M with arrow (each archer owns multiple arrows)
-- * M:M with tournament through registration
-- * 1:M with shooting (each archer can fire multiple shots)
