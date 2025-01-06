\c "arch-stats"
CREATE TABLE IF NOT EXISTS shooting (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_track_id UUID NOT NULL,
    arrow_id UUID NOT NULL,
    arrow_engage_time TIMESTAMP WITH TIME ZONE NOT NULL,
    draw_length REAL NOT NULL,
    arrow_disengage_time TIMESTAMP WITH TIME ZONE NOT NULL,
    arrow_landing_time TIMESTAMP WITH TIME ZONE,
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
CREATE OR REPLACE FUNCTION notify_shooting_change() RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('shooting_change', json_build_object(
        'operation', TG_OP,
        'data', row_to_json(NEW)
    )::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create the trigger that calls the function on INSERT or UPDATE
CREATE TRIGGER shooting_change_trigger
AFTER INSERT OR UPDATE ON shooting
FOR EACH ROW EXECUTE FUNCTION notify_shooting_change();
