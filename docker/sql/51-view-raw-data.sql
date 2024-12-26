CREATE OR REPLACE VIEW raw_data AS
SELECT
    s.arrow_engage_time,
    s.arrow_disengage_time,
    s.arrow_landing_time,
    s.x_coordinate AS shot_x_coordinate,
    s.y_coordinate AS shot_y_coordinate,
    s.pull_length,
    s.distance,
    s.target_id,
    t.x_coordinate AS target_x_coordinate,
    t.y_coordinate AS target_y_coordinate,
    t.radius,
    t.points,
    t.height AS target_height,
    t.name AS target_name
FROM shooting AS s
INNER JOIN target AS t ON s.target_id = t.id;
