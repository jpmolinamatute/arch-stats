-- Seed playground data for arch-stats
-- - One open session
-- - One target for that session (120 x 120 at 18m)
-- - Three arrows
-- - 110 shots (100 landed on target and 10 missed)
-- - Three WA faces (40 cm) positioned within target bounds

DO $$
DECLARE
    v_session_id UUID;
    v_target_id  UUID;
    v_ar1        UUID;
    v_ar2        UUID;
    v_ar3        UUID;
    v_base_ts    TIMESTAMPTZ := now();
    i            INT;
BEGIN
    INSERT INTO sessions (is_opened, start_time, location, is_indoor)
    VALUES (TRUE, now(), 'Playground Range', FALSE)
    RETURNING id INTO v_session_id;

    INSERT INTO targets (max_x, max_y, distance, session_id, target_sensor_id)
    VALUES (120.0, 120.0, 18, v_session_id, '2d7050ca-125c-4584-97ee-79cc9301c891')
    RETURNING id INTO v_target_id;

    REFRESH MATERIALIZED VIEW current_target;

    INSERT INTO arrows (id, length, human_identifier, is_programmed, label_position, weight, diameter, spine)
    VALUES (uuid_generate_v4(), 29.0, 'A1', FALSE, 1.0, 350.0, 5.7, 400)
    RETURNING id INTO v_ar1;

    INSERT INTO arrows (id, length, human_identifier, is_programmed, label_position, weight, diameter, spine)
    VALUES (uuid_generate_v4(), 29.5, 'A2', TRUE, 1.5, 355.0, 5.8, 410)
    RETURNING id INTO v_ar2;

    INSERT INTO arrows (id, length, human_identifier, is_programmed, label_position, weight, diameter, spine)
    VALUES (uuid_generate_v4(), 30.0, 'A3', FALSE, 2.0, 360.0, 5.9, 420)
    RETURNING id INTO v_ar3;

    FOR i IN 1..100 LOOP
        INSERT INTO shots (
            arrow_id,
            target_sensor_id,
            arrow_engage_time,
            arrow_disengage_time,
            arrow_landing_time,
            x,
            y
        ) VALUES (
            CASE (i % 3)
                WHEN 1 THEN v_ar1
                WHEN 2 THEN v_ar2
                ELSE v_ar3
            END,
            '2d7050ca-125c-4584-97ee-79cc9301c891',
            v_base_ts + make_interval(secs => 5 * i),
            v_base_ts + make_interval(secs => 5 * i + 2),
            v_base_ts + make_interval(secs => 5 * i + 4),
            -- keep within target 0..120 using modular drift
            (10 + ((i * 7) % 100))::real,
            (15 + ((i * 11) % 100))::real
        );
    END LOOP;

    FOR i IN 1..10 LOOP
        INSERT INTO shots (
            arrow_id,
            target_sensor_id,
            arrow_engage_time,
            arrow_disengage_time,
            arrow_landing_time,
            x,
            y
        ) VALUES (
            CASE (i % 3)
                WHEN 1 THEN v_ar1
                WHEN 2 THEN v_ar2
                ELSE v_ar3
            END,
            '2d7050ca-125c-4584-97ee-79cc9301c891',
            v_base_ts + make_interval(secs => 5 * (97 + i)),
            v_base_ts + make_interval(secs => 5 * (97 + i) + 2),
            NULL,
            NULL,
            NULL
        );
    END LOOP;


    INSERT INTO faces (x, y, human_identifier, is_reduced, wa_faces_id, target_id) VALUES
        (30.0, 30.0, 'F1', FALSE, '417d902f-3a4b-4262-a530-6d36f3dd47e7', v_target_id),
        (90.0, 30.0, 'F2', FALSE, '417d902f-3a4b-4262-a530-6d36f3dd47e7', v_target_id),
        (30.0, 90.0, 'F3', FALSE, '417d902f-3a4b-4262-a530-6d36f3dd47e7', v_target_id);
END $$;
