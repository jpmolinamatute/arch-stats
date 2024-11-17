use dotenv::dotenv;
use postgres::{Client, Error, NoTls};
use serde::Serialize;
use signal_hook::consts::signal::*;
use signal_hook::flag;
use std::env;
use std::net::UdpSocket;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use uuid::Uuid;

#[derive(Serialize)]
struct Shot {
    id: Uuid,
    lane_sensor_id: Uuid,
    bow_sensor_id: Uuid,
    arrow_engage_time: u32,
    arrow_disengage_time: u32,
    arrow_landing_time: u32,
    x_coordinate: f32,
    y_coordinate: f32,
    pull_length: f32,
    distance: f32,
    arrow_id: Uuid,
}

fn write_to_socket(
    socket: &UdpSocket,
    shot_id: Uuid,
    lane_sensor_id: Uuid,
    bow_sensor_id: Uuid,
    arrow_engage_time: u32,
    arrow_disengage_time: u32,
    arrow_landing_time: u32,
    x_coordinate: f32,
    y_coordinate: f32,
    pull_length: f32,
    distance: f32,
    arrow_id: Uuid,
) {
    println!("Inserted value into the database");

    let shot = Shot {
        id: shot_id,
        lane_sensor_id,
        bow_sensor_id,
        arrow_engage_time,
        arrow_disengage_time,
        arrow_landing_time,
        x_coordinate,
        y_coordinate,
        pull_length,
        distance,
        arrow_id,
    };

    let message = serde_json::to_string(&shot).expect("Failed to serialize shot");
    socket
        .send(message.as_bytes())
        .expect("Failed to send message");
}

fn write_shot(
    client: &mut Client,
    socket: &UdpSocket,
    lane_sensor_id: Uuid,
    bow_sensor_id: Uuid,
    arrow_engage_time: u32,
    arrow_disengage_time: u32,
    arrow_landing_time: u32,
    x_coordinate: f32,
    y_coordinate: f32,
    pull_length: f32,
    distance: f32,
    arrow_id: Uuid,
) {
    let shot_id = Uuid::new_v4();
    match client.execute(
        "INSERT INTO shots (id, lane_sensor_id, bow_sensor_id, arrow_engage_time, arrow_disengage_time, arrow_landing_time, x_coordinate, y_coordinate, pull_length, distance, arrow_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
        &[
            &shot_id,
            &lane_sensor_id,
            &bow_sensor_id,
            &arrow_engage_time,
            &arrow_disengage_time,
            &arrow_landing_time,
            &x_coordinate,
            &y_coordinate,
            &pull_length,
            &distance,
            &arrow_id,
        ],
    ) {
        Ok(_) => {
            println!("Inserted value into the database");
            write_to_socket(
                socket,
                shot_id,
                lane_sensor_id,
                bow_sensor_id,
                arrow_engage_time,
                arrow_disengage_time,
                arrow_landing_time,
                x_coordinate,
                y_coordinate,
                pull_length,
                distance,
                arrow_id,
            );
        },
        Err(e) => eprintln!("Failed to insert value into the database: {}", e),
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

fn listen(terminate: Arc<AtomicBool>, mut client: Client, socket: UdpSocket) {
    while !terminate.load(Ordering::Relaxed) {
        println!("Listening!");
        let lane_sensor_id_str = env::var("LANE_SENSOR_ID").expect("LANE_SENSOR_ID not set");
        let lane_sensor_id = Uuid::parse_str(&lane_sensor_id_str).expect("Invalid UUID format");

        // Example values for the parameters
        let bow_sensor_id = Uuid::new_v4();
        let arrow_engage_time: u32 = 123;
        let arrow_disengage_time: u32 = 456;
        let arrow_landing_time: u32 = 789;
        let x_coordinate = 1.23;
        let y_coordinate = 4.56;
        let pull_length = 7.89;
        let distance = 10.11;
        let arrow_id = Uuid::new_v4();

        // Call the write_shot function
        write_shot(
            &mut client,
            &socket,
            lane_sensor_id,
            bow_sensor_id,
            arrow_engage_time,
            arrow_disengage_time,
            arrow_landing_time,
            x_coordinate,
            y_coordinate,
            pull_length,
            distance,
            arrow_id,
        );
    }
    // Attempt to close the connection gracefully
    if let Err(e) = client.close() {
        eprintln!("Failed to close the database connection gracefully: {}", e);
    } else {
        println!("Database connection closed gracefully.");
    }

    println!("Terminating gracefully.");
}

fn main() {
    dotenv().ok();
    let terminate = Arc::new(AtomicBool::new(false));
    let term_clone = Arc::clone(&terminate);
    let socket = UdpSocket::bind("127.0.0.1:0").expect("Failed to bind socket");
    socket
        .connect("127.0.0.1:34254")
        .expect("Failed to connect socket");
    // Register signal handlers
    flag::register(SIGTERM, term_clone.clone()).expect("Failed to register SIGTERM handler");
    flag::register(SIGINT, term_clone.clone()).expect("Failed to register SIGINT handler");
    match connect_to_postgres() {
        Ok(client) => {
            listen(terminate, client, socket);
        }
        Err(e) => eprintln!("Failed to connect to PostgreSQL: {}", e),
    }
}
