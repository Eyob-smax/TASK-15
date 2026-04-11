# FitCommerce Quick Start

Minimal instructions to run the project locally with Docker.

## Prerequisites

- Docker Desktop (with `docker compose`)

## Startup

From this `repo/` directory:

```bash
docker compose up --build
```

## Access

- Frontend HTTPS: https://localhost:3443
- Frontend HTTP (redirects to HTTPS): http://localhost:3000
- Backend API (TLS): https://localhost:8080
- PostgreSQL: localhost:5432

## Run Tests

From this `repo/` directory:

```bash
./run_tests.sh
```

With coverage:

```bash
./run_tests.sh --coverage
```

## Environment Variables

All variables use the `FC_` prefix. Docker Compose supplies defaults for local development.

| Variable | Default | Description |
|---|---|---|
| `FC_SERVER_PORT` | `8080` | Backend API port |
| `FC_DATABASE_URL` | `postgres://fitcommerce:fitcommerce@localhost:5432/fitcommerce?sslmode=disable` | PostgreSQL connection string |
| `FC_RUN_MIGRATIONS_ON_STARTUP` | `true` | Run goose migrations on boot |
| `FC_RUN_SEED_ON_STARTUP` | `true` | Seed reference data on boot |
| `FC_SESSION_IDLE_TIMEOUT_MINUTES` | `30` | Session idle expiry |
| `FC_SESSION_ABSOLUTE_TIMEOUT_HOURS` | `12` | Session hard expiry |
| `FC_LOGIN_LOCKOUT_THRESHOLD` | `5` | Failed login attempts before lockout |
| `FC_LOGIN_LOCKOUT_DURATION_MINUTES` | `15` | Lockout duration |
| `FC_BACKUP_PATH` | `/var/backups/fitcommerce` | Directory for backup archives |
| `FC_BACKUP_ENCRYPTION_KEY_REF` | _(required for backups)_ | Master key reference for backup file encryption |
| `FC_EXPORT_PATH` | `/tmp/fitcommerce-exports` | Directory for export files |
| `FC_TLS_CERT_FILE` | _(empty)_ | Path to TLS certificate |
| `FC_TLS_KEY_FILE` | _(empty)_ | Path to TLS private key |
| `FC_ALLOW_INSECURE_HTTP` | `false` | Allow plain HTTP (dev only) |
| `FC_BIOMETRIC_MODULE_ENABLED` | `false` | Enable biometric enrollment endpoints |
| `FC_BIOMETRIC_KEY_ROTATION_DAYS` | `90` | Days between biometric key rotations |
| `FC_BIOMETRIC_MASTER_KEY_REF` | _(required when biometric enabled)_ | Master key reference for biometric DEK wrapping |
| `FC_CLUB_TIMEZONE` | `UTC` | Club local timezone for scheduling |
| `FC_LOG_LEVEL` | `info` | Log verbosity (`debug`, `info`, `warn`, `error`) |

> **Security note**: `FC_BACKUP_ENCRYPTION_KEY_REF` and `FC_BIOMETRIC_MASTER_KEY_REF` must be set to non-empty secrets in production. The server will refuse to perform backups or biometric operations if the corresponding key refs are missing.

## Stop

```bash
docker compose down
```

## Clean Reset (optional)

Removes containers, networks, and volumes (deletes local DB data):

```bash
docker compose down -v
```
