-- function calculate_points()
CREATE OR REPLACE FUNCTION calculate_points(
    shot_x REAL,
    shot_y REAL,
    target_x REAL,
    target_y REAL,
    radius REAL [],
    points INT []
) RETURNS INT AS $$
DECLARE
    distance REAL;
    ring_index INT;
    shot_points INT;
BEGIN
    -- Calculate the distance from the shooting to the center of the target
    distance := SQRT(POWER(shot_x - target_x, 2) + POWER(shot_y - target_y, 2));
    
    -- Find the appropriate ring
    FOR ring_index IN 1 .. array_length(radius, 1) LOOP
        IF distance <= radius[ring_index] THEN
            shot_points := points[ring_index];
            RETURN shot_points;
        END IF;
    END LOOP;
    
    -- If the shooting is outside all rings, return 0 points
    RETURN 0;
END;
$$ LANGUAGE plpgsql;

--View for shooting with calculated points
CREATE OR REPLACE VIEW shots_view AS
SELECT
    s.distance,
    (s.arrow_landing_time - s.arrow_disengage_time) AS flight_time,
    s.distance / (s.arrow_landing_time - s.arrow_disengage_time) AS speed,
    calculate_points(
        s.x_coordinate,
        s.y_coordinate,
        t.x_coordinate,
        t.y_coordinate,
        t.radius,
        t.points
    ) AS shot_points
FROM shooting AS s
INNER JOIN
    target AS t ON s.target_id = t.id;
