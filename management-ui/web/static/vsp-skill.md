# Local VSP — Developer Guide

This guide explains how to write and deploy applications on a Local VSP server.

---

## 1. What is Local VSP?

Local VSP (Virtual Server Platform) is a self-hosted platform that runs web applications and static websites on a local server using Docker. It provides:

- **Auto-detection** of Node.js, Python, Go, PHP, Rust, and static HTML projects
- **Automatic Dockerfile and docker-compose.yml generation**
- **One-click deploy** with real-time build output
- **Traefik reverse proxy** for subdomain-based routing
- **Cloudflare Tunnel** integration for internet access
- **Gitea** self-hosted Git server

The platform is managed through a web dashboard — no SSH or command-line access needed for regular deployments.

---

## 2. Where to Put Your Files

Applications are deployed via a Samba (Windows) network share:

```
\\<server-ip>\vsp-home\docker\your-app-name\
```

On the server, this maps to:

```
/home/vsp/docker/your-app-name/
```

Each subfolder in the `docker` share is treated as a separate project. The folder name becomes the project name shown in the dashboard.

For static HTML websites, use:

```
\\<server-ip>\vsp-home\html\your-site-name\
```

---

## 3. Supported Project Types

### Node.js (auto-detected)

**Trigger file:** `package.json`

When the system finds a `package.json` (and no existing `docker-compose.yml`), it auto-generates:

**Dockerfile:**
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3000
CMD ["npm", "start"]
```

**Requirements:**
- Your app **must** listen on port **3000**
- `npm start` must start a long-running web server (not a CLI script that exits)
- The `start` script in `package.json` should run your server entry point

### Python (auto-detected)

**Trigger file:** `requirements.txt`

When the system finds a `requirements.txt` (and no existing `docker-compose.yml`), it auto-generates:

**Dockerfile:**
```dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8000
CMD ["python", "app.py"]
```

**Requirements:**
- Your entry point must be named **`app.py`**
- Your app **must** listen on port **8000**
- The process must stay running (a web server, not a script that exits)

### Go (auto-detected)

**Trigger file:** `go.mod`

When the system finds a `go.mod` (and no existing `docker-compose.yml`), it auto-generates a multi-stage Dockerfile:

**Dockerfile:**
```dockerfile
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server .

FROM alpine:3.19
WORKDIR /app
COPY --from=build /app/server .
EXPOSE 8080
CMD ["./server"]
```

**Requirements:**
- Your app **must** listen on port **8080**
- The `main` package must be in the project root (where `go.mod` is)
- Include `go.sum` alongside `go.mod`
- Bind to **0.0.0.0** (not 127.0.0.1)

### PHP (auto-detected)

**Trigger file:** `composer.json`

When the system finds a `composer.json` (and no existing `docker-compose.yml`), it auto-generates:

**Dockerfile:**
```dockerfile
FROM composer:2 AS deps
WORKDIR /app
COPY composer.json composer.lock* ./
RUN composer install --no-dev --no-scripts --ignore-platform-reqs

