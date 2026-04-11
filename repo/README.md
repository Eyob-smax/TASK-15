# FitCommerce Operations & Inventory Suite

A full-stack, offline-first, local-network web application for fitness club operations, inventory management, procurement, and member engagement. Designed to run entirely on-premises with no internet dependency.

## Stack

| Layer            | Technology                | Version                              |
| ---------------- | ------------------------- | ------------------------------------ |
| Frontend         | React + TypeScript + Vite | React 18.3, TypeScript 5.7, Vite 6.0 |
| UI Components    | MUI (Material UI)         | 6.3                                  |
| Server State     | TanStack React Query      | 5.62                                 |
| Routing          | React Router              | 7.1                                  |
| Form Validation  | React Hook Form + Zod     | 7.54 / 3.24                          |
| Backend          | Go + Echo v4              | Go 1.22, Echo 4.13                   |
| Database Driver  | pgx                       | v5.7                                 |
| Migrations       | goose                     | v3.24                                |
| Database         | PostgreSQL                | 15+ (Alpine)                         |
| Containerization | Docker + Docker Compose   | Compose file version 3.9             |

## Repository Structure

```
repo/
  docker-compose.yml              # Orchestrates all three services
  run_tests.sh                    # Unified test runner (backend + frontend)
  backend/
    Dockerfile                    # Multi-stage Go build (golang:1.22-alpine -> alpine:3.20)
    go.mod                        # Go module definition
    cmd/
      api/
        main.go                   # Application entry point
    database/
      migrations/
        00001_enums_and_base.sql          # Enum types, locations, retention_policies, encryption_keys
        00002_users_and_auth.sql          # users, sessions, captcha_challenges, biometric_enrollments
        00003_catalog_and_inventory.sql   # items, availability/blackout windows, warehouse_bins, inventory, batch_edits
        00004_members_and_coaches.sql     # members, coaches
        00005_campaigns_and_orders.sql    # group_buy_campaigns, orders, participants, timeline, fulfillment
        00006_procurement.sql             # suppliers, purchase_orders, PO lines, variances, landed_costs
        00007_audit_reports_backups.sql   # audit_events, report_definitions, export_jobs, backup_runs
        00008_variance_resolution_and_retention_cleanup.sql # variance resolution metadata, retention seed cleanup
      seeds/
        seed.sql                  # Reserved system actor, active retention policies, and report definitions
    internal/
      application/
        services.go               # 17 service interfaces (auth, items, orders, campaigns, locations, members, coaches, etc.)
      domain/
        enums.go                  # All enum types with validation
        errors.go                 # 10 domain error types
        state_machines.go         # Order, campaign, PO state machines
        policies.go               # Publish validation, window overlap detection
        user.go                   # User, Session, CaptchaChallenge
        item.go                   # Item, AvailabilityWindow, BlackoutWindow
        order.go                  # Order, OrderTimelineEntry, FulfillmentGroup
        campaign.go               # GroupBuyCampaign, GroupBuyParticipant
        purchase_order.go         # PurchaseOrder, PurchaseOrderLine
        inventory.go              # InventorySnapshot, InventoryAdjustment, WarehouseBin
        supplier.go               # Supplier
        variance.go               # VarianceRecord, LandedCostEntry, allocation logic
        audit.go                  # AuditEvent with SHA-256 hash chain
        backup.go                 # BackupRun
        biometric.go              # BiometricEnrollment, EncryptionKey
        retention.go              # RetentionPolicy, retention checking
        report.go                 # ReportDefinition, ExportJob, filename generation
        batch_edit.go             # BatchEditJob, BatchEditResult
        location.go               # Location
        member.go                 # Member, Coach
      http/
        router.go                 # 21 route groups, fully wired handlers (no 501 stubs)
        middleware.go             # Auth, RequireRole, RequestID, Recover
        errors.go                 # Error envelope and domain-to-HTTP error mapping
        dto/
          requests.go             # 16 request DTO types
          responses.go            # 30+ response DTO types
      store/
        repositories.go           # 22 repository interfaces
      security/
        security.go               # Package doc (Argon2id, CAPTCHA, AES-256, masking, RBAC)
      jobs/
        jobs.go                   # Package doc (auto-close, cutoff, backup, retention, variance)
      reporting/
        reporting.go              # Package doc (KPIs, CSV/PDF generation, export management)
      platform/
        config.go                 # FC_* environment variable configuration
        logger.go                 # Structured JSON logger (slog) + Echo middleware
    unit_tests/
      domain/                     # Backend domain unit tests (directory created)
    api_tests/                    # Backend API integration tests (directory created)
  frontend/
    Dockerfile                    # Multi-stage build (node:20-alpine -> nginx:1.27-alpine)
    package.json                  # Dependencies and scripts
    tsconfig.json                 # TypeScript configuration
    vite.config.ts                # Vite config with @ alias, proxy, and vitest setup
    index.html                    # SPA entry point
    src/
      main.tsx                    # React root render
      app/
        App.tsx                   # Root component (Providers + RouterProvider)
        providers.tsx             # QueryClient, ThemeProvider, AuthProvider
        routes.tsx                # All route definitions with lazy loading and role guards
        theme.ts                  # MUI theme (green/orange palette)
      features/
        admin/types.ts            # Admin feature types
        auth/types.ts             # Auth feature types
        catalog/types.ts          # Catalog feature types and filters
        dashboard/types.ts        # KPI types and filters
        group-buy/types.ts        # Group-buy feature types
        inventory/types.ts        # Inventory feature types
        orders/types.ts           # Order feature types
        procurement/types.ts      # Procurement feature types
        reports/types.ts          # Report feature types
      components/
        Layout.tsx                # App shell (NavSidebar + AppBar + Outlet + ErrorBoundary)
        NavSidebar.tsx            # Role-aware sidebar navigation (ROLE_PERMISSIONS filter)
        DataTable.tsx             # Server-paginated table (loading skeleton, empty state, error)
        FilterBar.tsx             # Reusable filter bar (text / select / date field types)
        PageContainer.tsx         # Page wrapper with title, breadcrumbs, and actions slot
        StatCard.tsx              # KPI stat card (value, label, change indicator, skeleton)
        ConfirmDialog.tsx         # Blocking confirmation dialog (destructive variant)
        EmptyState.tsx            # Illustrated empty state with optional action
        ErrorBoundary.tsx         # React class error boundary (reset on click)
      lib/
        api-client.ts             # Fetch-based HTTP client with error handling and session-expired dispatch
        auth.tsx                  # AuthProvider, useAuth, ProtectedRoute, RequireRole
        constants.ts              # Role permissions, enum labels, configuration constants
        format.ts                 # Currency, date, filename, masking formatters
        notifications.tsx         # NotificationsProvider, useNotify (toast queue via MUI Snackbar)
        types.ts                  # All shared TypeScript interfaces
        validation.ts             # Zod schemas for all forms
        hooks/
          useDashboard.ts         # TanStack Query hooks for dashboard KPIs
          useItems.ts             # Item list, get, publish, unpublish, update, batch-edit
          useOrders.ts            # Order list, get, timeline, cancel, pay, refund, add-note
          useCampaigns.ts         # Campaign list, get, join, cancel, evaluate
          useInventory.ts         # Inventory snapshots, adjustments, warehouse bins
      routes/
        LoginPage.tsx             # Public login form
        DashboardPage.tsx         # KPI dashboard
        CatalogPage.tsx           # Item list
        CatalogDetailPage.tsx     # Item detail view
        CatalogFormPage.tsx       # Item create/edit form
        InventoryPage.tsx         # Inventory snapshots and adjustments
        GroupBuysPage.tsx         # Campaign list
        GroupBuyDetailPage.tsx    # Campaign detail with join
        OrdersPage.tsx            # Order list
        OrderDetailPage.tsx       # Order detail with actions
        ProcurementPage.tsx       # Procurement hub
        SuppliersPage.tsx         # Supplier management
        PurchaseOrdersPage.tsx    # PO list
        PurchaseOrderDetailPage.tsx # PO detail with lifecycle actions
        VariancesPage.tsx         # Variance list with overdue indicator and resolve dialog
        LandedCostsPage.tsx       # Landed cost search by item/period
        LocationsPage.tsx         # Location list
        MembersPage.tsx           # Member list
        CoachesPage.tsx           # Coach list
        ReportsPage.tsx           # Report list and export initiation
        AdminPage.tsx             # Admin hub
        UsersPage.tsx             # User management
        AuditPage.tsx             # Audit log viewer
        BackupsPage.tsx           # Backup history and manual trigger
        BiometricPage.tsx         # Biometric key management
    unit_tests/
      setup.ts                    # Vitest setup (imports @testing-library/jest-dom)
      auth/                       # Auth component tests (ProtectedRoute, RequireRole, session-timeout)
      lib/                        # Frontend lib unit tests (constants, validation)
      components/                 # Shared UI component tests (DataTable, FilterBar, Layout/NavSidebar)
      routes/                     # Page-level tests (LoginPage, DashboardPage)
```

