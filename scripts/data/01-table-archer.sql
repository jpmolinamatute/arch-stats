\c "arch-stats"

CREATE TYPE archer_type AS ENUM ('traditional', 'compound', 'olympic', 'barebow');

CREATE TABLE IF NOT EXISTS archer (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    type_of_archer ARCHER_TYPE NOT NULL,
    bow_weight REAL DEFAULT 0.0,
    draw_length REAL DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT current_timestamp
);
