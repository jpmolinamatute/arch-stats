# Arch Stats Database

We are using PostgreSQL as database for the Archery Stats project. The database will track shot by one or more archers within a tournament. Th `db.sql` script sets up a database to track archery tournaments, archers, arrows, lanes, targets, and shots. The database includes tables, functions, triggers, and views to manage and process the data efficiently.

## Database Structure

### Extensions

- `uuid-ossp`: Provides functions to generate universally unique identifiers (UUIDs).

### Functions

- `calculate_points(shot_x, shot_y, target_x, target_y, radius, points)`: Calculates the points for a shot based on its distance from the target center.
- `check_lane_number()`: Ensures that the lane number is within the valid range for the tournament.
- `check_shot_coordinates()`: Ensures that the shot coordinates are within the valid range for the lane.

### Types

- `archer_type`: Enum type for archer categories (`compound`, `traditional`, `barebow`, `olympic`).
- `archer_genre`: Enum type for archer gender (`male`, `female`, `no-answered`).

### Tables

- `tournament`: Stores tournament details including start time, end time, number of lanes, and status.
- `archer`: Stores archer details including name, email, genre, type, and bow poundage.
- `arrow`: Stores arrow details including weight, diameter, spine, length, and associated archer.
- `lane`: Stores lane details including tournament association, lane number, and maximum coordinates.
- `target`: Stores target details including coordinates, radius, points, height, and associated lane and archer.
- `registration`: Stores registration details linking archers to tournaments and lanes.
- `shot`: Stores shot details including timing, coordinates, pull length, distance, and associated lane and arrow.

### Triggers

- `check_lane_number_trigger`: Ensures valid lane numbers before inserting or updating rows in the `lane` table.
- `check_shot_coordinates_trigger`: Ensures valid shot coordinates before inserting or updating rows in the `shot` table.

### Views

- `shots_view`: (Commented out) Calculates and displays shot points based on shot and target data.
- `raw_data`: (Commented out) Displays raw shot and target data for analysis.

### Comments

Detailed comments are provided for tables, columns, and constraints to explain their purpose and usage.

## Usage

1. **Create Database**: Uncomment and execute the `CREATE DATABASE archery;` and `COMMENT ON DATABASE archery IS 'the database will track shot by one or more archers within a tournament.';` lines to create and comment on the database.
2. **Connect to Database**: Uncomment and execute the `\c archery;` line to connect to the `archery` database.
3. **Create Extension**: Ensure the `uuid-ossp` extension is available by executing `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`.
4. **Create Functions**: Execute the function definitions to create the necessary functions for calculating points and validating data.
5. **Create Types**: Execute the type definitions to create the necessary enum types for archer categories and gender.
6. **Create Tables**: Execute the table definitions to create the necessary tables for storing tournament, archer, arrow, lane, target, registration, and shot data.
7. **Create Triggers**: Execute the trigger definitions to create the necessary triggers for validating lane numbers and shot coordinates.
8. **Create Views**: Uncomment and execute the view definitions if needed for displaying calculated shot points and raw data.
9. **Add Comments**: Execute the comment definitions to add detailed comments to the tables and columns for better understanding and documentation.

## Data Flow

1. A tournament is created if none exists.
2. Each archer registers in the database if not already registered, along with their arrows.
3. Each archer selects a tournament and a lane.
4. The tournament starts after all archers have selected a lane.
5. Archers shoot at the target, and shots are recorded in the database.
6. The tournament ends when all archers have finished shooting.

## Notes

- Ensure that the database and required extensions are properly set up before executing the script.
- Follow the data flow to understand the sequence of operations and how data is managed in the database.
- Use the provided functions, triggers, and views to maintain data integrity and facilitate data processing.

This README provides a technical overview of the database structure and usage to help new developers understand and apply the SQL script effectively.
