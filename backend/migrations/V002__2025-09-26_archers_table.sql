CREATE TYPE bowstyle_type AS ENUM (
    'recurve',
    'compound',
    'barebow',
    'longbow'
);

CREATE TYPE gender_type AS ENUM (
    'male',
    'female',
    'non_binary',
    'other',
    'unspecified'
);

-- Table to store archer'
-- basic information (name, email, date_of_birth, gender)
-- and authentication info (google_subject, google_picture_url)
-- It also stores the current bowstyle, draw_weight and current_club_id
CREATE TABLE IF NOT EXISTS archer (
    archer_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(100) NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL CHECK (date_of_birth <= current_date),
    gender GENDER_TYPE NOT NULL,
    bowstyle BOWSTYLE_TYPE NOT NULL,
    draw_weight FLOAT NOT NULL CHECK (draw_weight > 0),
    club_id UUID,
    google_picture_url TEXT,
    google_subject TEXT NOT NULL UNIQUE,
    last_login_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


-- NOTE: bowstyle, draw_weight and current_club_id can change over time. An archer can change 
-- increase/decrease their draw_weight, change to another club, or change the bowstyle completely.
-- The archer table is going to hold the latest information.
