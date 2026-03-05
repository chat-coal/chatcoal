# chatcoal

A real-time chat platform built with a Go backend and Vue 3 frontend.

## Tech Stack

- **Backend**: Go + Fiber v2, GORM, MySQL 8, Redis 7, Firebase Auth
- **Frontend**: Vue 3 + Vite, Pinia, Tailwind CSS, Axios
- **Deployment**: Docker Compose + Traefik (auto-HTTPS)

## Single-Server Deployment

### Prerequisites

- Docker and Docker Compose installed
- A domain with DNS A records pointing to the server (four subdomains used by default: `chatcoal.com`, `api.chatcoal.com`, `app.chatcoal.com`, `livekit.chatcoal.com`)
- A Firebase project with Authentication enabled

### 1. Clone and configure

```sh
git clone <repo-url> chatcoal && cd chatcoal
cp .env.example .env
cp docker-compose.yml.example docker-compose.yml
cp livekit.yaml.example livekit.yaml
```

Edit `.env` and fill in:

| Variable | Description |
|---|---|
| `ACME_EMAIL` | Your email for Let's Encrypt TLS certificates |
| `MYSQL_ROOT_PASSWORD` | MySQL root password (change from default) |
| `DB_USER` / `DB_PASS` | MySQL application credentials |
| `FIREBASE_PROJECT_ID` | From Firebase console > Project settings |
| `VITE_FIREBASE_API_KEY` | Firebase client SDK config (all `VITE_FIREBASE_*` vars) |
| `VITE_FIREBASE_AUTH_DOMAIN` | |
| `VITE_FIREBASE_PROJECT_ID` | |
| `VITE_FIREBASE_STORAGE_BUCKET` | |
| `VITE_FIREBASE_MESSAGING_SENDER_ID` | |
| `VITE_FIREBASE_APP_ID` | |

### 2. Add Firebase service account key

Download a service account key from Firebase console > Project settings > Service accounts, and place it at:

```sh
cp ~/path/to/your-key.json backend/firebase-service-account.json
```

### 3. Update domain names

Edit `docker-compose.yml` and replace all occurrences of `chatcoal.com`, `api.chatcoal.com`, `app.chatcoal.com`, and `livekit.chatcoal.com` with your actual domain. The relevant lines are the Traefik host rules, the frontend `VITE_LIVEKIT_URL` build arg, and the backend `APP_DOMAIN` / `APP_ORIGINS` environment variables.

### 4. Start all services

```sh
docker compose up -d
```

This starts seven services:

| Service | Role |
|---|---|
| **Traefik** | Reverse proxy, auto-HTTPS via Let's Encrypt (ports 80/443) |
| **MySQL 8.0** | Database |
| **Redis 7** | Cache and sessions |
| **Backend** | Go API on port 3000 (runs DB migrations automatically on startup) |
| **Frontend** | Vue SPA served by Nginx |
| **Landing** | Marketing site served by Nginx |
| **LiveKit** | Voice channel media server (WSS via Traefik, UDP 50000-50100 for media) |

### 5. Verify

```sh
docker compose ps              # all services should show "Up (healthy)"
docker compose logs backend    # check that migrations ran successfully
```

Visit `https://app.<your-domain>` to access the app.

### Optional Features

| Feature | How to enable |
|---|---|
| **S3 storage** | Set `S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET` in `.env`. Without this, uploads use a local Docker volume. |
| **LiveKit** (voice channels) | Included in the stack. Default keys work out of the box. To use custom keys, update both `.env` and `livekit.yaml`. Ensure UDP ports 50000-50100 and TCP port 7881 are open on your firewall. |
| **Federation** (cross-instance login) | Set `FEDERATION_DOMAIN` to your public hostname. Ed25519 keys are auto-generated on first run. |
| **Monitoring dashboard** | See [monitoring/README.md](monitoring/README.md). Run `cd monitoring && docker compose up -d` to start the dashboard on port 9091. Set `METRICS_TOKEN` on both the backend and monitoring app to secure the endpoint. |

---

## Local Development

### Backend

```sh
cd backend
cp .env.example .env          # configure DB, Redis, Firebase settings
go run ./cmd/migrate -cmd up  # apply database migrations
air                           # run with hot-reload (or: go run server.go)
```

Requires Go 1.21+, a running MySQL and Redis instance, and `firebase-service-account.json` in the backend directory.

### Frontend

```sh
cd frontend
cp .env.example .env          # set VITE_API_URL, VITE_WS_URL, Firebase config
npm install
npm run dev                   # Vite dev server with hot-reload
```

### Database Migrations

Migrations use [goose](https://github.com/pressly/goose) and live in `backend/cmd/migrate/migrations/`.

```sh
cd backend
go run ./cmd/migrate -cmd up       # apply pending migrations
go run ./cmd/migrate -cmd down     # rollback last migration
go run ./cmd/migrate -cmd status   # show migration status
```

### Project Structure

```
backend/
  controllers/    # Request handlers
  models/         # Data models (GORM)
  services/       # Business logic
  routes/         # Route definitions
  database/       # DB connection setup
  middleware/     # Auth, rate limiting
  metrics/        # Internal metrics counters
  ws/             # WebSocket hub
  cmd/migrate/    # Goose migrations

frontend/
  src/
    assets/       # Static assets, styles
    components/   # Reusable Vue components
    router/       # Vue Router config
    services/     # API service layer (Axios)
    stores/       # Pinia state stores
    views/        # Page components
    App.vue       # Root component
    main.js       # Entry point

monitoring/         # Standalone monitoring dashboard
```
