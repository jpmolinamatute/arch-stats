use postgres::{Client, NoTls};
use std::env;

pub fn get_db_conn() -> Client {
    let db_name =
        env::var("ARCH_STATS_USER").expect("ARCH_STATS_USER environment variable is not set");
    let connection_string: String =
        format!("host=/var/run/postgresql name={}", db_name).to_string();

    Client::connect(&connection_string, NoTls).expect("Failed to initialize database connection")
}
