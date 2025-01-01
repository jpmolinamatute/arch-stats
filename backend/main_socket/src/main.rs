use futures_util::{stream::StreamExt, SinkExt};
use std::env;
use std::fs;
use std::io::Read;
use std::os::unix::net::UnixListener;
use std::str;
use tokio::net::TcpListener;
use tokio::sync::broadcast;
use tokio_tungstenite::accept_async;

#[tokio::main]
async fn main() -> std::io::Result<()> {
    let socket_path_str =
        env::var("ARCH_STATS_SOCKET").expect("ARCH_STATS_SOCKET environment variable not set");
    let socket_path = std::path::Path::new(&socket_path_str);

    // Remove the socket file if it already exists
    if fs::metadata(socket_path).is_ok() {
        fs::remove_file(socket_path)?;
    }

    // Create a UnixListener at the specified path
    let listener = UnixListener::bind(socket_path)?;

    println!("Listening on {}", socket_path_str);

    // Create a broadcast channel for WebSocket clients
    let (tx, _rx) = broadcast::channel(100);

    // Spawn a task to handle WebSocket connections
    tokio::spawn(async move {
        let ws_listener_str =
            env::var("ARCH_STATS_WS").expect("ARCH_STATS_WS environment variable not set");
        let ws_listener = TcpListener::bind(ws_listener_str).await.unwrap();
        println!("WebSocket server listening on ws://{}", ws_listener_str);

        while let Ok((stream, _)) = ws_listener.accept().await {
            let tx = tx.clone();
            tokio::spawn(async move {
                let ws_stream = accept_async(stream).await.unwrap();
                let (mut ws_sender, mut ws_receiver) = ws_stream.split();

                let mut rx = tx.subscribe();
                tokio::spawn(async move {
                    while let Ok(msg) = rx.recv().await {
                        ws_sender
                            .send(tokio_tungstenite::tungstenite::Message::Text(msg))
                            .await
                            .unwrap();
                    }
                });

                while let Some(Ok(_)) = ws_receiver.next().await {}
            });
        }
    });

    // Accept connections and process them serially
    for stream in listener.incoming() {
        match stream {
            Ok(mut stream) => {
                let mut buffer = [0; 1024];
                match stream.read(&mut buffer) {
                    Ok(size) => {
                        let received_message = match str::from_utf8(&buffer[..size]) {
                            Ok(v) => v,
                            Err(e) => {
                                eprintln!("Invalid UTF-8 sequence: {}", e);
                                continue;
                            }
                        };
                        println!("Received: {}", received_message);

                        // Send the message to WebSocket clients
                        tx.send(received_message).unwrap();
                    }
                    Err(err) => {
                        eprintln!("Failed to read from socket: {}", err);
                    }
                }
            }
            Err(err) => {
                eprintln!("Failed to accept connection: {}", err);
            }
        }
    }

    Ok(())
}
