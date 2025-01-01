extern crate dotenv;
extern crate uuid;

use dotenv::dotenv;
use std::env;
use std::fs;
use std::path::Path;

fn check_dir() {
    let dir_path = env::var("ARCH_STATS_DIR").expect("ARCH_STATS_DIR environment variable not set");
    let path = Path::new(&dir_path);

    // Check if the path exists and if it is a directory
    if path.exists() {
        if path.is_dir() {
            println!("The path '{}' exists and is a directory.", dir_path);
        } else {
            panic!("The path '{}' exists but is not a directory.", dir_path);
        }
    } else {
        // Create the directory if it doesn't exist
        match fs::create_dir_all(&dir_path) {
            Ok(_) => println!("Successfully created the directory '{}'.", dir_path),
            Err(e) => eprintln!("Failed to create the directory '{}': {}", dir_path, e),
        }
    }
}

fn write_id_file() {
    let id = uuid::Uuid::new_v4();
    let id_str = id.to_string();
    let dir_path = env::var("ARCH_STATS_DIR").expect("ARCH_STATS_DIR environment variable not set");
    let file_name =
        env::var("ARCH_STATS_ID_FILE").expect("ARCH_STATS_ID_FILE environment variable not set");
    let full_path = Path::new(&dir_path).join(file_name);
    if full_path.exists() {
        println!(
            "The file '{}' already exists and will not be overwritten.",
            full_path.display()
        );
    } else {
        match fs::write(&full_path, id_str) {
            Ok(_) => println!("Successfully wrote the UUID to '{}'.", full_path.display()),
            Err(e) => eprintln!(
                "Failed to write the UUID to '{}': {}",
                full_path.display(),
                e
            ),
        }
    }
}

fn main() {
    dotenv().ok();
    check_dir();
    write_id_file();
}
