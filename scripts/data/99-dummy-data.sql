\c "arch-stats"

INSERT INTO archer (id, first_name, last_name, email, type_of_archer, bow_weight, draw_length)
VALUES 
    ('3f1e2d4a-1c2b-4d5e-8f6a-7b8c9d0e1f2a', 'John', 'Doe', 'john.doe@example.com', 'traditional', 40.5, 28.0),
    ('4e2d3c1b-5a6b-7c8d-9e0f-1a2b3c4d5e6f', 'Jane', 'Smith', 'jane.smith@example.com', 'compound', 50.0, 27.5),
    ('539897d7-f521-42aa-a2b7-ed5263f2564f', 'Alice', 'Johnson', 'alice.johnson@example.com', 'olympic', 45.0, 29.0),
    ('7fab51db-0051-4df8-a8be-8bd819ded689', 'Bob', 'Brown', 'bob.brown@example.com', 'barebow', 35.0, 26.0),
    ('df38c156-e7e8-4036-86d9-9ead85e98b65', 'Charlie', 'Davis', 'charlie.davis@example.com', 'traditional', 42.0, 28.5);

INSERT INTO arrow (id, archer_id, weight, diameter, spine, length, name, label_position)
VALUES 
    -- Arrows for John Doe
    ('0dc6c7b5-f584-4538-82f3-c0081a6596a0', '3f1e2d4a-1c2b-4d5e-8f6a-7b8c9d0e1f2a', 30.5, 0.25, 400, 28.0, 'Arrow 1', 5.0),
    ('96458074-a2a9-44e0-87f3-4e3188da5a36', '3f1e2d4a-1c2b-4d5e-8f6a-7b8c9d0e1f2a', 30.5, 0.25, 400, 28.0, 'Arrow 2', 5.0),
    ('b0deea9c-ca98-4ee4-9a64-cd4631ce9f5a', '3f1e2d4a-1c2b-4d5e-8f6a-7b8c9d0e1f2a', 30.5, 0.25, 400, 28.0, 'Arrow 3', 5.0),
    ('93ea2707-eff7-4167-b2b3-ecb260919c5b', '3f1e2d4a-1c2b-4d5e-8f6a-7b8c9d0e1f2a', 30.5, 0.25, 400, 28.0, 'Arrow 4', 5.0),
    ('982757d6-d704-408c-9faf-67e600441d6c', '3f1e2d4a-1c2b-4d5e-8f6a-7b8c9d0e1f2a', 30.5, 0.25, 400, 28.0, 'Arrow 5', 5.0),

    -- Arrows for Jane Smith
    ('17e212aa-321b-4dd8-b3dc-55c513f07f9e', '4e2d3c1b-5a6b-7c8d-9e0f-1a2b3c4d5e6f', 32.0, 0.30, 350, 27.5, 'Arrow 1', 5.5),
    ('527b6cf8-de95-48d1-9137-656c173a5373', '4e2d3c1b-5a6b-7c8d-9e0f-1a2b3c4d5e6f', 32.0, 0.30, 350, 27.5, 'Arrow 2', 5.5),
    ('49c16351-4f01-4589-aa48-f9eecc1fb0fc', '4e2d3c1b-5a6b-7c8d-9e0f-1a2b3c4d5e6f', 32.0, 0.30, 350, 27.5, 'Arrow 3', 5.5),
    ('8ce33679-e0b0-4a96-ab5a-e4fbb9319b2d', '4e2d3c1b-5a6b-7c8d-9e0f-1a2b3c4d5e6f', 32.0, 0.30, 350, 27.5, 'Arrow 4', 5.5),
    ('1d4c8108-cc40-4731-b85c-503d56ade888', '4e2d3c1b-5a6b-7c8d-9e0f-1a2b3c4d5e6f', 32.0, 0.30, 350, 27.5, 'Arrow 5', 5.5),

    -- Arrows for Alice Johnson
    ('205822a5-dac4-48a2-abc3-f661f6414356', '539897d7-f521-42aa-a2b7-ed5263f2564f', 28.0, 0.20, 450, 29.0, 'Arrow 1', 6.0),
    ('53d34fb0-0562-42ce-abb6-2f6d24b85c6d', '539897d7-f521-42aa-a2b7-ed5263f2564f', 28.0, 0.20, 450, 29.0, 'Arrow 2', 6.0),
    ('3cb76719-ac79-4b31-8484-2007373d2e53', '539897d7-f521-42aa-a2b7-ed5263f2564f', 28.0, 0.20, 450, 29.0, 'Arrow 3', 6.0),
    ('d7c48744-cb5d-4913-923a-cc77d33bc6fe', '539897d7-f521-42aa-a2b7-ed5263f2564f', 28.0, 0.20, 450, 29.0, 'Arrow 4', 6.0),
    ('7ed563f3-6f23-478e-831b-ba52fbfff3c4', '539897d7-f521-42aa-a2b7-ed5263f2564f', 28.0, 0.20, 450, 29.0, 'Arrow 5', 6.0),

    -- Arrows for Bob Brown
    ('8f6f6b96-740a-44a0-b299-543de8dd7de4', '7fab51db-0051-4df8-a8be-8bd819ded689', 29.5, 0.22, 420, 26.0, 'Arrow 1', 4.5),
    ('5b34bca0-aaa3-43a7-94b3-de9ad345af32', '7fab51db-0051-4df8-a8be-8bd819ded689', 29.5, 0.22, 420, 26.0, 'Arrow 2', 4.5),
    ('450d96bf-072d-4d02-a741-a65a2e33c36d', '7fab51db-0051-4df8-a8be-8bd819ded689', 29.5, 0.22, 420, 26.0, 'Arrow 3', 4.5),
    ('f59d89a1-eb90-44ba-ad06-e9804bd5405d', '7fab51db-0051-4df8-a8be-8bd819ded689', 29.5, 0.22, 420, 26.0, 'Arrow 4', 4.5),
    ('4fac5600-5f6f-4de4-9353-bf078d2a56cd', '7fab51db-0051-4df8-a8be-8bd819ded689', 29.5, 0.22, 420, 26.0, 'Arrow 5', 4.5),

    -- Arrows for Charlie Davis
    ('30986755-5729-490d-8986-55dc39db4417', 'df38c156-e7e8-4036-86d9-9ead85e98b65', 31.0, 0.28, 370, 28.5, 'Arrow 1', 5.2),
    ('a066a6c4-ec79-4cde-86d7-60a8dc69c620', 'df38c156-e7e8-4036-86d9-9ead85e98b65', 31.0, 0.28, 370, 28.5, 'Arrow 2', 5.2),
    ('61e1dcfb-fc73-4af1-ab0a-fbe39ffb35c9', 'df38c156-e7e8-4036-86d9-9ead85e98b65', 31.0, 0.28, 370, 28.5, 'Arrow 3', 5.2),
    ('be6586cf-9af9-49e0-835b-0be2e67d2f8c', 'df38c156-e7e8-4036-86d9-9ead85e98b65', 31.0, 0.28, 370, 28.5, 'Arrow 4', 5.2),
    ('3a90515f-3401-426c-8b29-10c96b8d2dbb', 'df38c156-e7e8-4036-86d9-9ead85e98b65', 31.0, 0.28, 370, 28.5, 'Arrow 5', 5.2);