## Prerequisites

- **Docker** (version 20.10+)
- **Docker Compose** (version 2.0+ or the `docker compose` plugin)

No other tools are required for running the application. For development without Docker, you would need Go 1.22+, Node.js 20+, and PostgreSQL 15+.

## Getting Started

From the `repo/` directory:

```bash
docker compose up --build
```

This starts three services:

1. **PostgreSQL** -- initializes the database, runs goose migrations on first startup
2. **Backend** -- Go API server (waits for PostgreSQL health check)
3. **Frontend** -- nginx serving the built React SPA

## Services and Ports

| Service                    | Container Name  | URL                               | Port |
| -------------------------- | --------------- | --------------------------------- | ---- |
| Frontend (Docker redirect) | fitcommerce-ui  | http://localhost:3000 (redirects) | 3000 |
| Frontend (Docker HTTPS)    | fitcommerce-ui  | https://localhost:3443            | 3443 |
| Frontend (Vite dev)        | --              | http://localhost:5173             | 5173 |
| Backend API                | fitcommerce-api | https://localhost:8080            | 8080 |
| PostgreSQL                 | fitcommerce-db  | localhost:5432                    | 5432 |

The frontend nginx configuration serves HTTPS on port 3443 and redirects HTTP traffic from port 3000 to HTTPS. It proxies all `/api/*` requests to the backend container over HTTPS and forwards `X-Forwarded-Proto: https` so backend session cookies remain secure. In development mode (Vite), the proxy defaults to `https://localhost:8080` and skips certificate verification so the backend can use its generated self-signed certificate; set `VITE_DEV_API_TARGET` only when you intentionally run the backend over a different local endpoint.

