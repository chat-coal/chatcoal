# chatcoal Monitoring

Standalone dashboard that polls the backend `/internal/metrics` endpoint and displays real-time charts for WebSocket connections, shard queue depths, and cache hit rates.

## Tracked Metrics

- **WebSocket connections** — active, rejected, and dropped tasks
- **Platform** — total servers and total channels
- **Voice** — active voice channels and connected users (via LiveKit API)
- **Shard queue depths** — per-shard inbox lengths (16 shards)
- **Cache hit rates** — user, member, user_servers, server_channels, presigned URL caches

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `CHATCOAL_METRICS_URL` | `http://localhost:3000/internal/metrics` | Backend metrics endpoint URL |
| `METRICS_TOKEN` | _(empty)_ | Bearer token for authenticating with the backend (must match the backend's `METRICS_TOKEN`) |
| `PORT` | `9091` | Port the dashboard listens on |
| `POLL_INTERVAL` | `5` | Polling interval in seconds |

## Running Locally

```sh
export CHATCOAL_METRICS_URL=http://localhost:3000/internal/metrics
export METRICS_TOKEN=your-secret-token
go run .
```

The dashboard is available at `http://localhost:9091`.

## Running with Docker Compose

```sh
cd monitoring

# Copy and edit the env vars
export CHATCOAL_METRICS_URL=http://host.docker.internal:3000/internal/metrics
export METRICS_TOKEN=your-secret-token

docker compose up -d
```

When running the backend in a separate Docker Compose stack on the same host, use `http://host.docker.internal:3000/internal/metrics` as the metrics URL. If both stacks share a Docker network, use the backend's service name instead (e.g. `http://backend:3000/internal/metrics`).

## Running on a Separate Server

The monitoring app only needs HTTP access to the backend's `/internal/metrics` endpoint. Point `CHATCOAL_METRICS_URL` at the backend's reachable address and set `METRICS_TOKEN` to match.

```sh
export CHATCOAL_METRICS_URL=https://api.example.com/internal/metrics
export METRICS_TOKEN=your-secret-token
docker compose up -d
```

Consider firewalling the `/internal/metrics` path so only the monitoring server can reach it.