FROM php:8.3-apache
RUN docker-php-ext-install pdo pdo_mysql
COPY . /var/www/html/
COPY --from=deps /app/vendor /var/www/html/vendor
EXPOSE 80
```

**Requirements:**
- Your document root is the project folder (served from `/var/www/html/`)
- Apache serves on port **80** (default)
- PDO and PDO MySQL extensions are pre-installed
- Composer dependencies are installed automatically

### Rust (auto-detected)

**Trigger file:** `Cargo.toml`

When the system finds a `Cargo.toml` (and no existing `docker-compose.yml`), it auto-generates a multi-stage Dockerfile:

**Dockerfile:**
```dockerfile
FROM rust:1.77 AS build
WORKDIR /app
COPY . .
RUN cargo build --release
RUN cp target/release/$(cargo metadata --format-version=1 --no-deps \
    | grep -o '"name":"[^"]*"' | head -1 | cut -d'"' -f4) /app/server \
    || cp target/release/app /app/server

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /app/server .
EXPOSE 8080
CMD ["./server"]
```

**Requirements:**
- Your app **must** listen on port **8080**
- The binary name is auto-detected from `Cargo.toml`
- Bind to **0.0.0.0** (not 127.0.0.1)
- The first build will be slow (compiling all dependencies); rebuilds are faster

### Static HTML (auto-detected)

**Trigger file:** `index.html`

When the system finds an `index.html` (and no other trigger files or `docker-compose.yml`), it auto-generates:

**Dockerfile:**
```dockerfile
FROM nginx:alpine
COPY . /usr/share/nginx/html/
EXPOSE 80
```

**Requirements:**
- Your folder must contain an `index.html` file
- All assets (CSS, JS, images) are served as static files
- Nginx serves on port **80**

**Note:** This is checked last in the detection order. If your project also contains a `package.json`, `go.mod`, etc., those will take priority over `index.html`.

### Custom (docker-compose.yml)

For other languages, frameworks, or more complex setups, provide your own `docker-compose.yml` and optionally a `Dockerfile`.

The system will use your compose file directly. You can also include a `Dockerfile` referenced by `build: .` in your compose file.

---

## 4. How Auto-Detection Works

When you click **Deploy** in the dashboard, the system checks your project folder in this order:

1. If `docker-compose.yml` already exists → use it as-is
2. If `package.json` exists → generate Node.js Dockerfile + docker-compose.yml
3. If `requirements.txt` exists → generate Python Dockerfile + docker-compose.yml
4. If `go.mod` exists → generate Go Dockerfile + docker-compose.yml
5. If `composer.json` exists → generate PHP Dockerfile + docker-compose.yml
6. If `Cargo.toml` exists → generate Rust Dockerfile + docker-compose.yml
7. If `index.html` exists → generate Static HTML Dockerfile + docker-compose.yml
8. If none of the above → deploy fails with an error message

Auto-generated files are written to the project folder. They are **not** overwritten if they already exist — if you need to regenerate them, delete the existing `Dockerfile` and `docker-compose.yml` first.

---

## 5. Writing a Node.js App

### Minimal example (Express)

**package.json:**
```json
{
  "name": "my-app",
  "version": "1.0.0",
  "scripts": {
    "start": "node server.js"
  },
  "dependencies": {
    "express": "^4.18.0"
  }
}
```

**server.js:**
```javascript
const express = require('express');
const app = express();

app.get('/', (req, res) => {
  res.send('Hello from Local VSP!');
});

app.listen(3000, () => {
  console.log('Server running on port 3000');
});
```

### Key points:
- Always listen on **port 3000**
- The process must **not** exit — it must stay running as a server
- Use `npm start` as the entry point (configured in package.json `scripts.start`)
- If you need environment variables, create a `.env` file in your project folder

---

## 6. Writing a Python App

### Minimal example (Flask)

**requirements.txt:**
```
flask
```

**app.py:**
```python
from flask import Flask

app = Flask(__name__)

@app.route('/')
def hello():
    return '<h1>Hello from Local VSP!</h1>'

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8000)
```

### Key points:
- Always listen on **port 8000**
- Bind to **0.0.0.0** (not 127.0.0.1), otherwise Docker can't reach your app
- The entry point must be named **`app.py`**
- The process must **not** exit — it must stay running

---

## 7. Writing a Go App

### Minimal example (net/http)

**go.mod:**
```
module my-app