## Environment Variables

All backend configuration uses the `FC_` prefix. These are set in `docker-compose.yml` and can be overridden:

| Variable                            | Default                                                                       | Description                                                                                                                                      |
| ----------------------------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `FC_SERVER_PORT`                    | `8080`                                                                        | Backend API port (TLS by default)                                                                                                                |
| `FC_DATABASE_URL`                   | `postgres://fitcommerce:fitcommerce@postgres:5432/fitcommerce?sslmode=prefer` | PostgreSQL connection string                                                                                                                     |
| `FC_RUN_MIGRATIONS_ON_STARTUP`      | `true`                                                                        | Run goose migrations during backend startup                                                                                                      |
| `FC_RUN_SEED_ON_STARTUP`            | `true`                                                                        | Run idempotent SQL seed file during backend startup                                                                                              |
| `FC_LOG_LEVEL`                      | `info`                                                                        | Log level: debug, info, warn, error                                                                                                              |
| `FC_CLUB_TIMEZONE`                  | `America/New_York`                                                            | Default timezone for the club                                                                                                                    |
| `FC_BACKUP_PATH`                    | `/var/backups/fitcommerce`                                                    | Filesystem path for backup archives                                                                                                              |
| `FC_SESSION_IDLE_TIMEOUT_MINUTES`   | `30`                                                                          | Session idle timeout in minutes                                                                                                                  |
| `FC_SESSION_ABSOLUTE_TIMEOUT_HOURS` | `12`                                                                          | Session absolute timeout in hours                                                                                                                |
| `FC_LOGIN_LOCKOUT_THRESHOLD`        | `5`                                                                           | Failed login attempts before lockout                                                                                                             |
| `FC_LOGIN_LOCKOUT_DURATION_MINUTES` | `15`                                                                          | Lockout duration in minutes                                                                                                                      |
| `FC_BIOMETRIC_MODULE_ENABLED`       | `false`                                                                       | Enable/disable biometric module                                                                                                                  |
| `FC_BIOMETRIC_KEY_ROTATION_DAYS`    | `90`                                                                          | Days before encryption key rotation                                                                                                              |
| `FC_ALLOW_INSECURE_HTTP`            | `false`                                                                       | Allow plain HTTP without TLS. Set `true` for local development only.                                                                             |
| `FC_TLS_CERT_FILE`                  | _(empty)_                                                                     | Optional path to a TLS certificate. When omitted and secure transport is required, the API generates a local self-signed certificate at startup. |
| `FC_TLS_KEY_FILE`                   | _(empty)_                                                                     | Optional path to the matching TLS private key. Must be provided together with `FC_TLS_CERT_FILE` when you want a custom certificate.             |
| `FC_BACKUP_ENCRYPTION_KEY_REF`      | _(empty)_                                                                     | Reference for backup encryption key (required for successful backup trigger)                                                                     |
| `FC_EXPORT_PATH`                    | `/tmp/fitcommerce-exports` (local dev)                                        | Filesystem path for report export files. Overridden to `/var/exports/fitcommerce` (named Docker volume) in `docker-compose.yml`.                 |

