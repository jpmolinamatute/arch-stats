use dotenv::dotenv;
use postgres::{Client, Error, NoTls};
use serde::Serialize;
use signal_hook::consts::signal::*;
use signal_hook::flag;
use std::env;
use std::fs;
use std::os::unix::net::UnixDatagram;
use std::path::PathBuf;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use uuid::Uuid;

#[derive(Serialize)]
struct Shot {
    id: Option<Uuid>,
    arrow_engage_time: i64,
    arrow_disengage_time: i64,
    arrow_landing_time: i64,
    x_coordinate: f32,
    y_coordinate: f32,
    pull_length: f32,
    distance: f32,
    arrow_id: Uuid,
}

fn write_to_socket(socket: &UnixDatagram, socket_path: &PathBuf, shot: &Shot) {
    if fs::metadata(socket_path).is_ok() {
        let message = serde_json::to_string(&shot).expect("Failed to serialize shot");
        match socket.send_to(message.as_bytes(), socket_path) {
            Ok(_) => println!("Message sent successfully"),
            Err(err) => eprintln!("Failed to send message: {}", err),
        }
    } else {
        println!("Socket path does not exist yet");
    }
}

fn write_shot(client: &mut Client, shot: Shot) -> Option<Shot> {
    match client.query_one(
        "INSERT INTO shots (arrow_engage_time, arrow_disengage_time, arrow_landing_time, x_coordinate, y_coordinate, pull_length, distance, arrow_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
        &[
            &shot.arrow_engage_time,
            &shot.arrow_disengage_time,
            &shot.arrow_landing_time,
            &shot.x_coordinate,
            &shot.y_coordinate,
            &shot.pull_length,
            &shot.distance,
            &shot.arrow_id,
        ],
    ) {
        Ok(row) => {
            let id: Uuid = row.get("id");
            println!("Inserted value into the database with ID: {}", id);
            Some(Shot { id: Some(id), ..shot })
        },
        Err(e) => {
            eprintln!("Failed to insert value into the database: {}", e);
            None
        },
    }
}

fn connect_to_postgres() -> Result<Client, Error> {
    let user = env::var("POSTGRES_USER").expect("POSTGRES_USER not set");
    let password = env::var("POSTGRES_PASSWORD").expect("POSTGRES_PASSWORD not set");
    let dbname = env::var("POSTGRES_DB").expect("POSTGRES_DB not set");

    let config = format!(
        "host=localhost user={} password={} dbname={}",
        user, password, dbname
    );
    let client = Client::connect(&config, NoTls)?;

    println!("Connected to PostgreSQL database");

    Ok(client)
}

fn listen(
    terminate: Arc<AtomicBool>,
    mut client: Client,
    socket: UnixDatagram,
    socket_path: PathBuf,
) {
    while !terminate.load(Ordering::Relaxed) {
        println!("Listening!");

        let shot = Shot {
            id: None,
            arrow_engage_time: 123,
            arrow_disengage_time: 456,
            arrow_landing_time: 789,
            x_coordinate: 1.23,
            y_coordinate: 4.56,
            pull_length: 7.89,
            distance: 10.11,
            arrow_id: Uuid::parse_str("4c19ead8-9b49-4876-978c-f22b5ec5edbf")
                .expect("Failed to parse UUID"),
        };

        // Call the write_shot function
        if let Some(inserted_shot) = write_shot(&mut client, shot) {
            write_to_socket(&socket, &socket_path, &inserted_shot);
        }
    }
    // Attempt to close the connection gracefully
    if let Err(e) = client.close() {
        eprintln!("Failed to close the database connection gracefully: {}", e);
    } else {
        println!("Database connection closed gracefully.");
    }
    println!("Closing socket");
    drop(socket);
    println!("Terminating gracefully.");
}

fn main() {
    dotenv().ok();
    let terminate = Arc::new(AtomicBool::new(false));
    let term_clone = Arc::clone(&terminate);
    let socket_path = PathBuf::from("/tmp/my_socket");
    let socket = UnixDatagram::unbound().expect("Failed to create socket");

    // Register signal handlers
    flag::register(SIGTERM, term_clone.clone()).expect("Failed to register SIGTERM handler");
    flag::register(SIGINT, term_clone.clone()).expect("Failed to register SIGINT handler");
    match connect_to_postgres() {
        Ok(client) => {
            listen(terminate, client, socket, socket_path);
        }
        Err(e) => eprintln!("Failed to connect to PostgreSQL: {}", e),
    }
}
