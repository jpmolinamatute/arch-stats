-- Insert into tournament table
INSERT INTO tournament (
    id, start_time, end_time, number_of_lanes, is_opened
) VALUES (
    '550e8400-e29b-41d4-a716-446655440000', '2023-01-01 09:00:00', NULL, 10, TRUE
);

-- Insert into archer table
INSERT INTO archer (
    id, first_name, last_name, password, email, genre, type_of_archer, bow_poundage, created_at
) VALUES (
    '550e8400-e29b-41d4-a716-446655440001',
    'John',
    'Doe',
    'password123',
    'john.doe@example.com',
    'male',
    'olympic',
    40.5,
    current_timestamp
);

-- Insert into lane table
INSERT INTO lane (
    id, tournament_id, lane_number, number_of_archers,max_x_coordinate, max_y_coordinate
) VALUES (
    '550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440000', 1, 4, 100.0, 200.0
);

-- Insert into target table
INSERT INTO target (
    id,
    x_coordinate,
    y_coordinate,
    radius,
    points,
    height,
    human_readable_name,
    lane_id,
    archer_id
) VALUES (
    '550e8400-e29b-41d4-a716-446655440003',
    1.23,
    4.56,
    ARRAY[10.0, 20.0, 30.0],
    ARRAY[10, 9, 8],
    1.5,
    'Target 1',
    '550e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440001'
), (
    '550e8400-e29b-41d4-a716-446655440005',
    2.34,
    5.67,
    ARRAY[15.0, 25.0, 35.0],
    ARRAY[10, 9, 8],
    1.6,
    'Target 2',
    '550e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440001'
), (
    '550e8400-e29b-41d4-a716-446655440007',
    3.45,
    6.78,
    ARRAY[20.0, 30.0, 40.0],
    ARRAY[10, 9, 8],
    1.7,
    'Target 3',
    '550e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440001'
);



-- Insert into arrow table
INSERT INTO arrow (
    id, weight, diameter, spine, length, human_readable_name, archer_id, label_position
) VALUES (
    '4c19ead8-9b49-4876-978c-f22b5ec5edbf',
    30.0,
    5.0,
    400.0,
    28.0,
    'Arrow 1',
    '550e8400-e29b-41d4-a716-446655440001',
    10.0
), (
    '854eef0f-badb-414d-a580-2b4aa5df79eb',
    32.0,
    5.2,
    420.0,
    28.5,
    'Arrow 2',
    '550e8400-e29b-41d4-a716-446655440001',
    10.5
), (
    '7122a4ad-eece-46eb-9424-f680b7b12d4e',
    34.0,
    5.4,
    440.0,
    29.0,
    'Arrow 3',
    '550e8400-e29b-41d4-a716-446655440001',
    11.0
), (
    '1cd1d238-cf78-40ed-a32b-5f8208473183',
    36.0,
    5.6,
    460.0,
    29.5,
    'Arrow 4',
    '550e8400-e29b-41d4-a716-446655440001',
    11.5
), (
    '55af4551-bc88-4602-8a43-4837c16cdb81',
    38.0,
    5.8,
    480.0,
    30.0,
    'Arrow 5',
    '550e8400-e29b-41d4-a716-446655440001',
    12.0
), (
    'e1388bf0-df48-4a2a-867e-e882348c6c3d',
    40.0,
    6.0,
    500.0,
    30.5,
    'Arrow 6',
    '550e8400-e29b-41d4-a716-446655440001',
    12.5
), (
    'ed8f8586-1137-4192-bcc7-c0607d7c2df4',
    42.0,
    6.2,
    520.0,
    31.0,
    'Arrow 7',
    '550e8400-e29b-41d4-a716-446655440001',
    13.0
), (
    '1f91e8e0-77dc-4621-8108-f000b0e690a5',
    44.0,
    6.4,
    540.0,
    31.5,
    'Arrow 8',
    '550e8400-e29b-41d4-a716-446655440001',
    13.5
), (
    'edd2e424-9d62-4457-ad75-08b3de4f4b72',
    46.0,
    6.6,
    560.0,
    32.0,
    'Arrow 9',
    '550e8400-e29b-41d4-a716-446655440001',
    14.0
), (
    'd5e554dc-face-47b3-adca-d6354fbe6b1e',
    48.0,
    6.8,
    580.0,
    32.5,
    'Arrow 10',
    '550e8400-e29b-41d4-a716-446655440001',
    14.5
), (
    '6c34d6b9-c778-4e78-9ad8-9ade92890659',
    50.0,
    7.0,
    600.0,
    33.0,
    'Arrow 11',
    '550e8400-e29b-41d4-a716-446655440001',
    15.0
), (
    '6a1a0923-1bc2-40e0-87cb-5d342ea83643',
    52.0,
    7.2,
    620.0,
    33.5,
    'Arrow 12',
    '550e8400-e29b-41d4-a716-446655440001',
    15.5
);

-- Insert into registration table
INSERT INTO registration (
    id, archer_id, lane_id, tournament_id
) VALUES (
    uuid_generate_v4(),
    '550e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440000'
);
