# chatcoal

## Database Migrations

Migrations use [goose](https://github.com/pressly/goose). AutoMigrate has been removed.

**Location:** `backend/cmd/migrate/migrations/*.sql`

**Commands** (run from `backend/`):
```sh
go run ./cmd/migrate -cmd up       # apply pending migrations
go run ./cmd/migrate -cmd down     # rollback last migration
go run ./cmd/migrate -cmd status   # show applied/pending
go run ./cmd/migrate -cmd version  # current version
go run ./cmd/migrate -cmd reset    # rollback all
```

**Adding a new migration:**
1. Create `backend/cmd/migrate/migrations/NNN_description.sql`
2. Use `-- +goose Up` and `-- +goose Down` markers
3. Run `go run ./cmd/migrate -cmd up` before starting the server

Migrations are embedded in the binary via `//go:embed` — no external files needed at runtime.

## Frontend Conventions

- **Confirmations**: Use styled modals (Teleport to body) instead of `confirm()` / `alert()`. Follow the existing modal pattern: backdrop with `backdrop-blur-sm`, `rounded-2xl` card, Cancel + action buttons.