PostgreSQL credentials (set on the `postgres` service):

| Variable            | Default       |
| ------------------- | ------------- |
| `POSTGRES_USER`     | `fitcommerce` |
| `POSTGRES_PASSWORD` | `fitcommerce` |
| `POSTGRES_DB`       | `fitcommerce` |

## Database Migrations

Ten goose SQL migrations are located in `backend/database/migrations/`:

| Migration                                             | Description                                                                  |
| ----------------------------------------------------- | ---------------------------------------------------------------------------- |
| `00001_enums_and_base.sql`                            | Enum types (16 enums), locations, retention_policies, encryption_keys        |
| `00002_users_and_auth.sql`                            | users, sessions, captcha_challenges, biometric_enrollments                   |
| `00003_catalog_and_inventory.sql`                     | items, availability/blackout windows, warehouse_bins, inventory, batch edits |
| `00004_members_and_coaches.sql`                       | members, coaches                                                             |
| `00005_campaigns_and_orders.sql`                      | group_buy_campaigns, orders, participants, timeline, fulfillment groups      |
| `00006_procurement.sql`                               | suppliers, purchase_orders, PO lines, variance_records, landed_cost_entries  |
| `00007_audit_reports_backups.sql`                     | audit_events, report_definitions, export_jobs, backup_runs                   |
| `00008_variance_resolution_and_retention_cleanup.sql` | variance resolution metadata and retention-policy cleanup                    |
| `00009_retention_days_and_update_tracking.sql`        | day-granularity retention policies and `updated_at` tracking                 |
| `00010_captcha_answer_hashing.sql`                    | salted CAPTCHA verification hashes (replaces plaintext answers)              |

Seed data (`backend/database/seeds/seed.sql`) inserts a reserved system actor user, active retention policies for operational data, and report definitions. `audit_events` is intentionally excluded from retention purge targets.

Migrations are designed to run automatically on backend startup via goose. Each migration includes both `+goose Up` and `+goose Down` blocks for reversibility.

## Running Tests

Tests run entirely inside Docker. No host-installed Go, Node, or Vitest is required.

From the `repo/` directory:

```bash
# Run all suites (backend unit + API tests, frontend unit tests)
./run_tests.sh

# Run all suites with coverage output
./run_tests.sh --coverage
```

The script uses `docker compose --profile test run` to execute each suite in an isolated container. It fails fast — if the backend suite fails, the frontend suite does not run. Coverage reports (when `--coverage` is passed) are written to `backend/coverage_unit.out`, `backend/coverage_api.out`, and `frontend/coverage/`.

The test containers (`backend-test`, `frontend-test`) are declared with `profiles: ["test"]` in `docker-compose.yml` and are never started by `docker compose up`.

### Test Locations

| Suite               | Directory              | Runner                         | Coverage                    |
| ------------------- | ---------------------- | ------------------------------ | --------------------------- |
| Frontend unit tests | `frontend/unit_tests/` | Vitest (`@vitest/coverage-v8`) | `frontend/coverage/`        |
| Backend unit tests  | `backend/unit_tests/`  | `go test`                      | `backend/coverage_unit.out` |
| Backend API tests   | `backend/api_tests/`   | `go test`                      | `backend/coverage_api.out`  |

### Test Scope

Backend unit tests cover:

- Domain invariants, state machines, enums (14 domain test files)
- Security: password hashing, session management, CAPTCHA, RBAC, AES-256/HKDF crypto, masking, audit chain
- Services: items, orders, campaigns, procurement, variance, backup, user, reports, jobs
- Platform: config loading, env overrides, fallback behavior

Backend API tests cover:

