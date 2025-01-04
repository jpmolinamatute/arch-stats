use chrono::{DateTime, Utc};
use postgres::{Client, NoTls};
use std::env;
use std::fs;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use uuid::Uuid;

struct SensorData {
    target_track_id: Uuid,
    arrow_id: Uuid,
    arrow_engage_time: DateTime<Utc>,
    draw_length: f32,
    arrow_disengage_time: DateTime<Utc>,
    arrow_landing_time: Option<DateTime<Utc>>,
    x_coordinate: f32,
    y_coordinate: f32,
}

fn initialize_db_connection() -> Client {
    let db_host = env::var("DB_HOST").ok();
    let db_port = env::var("DB_PORT").ok();
    let db_user = env::var("ARCH_STATS_USER").ok();
    let db_password = env::var("DB_PASSWORD").ok();

    let connection_string = if db_host.is_some()
        && db_port.is_some()
        && db_user.is_some()
        && db_password.is_some()
    {
        format!(
            "postgres://{}:{}@{}:{}/{}",
            db_user.as_ref().unwrap(),
            db_password.as_ref().unwrap(),
            db_host.as_ref().unwrap(),
            db_port.as_ref().unwrap(),
            db_user.as_ref().unwrap()
        )
    } else {
        "host=/var/run/postgresql".to_string()
    };
    println!("Connecting to {}", connection_string);
    Client::connect(&connection_string, NoTls).expect("Failed to initialize database connection")
}

fn write_to_db(client: &mut Client, data: &SensorData) -> bool {
    println!("Writing arrow '{}' readings into the database...", data.arrow_id);
    let mut success: bool = true;
    if let Err(e) = client
        .execute(
            "INSERT INTO shooting (target_track_id, arrow_id, arrow_engage_time, draw_length, arrow_disengage_time, arrow_landing_time, x_coordinate, y_coordinate) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
            &[
                &data.target_track_id,
                &data.arrow_id,
                &data.arrow_engage_time,
                &data.draw_length,
                &data.arrow_disengage_time,
                &data.arrow_landing_time,
                &data.x_coordinate,
                &data.y_coordinate,
            
            ],
        ){
            eprintln!("Failed to insert data: {}", e);
            success = false;
        }
        success
}

fn get_target_track_id() -> Uuid {
    let dir = env::var("ARCH_STATS_DIR").expect("ARCH_STATS_DIR environment variable is not set");
    let file_name =
        env::var("ARCH_STATS_ID_FILE").expect("ARCH_STATS_ID_FILE environment variable is not set");
    let file_path = format!("{}/backend/{}", dir, file_name);
    let content = fs::read_to_string(&file_path).expect("Failed to read the UUID file");

    let uuid = Uuid::parse_str(content.trim()).expect("Invalid UUID format in the file");
    uuid
}

fn get_arrow_id() -> Option<Uuid> {
    Some(Uuid::new_v4())
}

fn get_arrow_engage_time() -> Option<DateTime<Utc>> {
    Some(Utc::now())
}

fn get_arrow_disengage_time() -> Option<DateTime<Utc>> {
    Some(Utc::now())
}

fn get_arrow_landing_time() -> Option<DateTime<Utc>> {
    Some(Utc::now())
}

fn get_x_coordinate() -> Option<f32> {
    Some(10.0)
}

fn get_y_coordinate() -> Option<f32> {
    Some(20.0)
}

fn get_draw_length() -> Option<f32> {
    Some(30.0)
}

fn read_sensor_data(target_track_id: Uuid) -> Option<SensorData> {
    let arrow_id = get_arrow_id();
    let arrow_engage_time = get_arrow_engage_time();
    let draw_length = get_draw_length();
    let arrow_disengage_time = get_arrow_disengage_time();
    let x_coordinate = get_x_coordinate();
    let y_coordinate = get_y_coordinate();
    if let (Some(arrow_id), Some(arrow_engage_time), Some(draw_length), Some(arrow_disengage_time), Some(x_coordinate), Some(y_coordinate)) = 
    (arrow_id, arrow_engage_time, draw_length, arrow_disengage_time, x_coordinate, y_coordinate) {
        Some(SensorData {
            target_track_id,
            arrow_id,
            arrow_engage_time,
            draw_length,
            arrow_disengage_time,
            arrow_landing_time: get_arrow_landing_time(),
            x_coordinate,
            y_coordinate,
        })
    } else {
        None
    }
}

fn main() {
    println!("Initializing the shooter recorder...");
    let mut client = initialize_db_connection();
    let running = Arc::new(AtomicBool::new(true));
    let r = running.clone();
    let target_track_id = get_target_track_id();
    ctrlc::set_handler(move || {
        r.store(false, Ordering::SeqCst);
    })
    .expect("Error setting Ctrl-C handler");

    println!("Press Ctrl+C to exit...");

    while running.load(Ordering::SeqCst) {
        if let Some(data) = read_sensor_data(target_track_id) {
            write_to_db(&mut client, &data);
        }
    }

    println!("Shutting down gracefully...");
}
