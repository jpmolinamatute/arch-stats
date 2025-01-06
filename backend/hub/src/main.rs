use postgres::fallible_iterator::FallibleIterator;
use shared::get_db_conn;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
fn main() {
    println!("Initializing the Hub...");
    let running = Arc::new(AtomicBool::new(true));
    let r = running.clone();
    let mut client = get_db_conn();

    client
        .batch_execute("LISTEN shooting_change")
        .expect("Failed to listen to the channel");
    ctrlc::set_handler(move || {
        r.store(false, Ordering::SeqCst);
    })
    .expect("Error setting Ctrl-C handler");

    println!("Press Ctrl+C to exit...");

    while running.load(Ordering::SeqCst) {
        match client.notifications().iter().next() {
            Ok(Some(notification)) => {
                println!("Received notification: {:?}", notification);
            }
            Ok(None) => {
                // No notification received, continue the loop
            }
            Err(err) => {
                eprintln!("Error receiving notification: {:?}", err);
            }
        }
    }

    println!("Shutting down gracefully...");
}
