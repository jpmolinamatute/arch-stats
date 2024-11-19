# Arch Stats

This project is designed to help archers enhance their performance through data-driven insights. By tracking shot accuracy, equipment details, and more, Arch Stats enables archers to monitor progress and make informed decisions. The goal is to capture reliable, actionable data and present it to the archer for their analysis.

To achieve this, the system is divided into four modules, each responsible for specific aspects of data collection, processing, and presentation:

## System Modules

### Hardware Module

This module involves the physical components needed for data collection:

* Raspberry Pi as the central processing unit.
* Sensors embedded in the archer's bow and the target, connected to the Raspberry Pi. (Specific sensor types will be determined later.)
* Bluetooth Low Energy (BLE) Beacons.
* Bluetooth tags for equipment identification.

### Sensor Data Collection Module

This module includes three Rust applications and a PostgreSQL database running on the Raspberry Pi:

* arrow_reader: Handles arrow registration and programming.
* target_reader: Collects and processes data from target sensors during target setup.
* shot_reader: Records shot details from bow and target sensors during archers' shots.

### Personal and Registration Module

This module manages user registration and equipment setup via:

* A Python web server running on the Raspberry Pi. The Python code will compiled using CPython
* A WebUI built with Angular, served by the Python server.

### Data Visualization and Analysis Module

Provides multiple charts & tools to query archers data for their analysis so that they draw a conclusion about their performance:

* A Python web server.
* A WebUI for presenting analysis, also built with Angular and integrated into the Python server.

## Workflow

1. Archer Registration: Archers register in the system through the WebUI using Google OAuth authentication. During registration, they provide the following details:
  * Full name.
  * Email.
  * Gender (male, female, no_specify).
  * Archery style (e.g., compound, traditional, barebow, Olympic).
  * Bow poundage.
  * NFC programming for each arrow (via the arrow_reader app) and collect the following arrow information:
    * length
    * weight
    * diameter
    * spine
2. Tournament Creation: Tournaments can be created with the status "open" if no other tournaments are currently active. The number of lanes is set with a maximum capacity (e.g., five lanes, two archers per lane).
3. Tournament Registration: Archers in the system can register for tournaments and select an available lane. During this process target data is collected via WebUI and the target_reader app:
  * x_coordinate: Collected from target sensors.
  * y_coordinate: Collected from target sensors.
  * radius: Collected from target sensors.
  * max_x_coordinate: Collected from target sensors.
  * max_y_coordinate: Collected from target sensors.
  * height: Collected from target sensors.
  * points: Entered manually by the archer via WebUI.
  * human-readable name: Entered manually by the archer via WebUI.
  * Sensor ID: this will be provided by the sensor on the raspberry pi.
4. Shooting and Data Collection: During a shooting session, data is captured and stored in real-time:
* Bow Sensors:
  * Arrow engage time.
  * Arrow disengage time.
  * Pull length.
  * Distance.
* Target Sensors:
  * Arrow landing time.
  * x_coordinate, y_coordinate.
5. Tournament Completion: The tournament ends with the status "closed," halting further data collection.

## Data Sources

The system relies on two primary sources of data:

### Sensors

Sensors embedded in the archer's bow and the target provide real-time performance metrics. This data is collected and processed by three Rust applications:
arrow_reader: Programs NFC tags on arrows and handles related data.
target_reader: Collects data from target sensors, including x/y coordinates, radius, and height.
shot_reader: Gathers shooting metrics such as arrow engage/disengage times, landing times, and distances.

### Archers (User-Provided Data)

Archers use the WebUI to enter information during specific activities:
User Registration: Archers input personal and equipment details, such as name, email, archery style, bow poundage, and arrow specifications.
Tournament Registration: Archers select a tournament and lane, providing target setup details when necessary.
Target Setup: Archers configure targets by specifying human-readable names and additional information to supplement sensor data.

All measurements are going to be stored and manipulated using the metric system. However, the measurements may be displayed in either in metric or imperial depending on user's preference.

## Data Source Integration in the Workflow

The following table illustrates the data sources at each step of the workflow:

| Step                    | Data Source        | Details                                        |
| ----------------------- | -------------------| -----------------------------------------------|
| User Registration       | Archer             | Personal and equipment data entered via WebUI. |
| Tournament Registration | Archer             | Tournament selection and lane assignment provided via WebUI. |
| Target Setup            | Archer and Sensors | Archers provide names; sensors supply x/y coordinates, radius, and height. |
| Shooting                | Sensors            | Metrics collected in real-time via shot_reader and stored in the database. |

## Database Tables

The system's PostgreSQL database includes the following tables:

| Table        | Description                            |
| ------------ | -------------------------------------- |
| tournaments  | Stores tournament-related data.        |
| archers      | Contains archer registration details.  |
| lanes        | Tracks lane assignments and data.      |
| arrows       | Holds arrow-specific details.          |
| targets      | Stores target specifications and data. |
| registration | Tracks archer participation in events. |
| shots        | Records shot-related metrics.          |

## File structure

The project's file organization is outlined below. This structure is a work in progress and will evolve during development:

```sh
tree  --gitignore -FalL3 -I .git
```

```plain

arch_stats
./
├── arrow_reader/
│   ├── Cargo.lock
│   ├── Cargo.toml
│   └── src/
│       └── main.rs
├── docker/
│   ├── db.sql
│   ├── docker-compose.yaml
│   └── setup.cfg
├── .gitignore
├── LICENSE
├── README.md
├── server/
│   ├── app/
│   │   ├── src/
│   │   └── tests/
│   ├── poetry.lock
│   ├── pyproject.toml
│   ├── .python-version
│   ├── README.md
│   └── tasks.py
├── shot_reader/
│   ├── Cargo.lock
│   ├── Cargo.toml
│   └── src/
│       └── main.rs
├── target_reader/
│   ├── Cargo.lock
│   ├── Cargo.toml
│   └── src/
│       └── main.rs
├── thoughts.md
└── webui/
    ├── angular.json
    ├── .editorconfig
    ├── package.json
    ├── package-lock.json
    ├── public/
    │   └── favicon.ico
    ├── README.md
    ├── src/
    │   ├── app/
    │   ├── index.html
    │   ├── main.ts
    │   └── styles.css
    ├── tsconfig.app.json
    ├── tsconfig.json
    └── tsconfig.spec.json


```

## Development Focus

Current Priorities:

Since hardware availability is limited, development will initially focus on:

Rust Applications:
    Completing shot_reader, which will:
        Read bow and target sensor data.
        Validate and store data in the database.
    Data to be recorded includes:
        UUID, arrow engage/disengage times, landing time, x/y coordinates, pull length, and distance.

Database Design:
    Defining the schema for efficient data storage and retrieval.

## Future Features

In the initial stages, the system will support only one lane and one archer due to hardware and volunteer constraints. Future enhancements may include:

Expanding the number of lanes and archers by integrating additional Raspberry Pis and sensors.
Developing a networked database to handle data from multiple Raspberry Pis, with authentication for secure data writing.
