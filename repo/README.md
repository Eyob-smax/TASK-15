# FitCommerce Quick Start

Minimal instructions to run the project locally with Docker.

Project Type: Fullstack (frontend + backend + PostgreSQL)

## Prerequisites

- Docker Desktop (with `docker compose`)

## Startup

From this `repo/` directory:

```bash
docker-compose up
```

Recommended (rebuild images when code changes):

```bash
docker compose up --build
```

## Access

- Frontend HTTPS: https://localhost:3443
- Frontend HTTP (redirects to HTTPS): http://localhost:3000
- Backend API (TLS): https://localhost:8080
- PostgreSQL: localhost:5432

## Demo Credentials

Authentication is required.

Use the following demo accounts for role-based verification:

| Role | Email | Password |
|---|---|---|
| Administrator | admin.demo@fitcommerce.local | Password123! |
| Operations Manager | ops.demo@fitcommerce.local | Password123! |
| Procurement Specialist | procurement.demo@fitcommerce.local | Password123! |
| Coach | coach.demo@fitcommerce.local | Password123! |
| Member | member.demo@fitcommerce.local | Password123! |

Notes:
- The default SQL seed file in `backend/database/seeds/seed.sql` only seeds system metadata and an inactive system actor.
- If your local environment does not include the demo users above, provision them in your environment-specific seed/bootstrap process before QA/UAT.

## Verification Method

After startup, verify both API and UI behavior.

### API Verification (curl)

1) Login and capture session cookie:

```bash
curl -k -i -c cookies.txt \
	-H "Content-Type: application/json" \
	-d '{"email":"admin.demo@fitcommerce.local","password":"Password123!"}' \
	https://localhost:8080/api/v1/auth/login
```

Expected:
- HTTP 200
- `Set-Cookie` header present
- JSON response with `data.user` and `data.session`

2) Validate authenticated session:

```bash
curl -k -i -b cookies.txt https://localhost:8080/api/v1/auth/session
```

Expected:
- HTTP 200
- `data.user.role` matches the logged-in demo user role

3) Verify role restrictions (example: Member against admin endpoint):

```bash
curl -k -i -c member.cookies \
	-H "Content-Type: application/json" \
	-d '{"email":"member.demo@fitcommerce.local","password":"Password123!"}' \
	https://localhost:8080/api/v1/auth/login

curl -k -i -b member.cookies https://localhost:8080/api/v1/admin/users
```

Expected:
- HTTP 403 for member on `/api/v1/admin/users`

### UI Verification

1) Open https://localhost:3443.
2) Login with `admin.demo@fitcommerce.local`.
3) Confirm admin-only sections are visible (Users, Audit, Backups, Biometrics).
4) Logout, then login with `member.demo@fitcommerce.local`.
5) Confirm member can access member-safe modules (catalog/group-buys/orders) and cannot access admin-only pages.

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
