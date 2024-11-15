# DB Considerations

## Flow

1. Archers register in the system. This is NOT the same as registration to a tournament.
   1. Archer add their arrows. This mean programming the NFC tag and adding the arrow info into the DB.
2. A tournament is created.
   1. Lanes are created.
3. Archers register in the tournament.
4. Archers shoot.

## Source of data

1. Web user.
   1. Sensor to program an arrow??
2. Rust app.
   1. Sensor in Bow.
   2. Senor in Raspberry Pi.

## Tables

* **tournaments table**: will be populated by only user data.
* **archers table**:  will be populated by only user data.
* **lanes table**: will be populated by both user & sensor data.
* **arrows table**: will be populated by both user & sensor data.
* **targets table**: will be populated by both user & sensor data.
* **registration**: will be populated by only user data.
* **shots**: will be populated by both user & sensor data.

## questions

* How are we going to register arrows NFC labels?
