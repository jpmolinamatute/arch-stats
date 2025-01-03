use actix_files as fs;
use actix_web::{App, HttpServer};
use std::env;
use std::path::Path;

fn get_webui_path() -> String {
    let webui_path = env::var("ARCH_STATS_WEBUI_PATH").expect("ARCH_STATS_WEBUI_PATH must be set");
    if Path::new(&webui_path).is_dir() {
        webui_path
    } else {
        panic!(
            "The path specified in 
         does not exist or is not a directory"
        );
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let webui_path = get_webui_path();
    HttpServer::new(move || {
        App::new().service(fs::Files::new("/", &webui_path).index_file("index.html"))
    })
    .bind("127.0.0.1:8080")?
    .run()
    .await
}
