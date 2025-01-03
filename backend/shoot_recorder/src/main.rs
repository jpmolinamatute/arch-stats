use chrono::{DateTime, Utc};
use postgres::{Client, NoTls};
use std::env;
use std::fs;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use uuid::Uuid;

struct SensorData {
    target_track_id: Uuid,
    arrow_engage_time: DateTime<Utc>,
    draw_length: i32,
    arrow_disengage_time: DateTime<Utc>,
    arrow_landing_time: DateTime<Utc>,
    x_coordinate: f32,
    y_coordinate: f32,
}

fn initialize_db_connection() -> Client {
    let db_host = env::var("DB_HOST").ok();
    let db_name = env::var("DB_NAME").ok();
    let db_port = env::var("DB_PORT").ok();
    let db_user = env::var("DB_USER").ok();
    let db_password = env::var("DB_PASSWORD").ok();

    let connection_string = if db_host.is_some()
        && db_name.is_some()
        && db_port.is_some()
        && db_user.is_some()
        && db_password.is_some()
    {
        format!(
            "postgres://{}:{}@{}:{}/{}",
            db_user.unwrap(),
            db_password.unwrap(),
            db_host.unwrap(),
            db_port.unwrap(),
            db_name.unwrap()
        )
    } else {
        "host=/var/run/postgresql".to_string()
    };

    Client::connect(&connection_string, NoTls).expect("Failed to initialize database connection")
}

fn write_to_db(client: &mut Client, data: &SensorData) {
    // @TODO: I don't like this conversion, I need to find a better way to do this.
    // I think we should reuse the same datatype (datetime) in the database and in the code.
    let arrow_engage_time = data.arrow_engage_time.to_rfc3339();
    let arrow_disengage_time = data.arrow_disengage_time.to_rfc3339();
    let arrow_landing_time = data.arrow_landing_time.to_rfc3339();
    let target_track_id = data.target_track_id.to_string();
    client
        .execute(
            "INSERT INTO shooting (target_track_id, arrow_engage_time, draw_length, arrow_disengage_time, arrow_landing_time, x_coordinate, y_coordinate) VALUES ($1, $2, $3, $4, $5, $6, $7)",
            &[
                &target_track_id,
                &arrow_engage_time,
                &data.draw_length,
                &arrow_disengage_time,
                &arrow_landing_time,
                &data.x_coordinate,
                &data.y_coordinate,
            ],
        )
        .expect("Failed to insert data into the database");
}

fn get_uuid() -> Uuid {
    let dir = env::var("ARCH_STATS_DIR").expect("ARCH_STATS_DIR environment variable is not set");
    let file_name =
        env::var("ARCH_STATS_ID_FILE").expect("ARCH_STATS_ID_FILE environment variable is not set");
    let file_path = format!("{}/{}", dir, file_name);
    let content = fs::read_to_string(&file_path).expect("Failed to read the UUID file");

    let uuid = Uuid::parse_str(content.trim()).expect("Invalid UUID format in the file");
    uuid
}

fn read_sensor_data() -> SensorData {
    let now = Utc::now();
    SensorData {
        target_track_id: get_uuid(),
        arrow_engage_time: now,
        draw_length: 30,
        arrow_disengage_time: now,
        arrow_landing_time: now,
        x_coordinate: 10.0,
        y_coordinate: 20.0,
    }
}

fn main() {
    println!("Initializing the shooter recorder...");
    let mut client = initialize_db_connection();
    let running = Arc::new(AtomicBool::new(true));
    let r = running.clone();

    ctrlc::set_handler(move || {
        r.store(false, Ordering::SeqCst);
    })
    .expect("Error setting Ctrl-C handler");

    println!("Press Ctrl+C to exit...");

    while running.load(Ordering::SeqCst) {
        let data = read_sensor_data();
        write_to_db(&mut client, &data);
    }

    println!("Shutting down gracefully...");
}
