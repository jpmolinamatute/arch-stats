\c "arch-stats"
CREATE TABLE IF NOT EXISTS shooting (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_track_id UUID NOT NULL,
    arrow_id UUID NOT NULL,
    arrow_engage_time TIMESTAMP WITH TIME ZONE NOT NULL,
    draw_length REAL NOT NULL,
    arrow_disengage_time TIMESTAMP WITH TIME ZONE NOT NULL,
    arrow_landing_time TIMESTAMP WITH TIME ZONE,
    distance REAL NOT NULL,
    x_coordinate REAL NOT NULL,
    y_coordinate REAL NOT NULL,
    FOREIGN KEY (arrow_id) REFERENCES arrow (id) ON DELETE CASCADE,
    FOREIGN KEY (target_track_id) REFERENCES target_track (id) ON DELETE CASCADE
);

COMMENT ON COLUMN shooting.target_track_id IS 'it will be provided by the Raspberry Pi';
COMMENT ON COLUMN shooting.arrow_id IS 'it will be read by the bow sensor and the target sensor';
COMMENT ON COLUMN shooting.arrow_engage_time IS 'it will be read by the bow sensor';
COMMENT ON COLUMN shooting.draw_length IS 'it will be read by the bow sensor';
COMMENT ON COLUMN shooting.arrow_disengage_time IS 'it will be read by the bow sensor';
COMMENT ON COLUMN shooting.arrow_landing_time IS 'it will be read by the target sensor';
COMMENT ON COLUMN shooting.x_coordinate IS 'it will be read by the target sensor';
COMMENT ON COLUMN shooting.y_coordinate IS 'it will be read by the target sensor';

-- Create a function that will be called by the trigger 
CREATE OR REPLACE FUNCTION notify_shooting_change(archer_id UUID)
RETURNS TRIGGER AS $$
BEGIN
    RAISE NOTICE 'The function notify_shooting_change was called';
    IF EXISTS (SELECT 1 FROM arrow WHERE id = NEW.arrow_id AND archer_id = archer_id) THEN
        PERFORM pg_notify('shooting_change', row_to_json(NEW)::text);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
