use postgres::{Client, NoTls};
use std::env;

pub fn get_db_conn() -> Client {
    let dbname =
        env::var("ARCH_STATS_USER").expect("ARCH_STATS_USER environment variable is not set");
    let dev = env::var("DEV").ok();
    let connection_string = if dev.is_some() {
        format!("host=/var/run/postgresql dbname={} user={}", dbname, dbname).to_string()
    } else {
        format!("host=/var/run/postgresql").to_string()
    };

    Client::connect(&connection_string, NoTls).expect("Failed to initialize database connection")
}
