use std::io::{Read, Write};
use std::net::TcpListener;
use std::process::Command;
use std::time::Instant;

fn main() {
    let start = Instant::now();
    let listener = TcpListener::bind("0.0.0.0:8080").expect("Failed to bind to port 8080");
    println!("hello-rust listening on port 8080");

    for stream in listener.incoming() {
        let mut stream = match stream {
            Ok(s) => s,
            Err(_) => continue,
        };

        let mut buf = [0u8; 1024];
        let _ = stream.read(&mut buf);

        let hostname = Command::new("hostname")
            .output()
            .map(|o| String::from_utf8_lossy(&o.stdout).trim().to_string())
            .unwrap_or_else(|_| "unknown".to_string());
        let uptime = start.elapsed().as_secs();
        let version = env!("CARGO_PKG_VERSION");

        let body = format!(
            r#"<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Hello Rust</title>
  <style>
    body {{ background: #0f172a; color: #e2e8f0; font-family: monospace;
           display: flex; align-items: center; justify-content: center;
           height: 100vh; margin: 0; }}
    .box {{ text-align: center; }}
    h1 {{ font-size: 3rem; color: #f472b6; margin-bottom: 0.5rem; }}
    p  {{ color: #94a3b8; }}
  </style>
</head>
<body>
  <div class="box">
    <h1>&#9632; Hello, Rust!</h1>
    <p>Running on Local VSP &mdash; Rust (hello-rust v{})</p>
    <p style="margin-top:1rem;font-size:0.75rem;color:#475569">
      hostname: {} &nbsp;|&nbsp; uptime: {}s
    </p>
  </div>
</body>
</html>"#,
            version, hostname, uptime
        );

        let response = format!(
            "HTTP/1.1 200 OK\r\nContent-Type: text/html; charset=utf-8\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
            body.len(),
            body
        );

        let _ = stream.write_all(response.as_bytes());
    }
}