- Auth flows (login, logout, session, lockout, captcha)
- Authorization failures (401/403 per role per route)
- CRUD and transition flows: items, campaigns, orders, procurement, admin, inventory
- Error envelope format, DTO validation, not-found paths

Frontend unit tests cover:

- Auth guards: ProtectedRoute, RequireRole, session-timeout redirect
- Shared components: DataTable, FilterBar, Layout/NavSidebar
- Library: Zod schemas (validation), constants, format helpers
- Route pages: Login, Dashboard (period/KPI), Catalog, CatalogForm (item validation), GroupBuyDetail, OrderDetail, PurchaseOrders, Variances, Suppliers, Inventory, Reports, Users, Audit, Admin

## Offline / Local-Network Constraints

This application is designed for air-gapped or LAN-only deployment:

- No external API calls are made at runtime by the frontend or backend.
- The frontend fetches all assets from the nginx container (no CDN).
- The API base URL defaults to `/api/v1` (relative), so the frontend works from any LAN IP or localhost.
- TanStack Query is configured with `refetchOnWindowFocus: false`, persisted IndexedDB cache hydration, and offline-aware error handling so cached reads remain available when the local backend is temporarily unreachable.
- Offline mode is intentionally read-only in this rollout: cached dashboard/catalog/campaign/order/inventory/procurement/report metadata remains visible, while writes, exports/downloads, login/logout/session refresh, and CAPTCHA verification require reconnecting to the backend.
- Container images are built once; after that, `docker compose up` requires no network access.

## Security Notes

- **Passwords**: Hashed with Argon2id and per-user salts. Never returned in API responses.
- **Sessions**: Server-side, stored in PostgreSQL. 30-minute idle timeout, 12-hour absolute timeout.
- **Login lockout**: After 5 failed attempts, account is locked for 15 minutes with CAPTCHA.
- **CAPTCHA storage**: CAPTCHA answers are never stored in plaintext; the server stores a salted SHA-256 verification hash and compares submissions in constant time.
- **RBAC**: Three enforcement layers -- route middleware, service checks, and data-scope filtering.
- **Audit log**: Tamper-evident with SHA-256 hash chaining across all events.
- **SQL injection**: Prevented via pgx parameterized queries throughout.
- **XSS**: Prevented via React JSX auto-escaping and JSON-only API responses.

## Backup Configuration

Set `FC_BACKUP_PATH` to the desired filesystem path for backup archives. In Docker, this is mounted as a named volume (`backups`). Backups are:

- AES-256 encrypted (mandatory; backup trigger fails when `FC_BACKUP_ENCRYPTION_KEY_REF` is unset; key derived via HKDF-SHA256)
- SHA-256 checksummed after encryption so the stored checksum matches the final retained archive
- Metadata tracked in the `backup_runs` database table
- Viewable and manually triggerable from the Admin panel

Restore is an operational procedure: decrypt the archive, verify the checksum, and run `pg_restore`.

**pg_dump integration**: The `BackupService` uses an injected `DumpFunc`. `cmd/api/main.go` shells out to `pg_dump` via `exec.CommandContext`. The **supported and primary deployment path is containerized**: the backend Docker image installs `postgresql-client` (via `apk add --no-cache postgresql-client` in the runtime stage), which provides the `pg_dump` binary. The entire backup pipeline — pg_dump invocation, checksum, encryption, audit event — is fully operational when running under Docker Compose. For local development outside Docker, install a PostgreSQL client (`brew install libpq`, `apt install postgresql-client`, etc.) to make `pg_dump` available in PATH; without it, backup operations will fail at trigger time but do not affect any other functionality.

## Backend Capabilities

The Go backend implements the following core capability domains:

