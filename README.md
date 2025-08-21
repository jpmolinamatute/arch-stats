# Arch Stats: track and understand your archery performance over time

- [Arch Stats: track and understand your archery performance over time](#arch-stats-track-and-understand-your-archery-performance-over-time)
  - [Key Features](#key-features)
  - [How It Works](#how-it-works)
    - [Arrow Registration](#arrow-registration)
    - [Sessions and Shots](#sessions-and-shots)
  - [Reviewing Your Performance](#reviewing-your-performance)
  - [Future Plans](#future-plans)
  - [For Software Developers](#for-software-developers)
    - [Backend](#backend)
    - [Frontend](#frontend)
    - [Development Philosophy](#development-philosophy)

Archery can feel inconsistent. Some days every shot hits the mark; other days, the archer may struggle to replicate past performance. With so many variables affecting each shot/form, equipment, fatigue, environment it's hard to know whether real progress is being made or what needs adjustment. Arch Stats is a tool designed to remove that guesswork. It collects data from each shooting session and presents clear insights so the archer can objectively measure and analyze performance and consistency over time. By recording every arrow and every shot, Arch Stats highlights areas for improvement and keeps the archer motivated with real, measurable progress.

## Key Features

- **Progress Tracking:** Every practice session and shot is logged, allowing the archer to track improvement over days, months, and years. Archers can see objective trends in accuracy and consistency, rather than relying on gut feeling. This helps determine whether they're plateauing or making steady gains.
- **Visual Charts & Graphs:** Arch Stats provides instant visual feedback on the archer's shooting. **Scatter plot charts** show where the archer's arrows land on the target, helping the archer analyze grouping and precision. **Line graphs** illustrate trends (like the archer's accuracy or shot timing) across multiple sessions. All charts update in real-time as the archer shoot, so the archer get immediate insights after each arrow.
- **Performance Analysis Tools:** The app includes built-in tools to help the archer analyze the archer's performance. The archer can filter and query the archer's shot data to spot patterns and trends. For example, compare performance between different practice sessions, see how the archer's accuracy varies at different distances, or even evaluate if a particular arrow consistently flies differently from the rest. These **pre-set analyses** make it easy to uncover trends without needing any technical skills.
- **Data Export:** Your data is the archer's. Arch Stats lets the archer **export all the archer's raw data in CSV format**, so the archer can open it in Excel or other programs. This is great if the archer want to do custom analysis or keep a personal log. The archer can back up the archer's entire shooting history or share it with a coach in a universally readable format.

> **Note**: Arch Stats is currently a work in progress, an early proof of concept or pre-alpha build. The current implementation is geared toward a single archer with one bow setup, under consistent conditions. But this README represents the broader vision: a more powerful, flexible system that will eventually support multiple archers, varied setups, and more complex scenarios. What the archer see here is just the beginning.

Clarifying Archery Terms in Arch Stats

Because Arch Stats tracks shots with multiple hardware components, it's important to use the right term at the right time. Four terms are closely related but different:

- Target: The physical structure designed to receive arrows (often called a butt). This is where the archer places both the target face and the target sensor.
- Target Face: The circular paper face with printed rings and scoring zones. This is what the archer visually aims at. An arrow can strike the target structure but miss the face entirely.
- Target Sensor: The electronic device mounted at the top of the target. It detects arrow impacts and reports their position. It is shielded to avoid damage and environmental noise.
- Target View: The field of detection of the target sensor. Think of it as the "scope" through which the sensor sees the target face. If an arrow lands outside this area, the system registers a miss.

Examples:

If an arrow hits the target but misses the target face, the sensor still records an impact and Arch Stats logs a shot. It is counted as a miss on the face, but a valid shot record exists.

If an arrow misses the entire target (no sensor impact detected), Arch Stats creates a shot record with only partial information (arrow was released, but no landing detected).

This distinction ensures the system can track performance accurately, even when arrows miss the scoring area.

## How It Works

Arch Stats combines a friendly web application with electronic sensors on the archer's equipment and on the target to seamlessly record data with minimal effort from the archer. The system runs on a small computer (such as a Raspberry Pi 5) connected to three types of sensors:

- **Arrow Reader:** Connected directly to the Raspberry Pi 5, this sensor assigns a unique ID to each arrow so the system knows exactly which one was shot.
- **Bow sensor:** Mounted on the archer's bow, it detects when an arrow is set and when it's released.
- **Target sensor:** Also connected to the Raspberry Pi 5 and mounted on the target structure, it detects when and where the arrow hits the target within the target view.

Using these components, Arch Stats tracks the entire lifecycle of a shot. There are two main parts to using Arch Stats: first **registering the archer's arrows**, and then **recording a shooting session**.

### Arrow Registration

Before the archer start logging practice sessions, the archer'll register all their arrows in the system. This process gives each arrow its own unique identifier and stores its characteristics (weight, length, spine, etc.). Arch Stats will guide the archer through entering these details via the web interface. Each arrow is then **programmed with a small tag or code** (using the Arrow Reader sensor) so that the system can recognize that arrow every time the archer shoot it. This way, the archer can later see if certain arrows perform differently from others.

The flowchart below shows how the arrow registration process works in Arch Stats:

```mermaid
    flowchart TB
    classDef user fill:#e3f2fd,stroke:#2196f3,color:#0d47a1,font-weight:bold;
    classDef app fill:#e8f5e9,stroke:#43a047,color:#1b5e20,font-weight:bold;
    classDef sensor fill:#fff3e0,stroke:#ff9800,color:#e65100,font-weight:bold;
    classDef subgraphStyle fill:#f9f9f9,stroke:#cccccc,color:#444;

    START(["Start Arrow Registration"])
    Q1{"registering a new arrow?"}
    R1["app: generates a<br />new UUID"]:::app
    F1["app: ask for<br />'human_identifier'"]:::app
    U1["user: provides arrow's<br />'human_identifier'"]:::user
    F2["app: ask for<br />'weight'<br />'diameter'<br />'spine'<br />'length'<br />'human_identifier'<br />'label_position'"]:::app
    R2["app: pulls arrow's data"]:::app
    U2["user: provides arrow<br />information"]:::user
    S1["app: validate arrow's<br />'human_identifier' and set<br />'is_programmed' to false<br />generate payload"]:::app
    S2["app: update arrow's<br />'is_programmed' to true<br />in arrows table"]:::app
    S3["app: update arrow's<br />'is_programmed' to true<br />in payload"]:::app
    S4["app: save payload into<br />arrows table"]:::app
    END1(["End."])
    subgraph PROGRAM["Program arrow with a unique identifier"]
        class PROGRAM subgraphStyle;
        SCAN["sensor: scan arrow"]:::sensor
        VERIFY{"app: Success?"}:::app
        SUCCESS["app: display<br />an OK on WebUI"]:::app
        FAIL["app:  display<br />a Retry? on WebUI"]:::app
        SCAN --> VERIFY
        VERIFY -->|Yes| SUCCESS
        VERIFY -->|No| FAIL
        FAIL -->|Yes| SCAN
    end

    START --> Q1
    Q1 -->|Yes.|R1
    Q1 -->|No.|F1

    R1 --> F2
    F1 --> U1
    U1 --> R2
    
    R2 --> PROGRAM
    F2 --> U2
    U2 --> S1
    PROGRAM --> S2
    S2 --> END1
    S1 --> PROGRAM
    PROGRAM --> S3
    S3 --> S4
    S4 --> END1
```

### Sessions and Shots

Once the archer's arrows are registered, they are ready to record a shooting session. A session represents a round of shooting. Using the Arch Stats web app, the archer opens a new session at the start of shooting. The app prompts the archer to enter session details such as the shooting location and distance to the target, indoors/outdoors, providing essential context for the recorded data. The archer may want to also calibrate the target(s) used in the session.

After that, the archer simply shoot as they normally would. Every time the archer shoot an arrow:

- The bow sensor detects the moment an arrow enter in contact with the bow and the moment the arrow leave the bow (recording events like "arrow engaged" and "arrow released").
- The target sensor detects the arrow hitting the target, recording the exact time and the location of impact (coordinates) within "target view".
  - If an arrow lands outside the target face but still strikes the target, Arch Stats records a shot with coordinates but scores it as a miss.
  - If the arrow misses the whole target, Arch Stats records a shot with partial information (arrow release only, no landing).
- The system automatically links this information with the specific arrow that the archer shot (thanks to the arrow's ID) and creates a new entry in the database for that shot.

If an arrow misses the target view, the system notes it as a miss (no impact recorded within a short time window). Throughout the session, the archer can glance at the web interface to see the archer's shots plotting in real-time on a virtual target and key stats updating live. This immediate feedback can help the archer adjust during practice.

When the archer finish shooting, the archer **close the session** in the app. The system will mark the session as completed and log the end time. All the data, each shot's time, its hit position, which arrow was used, etc. - is now saved for the archer to review later.

The flowchart below illustrates the lifecycle of a session and how shots are recorded in Arch Stats:

```mermaid
    flowchart TB
    classDef user fill:#e3f2fd,stroke:#2196f3,color:#0d47a1,font-weight:bold;
    classDef app fill:#e8f5e9,stroke:#43a047,color:#1b5e20,font-weight:bold;
    classDef sensor fill:#fff3e0,stroke:#ff9800,color:#e65100,font-weight:bold;
    classDef subgraphStyle fill:#f9f9f9,stroke:#cccccc,color:#444;

    START(["Open a session"])
    S1["app: get current datetime<br />as 'start_time' and set<br />'is_opened' to true"]:::app
    U1["user: provides 'location'"]:::user

    S2["app: add a new row<br />in the sessions table and<br />one or more rows in<br />targets table"]:::app
    U4["user: close session"]:::user
    S3["app: get current datetime<br />as 'end_time' and set<br />'is_opened' to false"]:::app
    S4["app: update the<br />session table"]:::app
    ENDS(["Session is closed"])

    subgraph TARGET_CALIBRATION["Target(s) calibration*"]
        class TARGET_CALIBRATION subgraphStyle;
        SU1["user: installs<br />target sensors in butt"]:::user
        SU2["user: setup pin<br />for target face(s) center"]:::user
        SU3["user: setup pin<br />for target face(s) rings"]:::user
        SS1["sensor: reads pins"]:::sensor
        SS2["app: display info back<br />to user through WebUI"]:::app
        VERIFY{"System: Success?"}:::app
        SUCCESS["System: OK"]:::app
        FAIL["System: Retry?"]:::app

        SU1 --> SU2
        SU2 --> SU3
        SU3 --> SS1
        SS1 --> VERIFY
        VERIFY -->|Yes| SUCCESS
        VERIFY -->|No| FAIL
        FAIL -->|Yes| SS1
        SUCCESS --> SS2
    end
    subgraph SHOOTING["User shoots arrows"]
        class SHOOTING subgraphStyle;
        SU4["user: set arrow in bow"]:::user
        HW1["sensor: bow sensor set<br />'arrow_id'<br />'arrow_engage_time'"]:::sensor
        SU5["user: fires arrow"]:::user
        HW2["sensor: bow sensor set<br />'arrow_disengage_time'"]:::sensor
        Q1{"did the target sensor<br />registered an arrow in less<br />than X seconds?"}
        HW3["sensor: target sensor set<br />'arrow_landing_time'<br />'x'<br />'y'"]:::sensor
        SS3["app: collect data<br />and save it into shots table"]:::app
        MISS["Shot missed"]

        SU4 --> HW1
        HW1 --> SU5
        SU5 --> HW2
        HW2 --> Q1
        Q1 -->|Yes| HW3
        Q1 -->|No| MISS
        MISS --> SU4
        HW3 --> SS3
    end

    START --> S1
    S1 --> U1
    U1 --> TARGET_CALIBRATION
    TARGET_CALIBRATION --> S2
    S2 --> SHOOTING
    SHOOTING --> U4
    U4 --> S3
    S3 --> S4
    S4 --> ENDS
```

## Reviewing Your Performance

After the archer's sessions are recorded, Arch Stats really shines in helping the archer make sense of the data. The **dashboard** in the web app gives the archer a clear overview of the archer's performance. The archer can see summary statistics for a session (like number of shots, hits vs. misses, etc.), and key performance indicators such as the archer's average shot spacing or consistency. The data visualization tools let the archer dive deeper:

- **Target Maps:** For each session, the archer can view a scatter plot of the archer's arrow impacts on a target diagram. This shows the archer's grouping and spread, so the archer can identify patterns (for example, are the archer's shots clustering low and left?).
- **Timeline Graphs:** Arch Stats can plot metrics over time - for instance, tracking the archer's average score or group size across all sessions, so the archer can see long-term trends. Did the archer's form change lead to improvement over several weeks? The charts will show the archer.
- **Session Comparison:** The archer can compare one session to another. By filtering the archer's data (e.g. by date range or by the specific bow or arrow used), the archer might discover insights such as *"I shoot more consistently at 18m than at 30m"* or *"My second practice session of the day tends to have tighter groupings than the first."* The app's built-in queries make it easy to get these answers without manual calculations.
- **Arrow-specific Insights:** Because Arch Stats knows each arrow by its ID, the archer can evaluate the performance of individual arrows. For example, the archer might find that **Arrow #7** consistently lands a bit high - indicating it might be slightly different or damaged compared to the archer's others. This level of detail helps the archer fine-tune the archer's equipment and ensure consistency.

And remember, if the archer want to perform any analysis not directly supported in the app, the archer can always export their data and analyze it however they like. **By having all the archer's shots logged and accessible, the archer can base the archer's training decisions on real evidence rather than hunches.** The end result is a clearer understanding of the archer's strengths and weaknesses as an archer, and a record of progress the archer can look back on.

## Future Plans

Arch Stats is an active project, and there are exciting enhancements on the horizon. Planned improvements include:

- **Multi-Archer Support:** In the future, Arch Stats will allow multiple archers to use the system (for example, a coach tracking data for several students, or multiple team members using one setup). This will likely include user profiles or accounts so each archer's data stays separate.
- **Tournament/Scoring Features:** Beyond practice sessions, the system aims to support scoring formats for competitions or scoring rounds. This means the archer could use Arch Stats in a tournament setting to log arrow scores and analyze performance under pressure.
- **Advanced Analytics:** More sophisticated analysis tools are in development, such as automatic grouping size calculations, trend predictions, and integration with scoring zones to calculate scores (if the archer use standard target faces). The goal is to continue providing archers with **actionable insights** that go beyond simple record-keeping.

**Arch Stats is built by an archer, for archers.** It is a passion project designed to bring modern data tracking to the ancient sport of archery. By leveraging technology in a user-friendly way, Arch Stats empowers archers to make informed adjustments to their practice and equipment. Whether a competitive shooter aiming for the podium or a hobbyist trying to beat a personal best, the archer gains the feedback needed to focus training and see real improvement over time.If the you're new, reading both the frontend and backend READMEs will give you a clear roadmap to getting started quickly.- [Arch Stats: track and understand your archery performance over time](#arch-stats-track-and-understand-your-archery-performance-over-time)

- [Arch Stats: track and understand your archery performance over time](#arch-stats-track-and-understand-your-archery-performance-over-time)
  - [Key Features](#key-features)
  - [How It Works](#how-it-works)
    - [Arrow Registration](#arrow-registration)
    - [Sessions and Shots](#sessions-and-shots)
  - [Reviewing Your Performance](#reviewing-your-performance)
  - [Future Plans](#future-plans)
  - [For Software Developers](#for-software-developers)
    - [Backend](#backend)
    - [Frontend](#frontend)
    - [Development Philosophy](#development-philosophy)

## For Software Developers

If the you're a developer interested in contributing to Arch Stats - whether to help expand features, integrate new sensor types, improve performance, or refine the user experience - you're very welcome! While this README focuses on the end-user experience for archers, the system is built as a modular, modern monorepo and is fully open-source.

Arch Stats is divided into two main developer domains:

### Backend

The backend is a Python +3.13 FastAPI application that coordinates data across sessions, arrows, shots, and sensors. It uses asyncpg for PostgreSQL access, Pydantic v2 for schema validation, and WebSockets for real-time communication with the frontend.

For a complete breakdown of the backend modules (including architecture, development workflow, VS Code tasks, testing setup, and API routes), for more information read the [backend/README.md](./backend/README.md)

### Frontend

The frontend is a TypeScript + Vue 3 + Vite application, designed for clarity and performance. It provides real-time visualizations of shot data, arrow/session forms, and performance dashboards. Prettier and ESLint are used for formatting and quality.

To learn how the WebUI is structured, how to develop with hot reload, and how it communicates with the backend, for more information read the [frontend/README.md](./frontend/README.md)

### Development Philosophy

This project is designed for **fast iteration**, **strict type checking**, and **hardware-software integration**. All core services are containerized and the monorepo includes tasks for bringing up everything with a single command. Contributions are welcome, whether the you're improving code, sensors, docs, or analytics!