go 1.22
```

**main.go:**
```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, "<h1>Hello from Local VSP!</h1>")
    })
    fmt.Println("Server running on port 8080")
    http.ListenAndServe("0.0.0.0:8080", nil)
}
```

### Key points:
- Always listen on **port 8080**
- Bind to **0.0.0.0**
- The `main` package must be in the project root
- Include both `go.mod` and `go.sum` (run `go mod tidy` before deploying)
- Multi-stage build keeps the final image small (~15 MB)

---

## 8. Writing a PHP App

### Minimal example (plain PHP)

**composer.json:**
```json
{
  "name": "my-app",
  "require": {}
}
```

**index.php:**
```php
<?php
echo '<h1>Hello from Local VSP!</h1>';
echo '<p>Server time: ' . date('Y-m-d H:i:s') . '</p>';
?>
```

### Minimal example (with a framework — Laravel-style)

For Laravel or other frameworks, include the full project with `composer.json`. Dependencies are installed automatically via the multi-stage Composer build.

### Key points:
- Apache serves files from your project root
- Port **80** is used (Apache default)
- PDO and PDO MySQL extensions are pre-installed for database access
- Add more PHP extensions in a custom Dockerfile if needed

---

## 9. Writing a Rust App

### Minimal example (Actix Web)

**Cargo.toml:**
```toml
[package]
name = "my-app"
version = "0.1.0"
edition = "2021"

[dependencies]
actix-web = "4"
```

**src/main.rs:**
```rust
use actix_web::{web, App, HttpServer, HttpResponse};

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    println!("Server running on port 8080");
    HttpServer::new(|| {
        App::new()
            .route("/", web::get().to(|| async {
                HttpResponse::Ok().body("<h1>Hello from Local VSP!</h1>")
            }))
    })
    .bind("0.0.0.0:8080")?
    .run()
    .await
}
```

### Key points:
- Always listen on **port 8080**
- Bind to **0.0.0.0**
- First build is slow (compiles all dependencies from source)
- The binary name is auto-detected from `Cargo.toml`

---

## 10. Custom docker-compose.yml

For any language or framework, you can write your own configuration.

### Example: Multi-service app (app + database)

**docker-compose.yml:**
```yaml
services:
  web:
    build: .
    restart: unless-stopped
    depends_on:
      - db
    environment:
      DATABASE_URL: postgres://user:pass@db:5432/mydb

  db:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: mydb
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

### Important notes:
- Use `EXPOSE` in your Dockerfile to declare the port your app listens on. The system reads this to configure port mapping.
- The first service in the compose file is the one that gets Traefik routing labels.
- If your compose file has no `EXPOSE` and the system can't detect the port, it defaults to port 80.

---

## 11. Port & Subdomain Configuration

### Port allocation

- **Custom Apps:** Ports are auto-allocated in range **8300–8399**
- **Static Websites:** Ports are auto-allocated in range **8200–8299**
- Port assignments are persistent and stored in `/vsp-home/.localvsp/autostart.json`

Your app is immediately accessible at `http://<server-ip>:<port>` on the local network.

### Subdomain configuration

To access an app via a domain name (e.g., `my-app.example.com`):

1. Configure a **Cloudflare Tunnel Token** and **Base Domain** in the Settings page
2. In the dashboard, type a subdomain name into the **Web Address** field next to your app
3. The system generates a `docker-compose.override.yml` with Traefik routing labels

The override file connects the service to the `vsp-network` Docker network and adds labels like:

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.vsp-my-app.rule=Host(`my-app.example.com`)"
  - "traefik.http.services.vsp-my-app.loadbalancer.server.port=3000"