- **Catalog**: item CRUD (with SKU and unit price) with optimistic concurrency, publish/unpublish workflow, batch editing, availability/blackout windows
- **Inventory**: snapshots, manual adjustments, warehouse bin management, and transactional quantity synchronization across items, adjustments, and snapshots
- **Group-Buy Campaigns**: campaign lifecycle (Active→Succeeded/Failed/Cancelled), participant join, cutoff evaluation, inventory reservation with stock and availability-window enforcement, and item-linked detail views
- **Orders**: full state machine (Created→Paid→Cancelled/Refunded/AutoClosed), split/merge, timeline, auto-close background job, and transactional stock release on cancel/refund/auto-close
- **Procurement**: supplier management, purchase order lifecycle (Created→Approved→Received→Returned/Voided), quantity and price variance detection, explicit variance resolution actions (`adjustment` or `return`), landed-cost entry creation per PO receipt, and landed-cost query by item/period or by PO
- **Locations**: gym/branch location management (create, list, get by ID)
- **Members**: member enrollment with status tracking (create, list, get by ID; filterable by location). Directory access is limited to administrators and operations managers
- **Coaches**: coach assignment with specialization (create, list, get by ID; filterable by location)
- **Dashboard KPIs**: six real-time metrics (member growth, churn rate, renewal rate, engagement, class fill rate, coach productivity); role-protected via `ActionViewDashboard`
- **Reports & Exports**: predefined report definitions with role-aware access control, CSV and PDF export generation (`gofpdf`), export job lifecycle tracking, whitelisted export filter parameters (`location_id`, `coach_id`, `category`, `from`, `to`, `status`), and a file download endpoint
- **Admin / Audit**: append-only SHA-256 hash-chained audit event log with system actor support, security-event inspection endpoint (login failures, lockouts, session expiry, CAPTCHA results), user management with role-based data masking
- **Backups**: manual and scheduled (every 24 h) backup runs with real `pg_dump` invocation, AES-256-GCM encryption, SHA-256 checksum verification against the final encrypted archive, and key derivation via HKDF-SHA256
- **Retention**: configurable retention policies per entity type; cleanup job runs every 24 h, records per-record deletion audit events for purgeable entities, and never purges append-only `audit_events`
- **Biometric Controls**: enrollment administration, template redaction in responses, revocation, and 90-day key rotation with re-encryption of existing active enrollments; the module can be globally disabled via `FC_BIOMETRIC_MODULE_ENABLED`

## Background Jobs

Five background jobs start automatically on API server startup:

| Job                   | Interval | Behaviour                                                                         |
| --------------------- | -------- | --------------------------------------------------------------------------------- |
| `AutoCloseJob`        | 1 min    | Auto-closes orders past their close date                                          |
| `CutoffEvalJob`       | 1 min    | Evaluates campaign cutoffs, marks Succeeded/Failed                                |
| `BackupJob`           | 24 h     | Calls `BackupService.Trigger(ctx, nil)` (system actor)                            |
| `VarianceDeadlineJob` | 1 h      | Escalates overdue open variances and records audit events                         |
| `RetentionCleanupJob` | 24 h     | Calls `RetentionService.RunCleanup` to audit and purge eligible non-audit records |

## Development Status

All application service implementations, HTTP handlers, background jobs, frontend pages, and tests are authored. Key status:

- **Docker configuration**: `docker-compose.yml`, backend `Dockerfile`, and frontend `Dockerfile` are authored. The frontend nginx configuration is in `frontend/nginx.conf` (COPY-based, no heredoc). None have been executed in a live Docker environment.
- **Database migrations**: All 10 goose migrations and seed data are written. They have not been run against a live PostgreSQL instance.
- **Backend**: Go application is fully implemented — domain layer, store interfaces, Postgres store implementations, application services, HTTP handlers, RBAC middleware, and background jobs. All endpoints are implemented; no 501 stubs remain in the production route set.
- **Frontend**: React application is fully implemented — routing, providers, TanStack Query hooks, and page components with API integration aligned to the shipped backend contracts. This includes report exports with filter forwarding, item-backed group-buy detail pages, procurement receive/resolve dialogs, and route guards that keep member-directory access limited to administrators and operations managers.
- **Tests**: Backend unit tests cover domain invariants, state machines, service business rules (procurement, variance, backup, report masking, user service, config loading), and crypto primitives. Backend API tests cover error envelopes, DTO validation, auth flows, procurement, admin, and inventory endpoints. Frontend tests cover 14 route pages including item-form validation, inventory tabs, variance resolution, supplier management, reports export, and admin surfaces.
- **Request validation**: All mutation endpoints (create item, update item, create PO, receive PO, create supplier, update supplier, create order) run `c.Validate(&req)` after binding, enforcing struct tags via go-playground/validator.
- **System actor**: A reserved non-human user (`system@fitcommerce.internal`, UUID `00000000-0000-0000-0000-000000000001`, status `inactive`) is seeded for all system-initiated audit events. The auth login flow explicitly blocks inactive accounts, so this actor cannot be used to obtain a session.
