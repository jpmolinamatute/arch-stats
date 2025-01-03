\c "arch-stats"

CREATE TABLE IF NOT EXISTS registration (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL,
    archer_id UUID NOT NULL,
    lane_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT current_timestamp,
    UNIQUE (archer_id, tournament_id),
    UNIQUE (archer_id, lane_id, tournament_id),
    FOREIGN KEY (tournament_id) REFERENCES tournament (id) ON DELETE CASCADE,
    FOREIGN KEY (archer_id) REFERENCES archer (id) ON DELETE CASCADE,
    FOREIGN KEY (lane_id) REFERENCES lane (id) ON DELETE CASCADE
);
COMMENT ON COLUMN registration.archer_id IS 'it will be provided by the WebUI';
COMMENT ON COLUMN registration.tournament_id IS 'it will be provided by the WebUI';
COMMENT ON COLUMN registration.lane_id IS 'it will be provided by the WebUI';