```

---

## 12. Build & Deploy Process

When you click **Deploy** in the dashboard, the following happens:

1. **Auto-generation** — If needed, Dockerfile and docker-compose.yml are generated
2. **Port allocation** — A free host port is found and assigned
3. **Override generation** — `docker-compose.override.yml` is created with Traefik labels (if subdomain is set)
4. **Build** — `docker compose build --no-cache` runs (output streamed to Build Panel)
5. **Start** — `docker compose up --detach --remove-orphans`
6. **Health check** — After 3 seconds, the system verifies all services are running
7. **Restart policy** — Set to `unless-stopped` (if Auto-Start is on) or `no`

If the health check fails (containers exited), the build is marked as **failed** and the last 30 lines of container logs are captured in the Build Panel.

Build timeout is **10 minutes**.

---

## 13. Debugging

### Build Panel

The slide-out Build Panel shows real-time output during deployment. After a deploy finishes (success or failure), click **View Build** to reopen the log. Build logs are persistent — they survive page refreshes and server restarts.

### Runtime Logs

For a running app, click **Logs** in the Custom Apps table (or navigate to the Logs page). This shows live `docker compose logs` output with auto-refresh every 5 seconds.

### Common errors

| Problem | Cause | Fix |
|---------|-------|-----|
| "containers exited after start" | App crashed immediately | Check Build Panel for container output. Ensure your app stays running. |
| Build succeeds but app shows "Stopped" | App exits after printing output | Your app must be a long-running server, not a script. |
| Port not accessible | App binds to 127.0.0.1 | Bind to `0.0.0.0` so Docker networking can reach it. |
| "could not auto-detect project type" | No recognized trigger file | Add one of the supported files, or provide your own docker-compose.yml. |
| Build times out | Slow build or hanging process | Check your Dockerfile for slow steps. Build timeout is 10 minutes. |
| Go/Rust build is very slow | First build downloads all dependencies | Subsequent rebuilds will be faster. Consider using a custom Dockerfile with caching. |

---

## 14. Gitea Integration

Local VSP includes Gitea, a self-hosted Git server accessible at:

```
http://<server-ip>:3000
```

You can:
- Create Git repositories to store your source code
- Push code from your development machine
- Use Gitea as a private package registry

Gitea runs independently from the deploy system — it doesn't auto-deploy on push. To deploy, copy your files to the network share and click Deploy in the dashboard.

---

## 15. Tips & Best Practices

### Auto-Start
Enable **Auto-Start** for production apps so they restart automatically after a server reboot or power outage. This sets the Docker restart policy to `unless-stopped`.

### Environment variables
Place a `.env` file in your project folder. Docker Compose automatically reads it. For sensitive values, the file is on the server's local disk (not exposed to the network share).

### Updating your app
1. Copy updated files to the network share (overwrite existing files)
2. Click **Deploy** (or **Redeploy**) in the dashboard
3. The system rebuilds from scratch (`--no-cache`) and restarts

### Multi-service apps
You can run databases, caches, or other services alongside your app in a single `docker-compose.yml`. Use Docker volumes to persist data.

### File structure example

```
\\server-ip\vsp-home\docker\
  my-node-app\
    package.json
    server.js
    src\
      ...
  my-python-app\
    requirements.txt
    app.py
  my-go-app\
    go.mod
    go.sum
    main.go
  my-php-app\
    composer.json
    index.php
  my-rust-app\
    Cargo.toml
    src\
      main.rs
  my-custom-app\
    Dockerfile
    docker-compose.yml
    src\
      ...
```

### What NOT to include
- Don't include `node_modules/`, `__pycache__/`, `vendor/`, or `target/` — these are installed during the Docker build
- Don't include `.git/` directories — they add unnecessary size
- Don't include large data files — use Docker volumes instead

---

## Quick Reference

| Feature | Node.js | Python | Go | PHP | Rust | Static HTML | Custom |
|---------|---------|--------|----|----|------|-------------|--------|
| Trigger file | `package.json` | `requirements.txt` | `go.mod` | `composer.json` | `Cargo.toml` | `index.html` | `docker-compose.yml` |
| Default port | 3000 | 8000 | 8080 | 80 | 8080 | 80 | From `EXPOSE` or 80 |
| Entry point | `npm start` | `python app.py` | `./server` | Apache | `./server` | Nginx | Your CMD |
| Base image | node:18-alpine | python:3.11-slim | golang:1.22-alpine | php:8.3-apache | rust:1.77 | nginx:alpine | Your choice |
| Multi-stage | No | No | Yes | Yes | Yes | No | Optional |
| Auto-generated | Yes | Yes | Yes | Yes | Yes | Yes | No |
