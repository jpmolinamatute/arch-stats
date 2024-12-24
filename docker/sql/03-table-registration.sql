-- Registration table
--
-- Links archers to tournaments (junction table).
-- Ensures that each archer can register for multiple tournaments.

CREATE TABLE IF NOT EXISTS registration (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    archer_id UUID REFERENCES archer (id) ON DELETE CASCADE,
    tournament_id UUID REFERENCES tournament (id) ON DELETE CASCADE,
    lane_id UUID REFERENCES lane (id) ON DELETE CASCADE,
    UNIQUE (archer_id, tournament_id),
    UNIQUE (archer_id, lane_id, tournament_id)
);
-- Relationships:
-- - M:1 with archer (each registration belongs to one archer)
-- - M:1 with tournament (each registration belongs to one tournament)

COMMENT ON TABLE registration IS 'An archer can be registered in only one tournament at the time. Within a tournament, an archer can be registered in only one lane.';
