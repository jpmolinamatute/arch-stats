# Arch Stats: Analyzing Archery Performance Through Data

## Background

Practicing archery can feel inconsistent: some days, I see improvement; other days, it feels like I’ve hit a plateau—or even regressed. Tracking performance is challenging due to the numerous variables influencing each shot.

## The Problem

I need a method to objectively measure and track my archery performance over time.

## The Solution

To address this, the solution involves:

1. Identifying measurable variables that impact performance.
2. Consistently collecting data on these variables.
3. Recording the data over time.
4. Presenting the information in an intuitive, actionable format.

## The App

The application, Arch Stats, helps archers improve by providing data-driven insights. By tracking shot accuracy, equipment details, and other metrics, it enables archers to monitor progress and make informed decisions.

## Actors

* Archer: A WebUI user who registers and competes in tournaments, and reviews their own performance data.
* Tournament Organizer: A WebUI user who creates and manages tournaments and register new archers.

## Epic Overview

To achieve this, the app includes four epics for data collection, processing, and presentation:

### Hardware Integration

This epic enable Archer and Tournament Organizer to utilize hardware components for data collection this includes:

* Raspberry Pi: Serves as the central processing unit.
* Sensors: Embedded in the bow and target, connected to the Raspberry Pi.
* Bluetooth Low Energy (BLE) Beacons: Used for short-range communication.
* Bluetooth Tags: Identify equipment during data collection.
* The Frame: Holds the sensors and Raspberry Pi in place.

The goal of this Epic is to organize all user stories related to sensors, Raspberry Pi and the frame

#### Out of Scope

any software or coding

### Data collection & storage

This epic enables Archer to collect & store their metrics through sensors during shooting. This includes:

* Rust Applications:
  * arrow_reader: Registers and programs arrows.
  * bow_reader: Collects data from bow sensors during setup.
  * target_reader: Records details of each shot.
* PostgreSQL Database: Stores the collected data.

#### Out of Scope

Any data processing or visualization

### Archer & tournament management

This epic enables Tournament Organizer to register new archers in the system and add them to tournaments. This includes registering the archer's personal and equipment details. This will be done through:

* Python Web Server: Serves the interface (compiled with CPython).
* WebUI: Built with Angular for a user-friendly interaction.

### Data visualization & analysis

This epic enables Archers to analyze their metrics & performance with visual aids, this includes:

* Python Web Server: Processes data for visualization. (compiled with CPython).
* WebUI: Displays user-friendly analytics, built with Angular.

## Workflow

So far there are 3 workflows identified:

1. [Archer Registration.](./docs/archer_registration.png)
2. [Tournament Creation.](./docs/tournament_creation.png)
3. [Tournament.](./docs/tournament_flow.png)

### Shooting and Data Collection

During a session, the system captures real-time data:

* Bow Sensors: Arrow engage/disengage times, pull length, and distance.
* Target Sensors: Landing time and coordinates (x, y).

## Data Sources

### Sensors

Sensors in the bow and target capture:

* Bow Metrics: Pull length, engage/disengage times, and shot distance.
* Target Metrics: Landing coordinates, radius, and height.

### User-Provided Data

Archers input personal and equipment details (via WebUI) during registration and target setup.

## Data Source Integration in the Workflow

The following table illustrates the data sources at each step of the workflow:

| Step                    | Data Source        | Details                                        |
| ----------------------- | -------------------| -----------------------------------------------|
| User Registration       | Archer             | Personal and equipment data entered via WebUI. |
| Tournament Registration | Archer             | Tournament selection and lane assignment provided via WebUI. |
| Target Setup            | Archer and Sensors | Archers provide names; sensors supply x/y coordinates, radius, and height. |
| Shooting                | Sensors            | Metrics collected in real-time via target_reader and stored in the database. |

## Raspberry Pi Setup

The Raspberry Pi 5 (256GB NVMe SSD) setup includes:

* PostgreSQL for data storage.
* Python and Rust services for data processing.
* Proper allocation of memory and CPU resources to balance services.

## Work Effort

The work is going to be tracked in [JIRA](https://jpmolinamatute.atlassian.net). However, we are going to have an introduction on each module in this repository as follow:

* [The Server](./server/README.md)
* [The WebUI](./webui/README.md)
* [The Shot Reader](./target_reader/README.md)
* [The Bow Reader](./bow_reader/README.md)
* [The Arrow Reader](./arrow_reader/README.md)
* [The Database](./docker/README.md)
* [The Raspberry Pi Setup](./os/README.md)
* [Tools](./scripts/README.md)

### The Frame (structure to hold the sensors and the Raspberry Pi)

TBD

## Future Features

<!--
@TODO: Under "Future Features," add more emphasis on the scalability strategy and potential roadmap for adding new technologies or integrations.
-->

In the initial stages, the system will support only one lane and one archer due to hardware and volunteer constraints. Future enhancements may include:

Expanding the number of lanes and archers by integrating additional Raspberry Pis and sensors.
Developing a networked database to handle data from multiple Raspberry Pis, with authentication for secure data writing.
