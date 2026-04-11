# FitCommerce Operations & Inventory Suite -- Architecture & Design Document

## 1. System Overview

FitCommerce is an **offline-first, local-network** web application designed for fitness club operations. It runs entirely on-premises with no internet dependency. All three services (frontend, backend, database) are deployed via Docker Compose on a single host or LAN segment.

- **Frontend** -- React SPA served over HTTP. Members and staff access it via a browser pointed at the host's LAN IP or `localhost`.
- **Backend** -- Go API server. The single backend process handles all HTTP requests, runs background jobs, and manages the backup subsystem.
- **Database** -- PostgreSQL 15+. System of record for all persistent state.

No cloud services, CDNs, or external package registries are contacted at runtime. Container images are built once; after that, the system operates in a fully air-gapped environment.

## 2. Architecture Diagram

```
 +---------------------------------------------------------+
 |                      LAN / localhost                     |
 +---------------------------------------------------------+
         |                        |
   +-----v------+          +-----v------+
   |  Browser    |          |  Browser    |
   |  (Staff)    |          |  (Member)   |
   +-----+------+          +-----+------+
         |                        |
         +----------+-------------+
                    |
              +-----v------+
              |   nginx     |  :3000  (Docker)
              |  (frontend) |  :5173  (Vite dev)
              +-----+------+
                    |
                    | /api/* proxy_pass
                    |
              +-----v------+
              |   Go API    |  :8080
              |  (backend)  |
              +--+-----+---+
                 |     |
    +------------+     +-------------+
    |                                |
    v                                v
 +-------+                  +----------------+
 | pgx   |                  | Background     |
 | pool   |                 | Jobs           |
 +---+---+                  | - auto-close   |
     |                      | - cutoff eval  |
     v                      | - variance     |
 +--------+                 | - backup       |
 |Postgres|  :5432          | - retention    |
 |  15    |                 | - key rotation |
 +--------+                 +-------+--------+
     ^                              |
     +------------------------------+
         (jobs also use pgx pool)
```

### Data Flow

1. Browser sends HTTP requests to nginx on port 3000.
2. nginx serves static assets directly; routes matching `/api/*` are reverse-proxied to the Go backend on port 8080.
3. The backend processes the request through middleware (request-ID, recovery, auth, RBAC), delegates to the application service layer, which calls domain logic and persists via the store layer to PostgreSQL.
4. Background jobs run as goroutines within the Go process, sharing the pgx connection pool.
5. Nightly backups write encrypted archives to the admin-configured filesystem path.

## 3. Frontend Architecture

### Technology Stack

| Concern | Library | Version |
|---|---|---|
| UI Framework | React | 18.3 |
| Language | TypeScript | 5.7 |
| Build Tool | Vite | 6.0 |
| Component Library | MUI (Material UI) | 6.3 |
| Server State | TanStack React Query | 5.62 |
| Routing | React Router | 7.1 |
| Form Validation | React Hook Form + Zod | 7.54 / 3.24 |
| Date Handling | Day.js | 1.11 |
| Testing | Vitest + Testing Library | 2.1 / 16.1 |

### TanStack Query Configuration

Configured in `src/app/providers.tsx`:

- `staleTime`: 30 seconds
- `refetchOnWindowFocus`: false (critical for offline/LAN use)
- `retry`: 1 for queries, 0 for mutations
- `refetchOnReconnect`: false
- IndexedDB-backed query persistence hydrates cached reads for dashboard, catalog, campaigns, orders, inventory, procurement, and report metadata
- Offline mode is intentionally read-only in this rollout; writes and exports require reconnecting to the backend

### Routing

All route components are **lazy-loaded** via `React.lazy()` with a `Suspense` fallback showing a `CircularProgress` spinner. Routes are defined in `src/app/routes.tsx` using `createBrowserRouter`.

Protected routes use `ProtectedRoute` from `src/lib/auth.tsx`, which checks authentication status and optionally enforces an `allowedRoles` array.

### Folder Structure

```
frontend/src/
  app/
    App.tsx             -- Root component, wires Providers + RouterProvider
    providers.tsx       -- QueryClient, ThemeProvider, AuthProvider
    routes.tsx          -- All route definitions with lazy loading + role guards
    theme.ts            -- MUI theme (green primary, orange secondary)
  features/
    admin/types.ts      -- Admin-specific types
    auth/types.ts       -- LoginFormData, SessionInfo, LoginResponse
    catalog/types.ts    -- ItemFilters, ItemFormData, BatchEditFormData
    dashboard/types.ts  -- KPIFilters, KPICardData, DashboardSummary
    group-buy/types.ts  -- Campaign feature types
    inventory/types.ts  -- Inventory feature types
    orders/types.ts     -- Order feature types
    procurement/types.ts -- Supplier/PO feature types
    reports/types.ts    -- Report/export feature types
  lib/
    api-client.ts       -- Fetch-based HTTP client with error handling
    auth.tsx            -- AuthProvider, useAuth, ProtectedRoute, RequireRole
    constants.ts        -- Role permissions, enum labels, config constants
    format.ts           -- Currency, date, filename, masking formatters
    types.ts            -- All shared TypeScript interfaces and type aliases
    validation.ts       -- Zod schemas for every form in the application
  routes/
    LoginPage.tsx       -- Login form (public)
    DashboardPage.tsx   -- KPI dashboard (all authenticated)
    CatalogPage.tsx     -- Item list (admin, ops_mgr)
    CatalogDetailPage.tsx
    CatalogFormPage.tsx -- Create/edit item
    InventoryPage.tsx   -- Snapshots and adjustments (admin, ops_mgr)
    GroupBuysPage.tsx   -- Campaign list (all authenticated)
    GroupBuyDetailPage.tsx
    OrdersPage.tsx      -- Order list (all authenticated)
    OrderDetailPage.tsx
    ProcurementPage.tsx -- Procurement hub (admin, ops_mgr, procurement)
    SuppliersPage.tsx
    PurchaseOrdersPage.tsx
    PurchaseOrderDetailPage.tsx
    LandedCostsPage.tsx
    ReportsPage.tsx     -- Reports list (all authenticated)
    AdminPage.tsx       -- Admin hub (administrator only)
    UsersPage.tsx
    AuditPage.tsx
    BackupsPage.tsx
    BiometricPage.tsx
```

### Role-Aware Navigation

The `ROLE_PERMISSIONS` map in `constants.ts` defines which modules each role can access:

| Role | Modules |
|---|---|
| Administrator | dashboard, catalog, inventory, group_buys, orders, procurement, reports, admin, audit, backups, biometric, users |
| Operations Manager | dashboard, catalog, inventory, group_buys, orders, procurement, reports |
| Procurement Specialist | catalog, procurement, reports |
| Coach | dashboard, group_buys, orders, reports |
| Member | catalog, group_buys, orders |

## 4. Backend Architecture

### Technology Stack

| Concern | Library | Version |
|---|---|---|
| Language | Go | 1.22 |
| HTTP Framework | Echo | v4.13 |
| Database Driver | pgx | v5.7 |
| Migrations | goose | v3.24 |
| UUID | google/uuid | 1.6 |
| Crypto | golang.org/x/crypto | 0.31 |
| Logging | log/slog (stdlib) | -- |

### Layered Architecture

```
cmd/api/main.go
  |
  +-- internal/http/          -- HTTP layer (routes, middleware, DTOs, error handling)
  |     router.go             -- Route registration (18 route groups)
  |     middleware.go          -- AuthMiddleware, RequireRole, RequestID, Recover
  |     errors.go             -- ErrorResponse envelope, HandleDomainError mapper
  |     dto/requests.go       -- Request DTOs (16 request types)
  |     dto/responses.go      -- Response DTOs (30+ response types)
  |
  +-- internal/application/   -- Application service interfaces
  |     services.go           -- 14 service interfaces defining all use cases
  |
  +-- internal/domain/        -- Pure domain logic (no external dependencies)
  |     enums.go              -- All enum types with validation helpers
  |     errors.go             -- Domain error types (10 error types)
  |     state_machines.go     -- Order, Campaign, PO state machines
  |     policies.go           -- Publish validation, window overlap detection
  |     user.go               -- User, Session, CaptchaChallenge
  |     item.go               -- Item, AvailabilityWindow, BlackoutWindow
  |     order.go              -- Order, OrderTimelineEntry, FulfillmentGroup
  |     campaign.go           -- GroupBuyCampaign, GroupBuyParticipant
  |     purchase_order.go     -- PurchaseOrder, PurchaseOrderLine
  |     inventory.go          -- InventorySnapshot, InventoryAdjustment, WarehouseBin
  |     supplier.go           -- Supplier
  |     variance.go           -- VarianceRecord, LandedCostEntry, allocation logic
  |     audit.go              -- AuditEvent with SHA-256 hash chaining
  |     backup.go             -- BackupRun
  |     biometric.go          -- BiometricEnrollment, EncryptionKey
  |     retention.go          -- RetentionPolicy, retention check logic
  |     report.go             -- ReportDefinition, ExportJob, filename generation
  |     batch_edit.go         -- BatchEditJob, BatchEditResult
  |     location.go           -- Location
  |     member.go             -- Member, Coach
  |
  +-- internal/store/         -- Repository interfaces (PostgreSQL persistence)
  |     repositories.go       -- 22 repository interfaces
  |
  +-- internal/security/      -- Cryptographic and access-control utilities
  |     security.go           -- Package doc (Argon2id, CAPTCHA, AES-256, masking, RBAC)
  |
  +-- internal/jobs/          -- Background job runners
  |     jobs.go               -- Package doc (auto-close, cutoff, backup, retention, variance)
  |
  +-- internal/reporting/     -- KPI aggregation and export generation
  |     reporting.go          -- Package doc (KPIs, CSV, PDF, export management)
  |
  +-- internal/platform/      -- Cross-cutting infrastructure
        config.go             -- Environment-based configuration (FC_ prefix)
        logger.go             -- Structured JSON logger via slog
```

### Startup Sequence

1. `main.go` loads configuration from `FC_*` environment variables.
2. Creates a structured JSON logger.
3. Establishes a pgx connection pool (25 max, 5 min connections).
4. Creates an Echo instance with request-ID and recovery middleware.
5. Registers all route groups via `RegisterRoutes`.
6. Starts the HTTP server (with optional TLS).
7. Listens for SIGINT/SIGTERM for graceful shutdown (15s timeout).

### Route Groups

The backend registers 18 route groups under `/api/v1/`:

| Group | Prefix | Endpoint Count |
|---|---|---|
| Auth | `/auth` | 4 |
| Dashboard | `/dashboard` | 1 |
| Items | `/items` | 7 |
| Inventory | `/inventory` | 3 |
| Warehouse Bins | `/warehouse-bins` | 3 |
| Campaigns | `/campaigns` | 6 |
| Orders | `/orders` | 10 |
| Suppliers | `/suppliers` | 4 |
| Purchase Orders | `/purchase-orders` | 7 |
| Variances | `/variances` | 3 |
| Procurement | `/procurement` | 2 |
| Reports | `/reports` | 2 |
| Exports | `/exports` | 3 |
| Admin | `/admin` | 10 |
| Locations | `/locations` | 3 |
| Coaches | `/coaches` | 3 |
| Members | `/members` | 3 |

## 5. Domain Model

### Entity Relationship Summary

```
Users ─────┬── Sessions
           ├── CaptchaChallenges
           ├── BiometricEnrollments ──── EncryptionKeys
           ├── Members ──── Locations
           ├── Coaches ──── Locations
           ├── Items ───┬── AvailabilityWindows
           |            ├── BlackoutWindows
           |            ├── InventorySnapshots
           |            ├── InventoryAdjustments
           |            └── BatchEditJobs ──── BatchEditResults
           ├── Orders ──┬── OrderTimelineEntries
           |            ├── FulfillmentGroups ──── FulfillmentGroupOrders
           |            └── GroupBuyParticipants
           ├── GroupBuyCampaigns ──── GroupBuyParticipants
           ├── Suppliers ──── PurchaseOrders ──┬── PurchaseOrderLines
           |                                   ├── VarianceRecords
           |                                   └── LandedCostEntries
           └── AuditEvents
               BackupRuns
               RetentionPolicies
               ReportDefinitions ──── ExportJobs
```

### Entity Details

| Entity | Key Fields | Table |
|---|---|---|
| **User** | id, email, password_hash, salt, role, status, display_name, location_id, failed_login_count, locked_until | `users` |
| **Session** | id, user_id, token, idle_expires_at, absolute_expires_at | `sessions` |
| **CaptchaChallenge** | id, user_id, challenge_data, answer_hash, answer_salt, verified, expires_at | `captcha_challenges` |
| **Location** | id, name, address, timezone, is_active | `locations` |
| **Member** | id, user_id, location_id, membership_status, joined_at, renewal_date | `members` |
| **Coach** | id, user_id, location_id, specialization, is_active | `coaches` |
| **Item** | id, name, description, category, brand, sku, condition, billing_model, unit_price, refundable_deposit, quantity, status, location_id, created_by, version | `items` |
| **AvailabilityWindow** | id, item_id, start_time, end_time | `item_availability_windows` |
| **BlackoutWindow** | id, item_id, start_time, end_time | `item_blackout_windows` |
| **WarehouseBin** | id, location_id, name, description | `warehouse_bins` |
| **InventorySnapshot** | id, item_id, quantity, location_id, recorded_at | `inventory_snapshots` |
| **InventoryAdjustment** | id, item_id, quantity_change, reason, created_by | `inventory_adjustments` |
| **BatchEditJob** | id, created_by, total_rows, success_count, failure_count | `batch_edit_jobs` |
| **BatchEditResult** | id, batch_id, item_id, field, old_value, new_value, success, failure_reason | `batch_edit_results` |
| **GroupBuyCampaign** | id, item_id, min_quantity, current_committed_qty, cutoff_time, status, created_by, evaluated_at | `group_buy_campaigns` |
| **GroupBuyParticipant** | id, campaign_id, user_id, quantity, order_id, joined_at | `group_buy_participants` |
| **Order** | id, user_id, item_id, campaign_id, quantity, unit_price, total_amount, status, settlement_marker, notes, auto_close_at, paid_at, cancelled_at, refunded_at | `orders` |
| **OrderTimelineEntry** | id, order_id, action, description, performed_by | `order_timeline_entries` |
| **FulfillmentGroup** | id, supplier_id, warehouse_bin_id, pickup_point, status | `fulfillment_groups` |
| **FulfillmentGroupOrder** | id, fulfillment_group_id, order_id, quantity | `fulfillment_group_orders` |
| **Supplier** | id, name, contact_name, contact_email, contact_phone, address, is_active | `suppliers` |
| **PurchaseOrder** | id, supplier_id, status, total_amount, created_by, approved_by, approved_at, received_at, version | `purchase_orders` |
| **PurchaseOrderLine** | id, purchase_order_id, item_id, ordered_quantity, ordered_unit_price, received_quantity, received_unit_price | `purchase_order_lines` |
| **VarianceRecord** | id, po_line_id, type, expected_value, actual_value, difference_amount, status, resolution_due_date, resolved_at, resolution_notes | `variance_records` |
| **LandedCostEntry** | id, item_id, purchase_order_id, po_line_id, period, cost_component, raw_amount, allocated_amount, allocation_method | `landed_cost_entries` |
| **AuditEvent** | id, event_type, entity_type, entity_id, actor_id, details (JSONB), integrity_hash, previous_hash | `audit_events` |
| **BackupRun** | id, archive_path, checksum, checksum_algorithm, encryption_key_ref, status, file_size, started_at, completed_at | `backup_runs` |
| **RetentionPolicy** | id, entity_type, retention_days, description, updated_at | `retention_policies` |
| **ReportDefinition** | id, name, report_type, description, allowed_roles (user_role[]), filters (JSONB) | `report_definitions` |
| **ExportJob** | id, report_id, format, filename, status, file_path, created_by, completed_at | `export_jobs` |
| **BiometricEnrollment** | id, user_id, encrypted_data (BYTEA), encryption_key_id | `biometric_enrollments` |
| **EncryptionKey** | id, key_reference, purpose, status, activated_at, rotated_at, expires_at | `encryption_keys` |

## 6. State Machines

### Order State Machine

```
                          +----------+
                          |          |
               +--------->| paid     +-------+-------+
               |          |          |       |       |
               |          +----------+       |       |
               |                             v       v
          +----+-----+              +--------+-+ +---+------+
          |          |              |          | |          |
  Start-->| created  +------------->| cancelled| | refunded |
          |          |              |          | |          |
          +----+-----+              +----------+ +----------+
               |
               |
               v
          +----+-------+
          |            |
          | auto_closed|
          |            |
          +------------+
```

Valid transitions defined in `domain/state_machines.go`:

- `created` -> `paid`, `cancelled`, `auto_closed`
- `paid` -> `cancelled`, `refunded`
- All of `paid`, `cancelled`, `refunded`, `auto_closed` are terminal states (except `paid` allows further transitions).

### Campaign State Machine

```
          +----------+
          |          |
  Start-->|  active  +------+-------+-------+
          |          |      |       |       |
          +----------+      v       v       v
                    +-------+-+ +---+----+ +---+------+
                    |         | |        | |          |
                    |succeeded| | failed | | cancelled|
                    |         | |        | |          |
                    +---------+ +--------+ +----------+
```

Valid transitions:

- `active` -> `succeeded`, `failed`, `cancelled`
- All terminal states have no outbound transitions.
- Evaluation logic in `GroupBuyCampaign.Evaluate()`: at cutoff time, if `current_committed_qty >= min_quantity` then `succeeded`, else `failed`.

### Purchase Order State Machine

```
          +---------+        +----------+        +----------+
          |         |        |          |        |          |
  Start-->| created +------->| approved +------->| received |
          |         |        |          |        |          |
          +---------+        +----+-----+        +---+--+---+
                                  |                  |  |
                                  v                  |  v
                             +----+---+              | +---+-----+
                             |        |              | |         |
                             | voided |<-------------+ | returned|
                             |        |                |         |
                             +--------+                +---------+
```

Valid transitions:

- `created` -> `approved`
- `approved` -> `received`, `voided`
- `received` -> `returned`, `voided`

## 7. Security Model

### Password Hashing

- **Algorithm**: Argon2id (via `golang.org/x/crypto`)
- **Per-user salts** stored alongside password hashes in the `users` table
- Password hash and salt columns are `TEXT` type, never returned in API responses

### Login Lockout & CAPTCHA

- After **5 consecutive failed logins** (`FC_LOGIN_LOCKOUT_THRESHOLD`), the account is locked for **15 minutes** (`FC_LOGIN_LOCKOUT_DURATION_MINUTES`)
- During lockout, a server-side CAPTCHA challenge is generated and stored in `captcha_challenges` with salted one-way verification material rather than a plaintext answer
- The `ErrAccountLocked` and `ErrCaptchaRequired` domain errors map to `ACCOUNT_LOCKED` and `CAPTCHA_REQUIRED` HTTP error codes
- `User.IncrementFailedLogin()` and `User.Lock(duration)` implement the lockout logic
- `User.ResetFailedLogin()` clears the counter on successful login

### Session Management

- Server-side sessions stored in the `sessions` PostgreSQL table
- **Idle timeout**: 30 minutes (`FC_SESSION_IDLE_TIMEOUT_MINUTES`) -- extended on each request
- **Absolute timeout**: 12 hours (`FC_SESSION_ABSOLUTE_TIMEOUT_HOURS`) -- hard cap
- `Session.RefreshIdle()` extends idle expiry but never beyond absolute expiry
- `Session.IsExpired()` checks both idle and absolute deadlines
- Sessions are delivered via HTTP cookie (`credentials: 'include'` on the frontend)

### Role-Based Access Control (RBAC)

Three enforcement layers:

1. **Route-level middleware** -- `RequireRole()` Echo middleware blocks requests from unauthorized roles at the route group level
2. **Application service checks** -- Service methods verify the caller's role before performing the action
3. **Object/data-scope filtering** -- Members see only their own orders; operations managers can see all orders but are constrained to their assigned location for member/coach directory data; staff-only personnel endpoints remain service-guarded even if a future handler is added incorrectly

### Role Permissions Matrix

| Module | Administrator | Operations Manager | Procurement Specialist | Coach | Member |
|---|:---:|:---:|:---:|:---:|:---:|
| Dashboard | X | X | -- | X | -- |
| Catalog | X | X | X (read) | -- | X (read) |
| Inventory | X | X | -- | -- | -- |
| Group Buys | X | X | -- | X | X |
| Orders | X | X | -- | X | X (own) |
| Procurement | X | X | X | -- | -- |
| Reports | X | X | X | X | -- |
| Admin | X | -- | -- | -- | -- |
| Audit Log | X | -- | -- | -- | -- |
| Backups | X | -- | -- | -- | -- |
| Biometric | X | -- | -- | -- | -- |
| Users | X | -- | -- | -- | -- |

### Sensitive Field Masking

The `maskField()` utility in `frontend/src/lib/format.ts` masks PII fields (email shows first char + domain, phone shows last 4 digits). Backend `internal/security` package provides server-side masking for API responses and exports based on the caller's role.

### Input Security

- **SQL injection prevention**: All database access uses parameterized queries via pgx (no string concatenation in SQL)
- **XSS prevention**: React's JSX escaping handles output encoding; API responses are JSON-only
- **CSRF**: Session cookies with `credentials: 'include'`; same-origin enforcement

### Optional Biometric Module

- Controlled by `FC_BIOMETRIC_MODULE_ENABLED` environment variable (default: false)
- **AES-256 envelope encryption** for biometric data at rest (`biometric_enrollments.encrypted_data` is BYTEA)
- **Scheduled key rotation** (`FC_BIOMETRIC_KEY_ROTATION_DAYS`, default 90), tracked via `encryption_keys` table and enforced by the background job scheduler
- `EncryptionKey.NeedsRotation()` checks days since activation
- Key status lifecycle: `active` -> `rotated` -> `revoked`
- TLS transport is enforced by default. `FC_TLS_CERT_FILE` and `FC_TLS_KEY_FILE` provide custom certificates, while the API can generate a local self-signed certificate when secure transport is required and no explicit certificate is mounted.

### Tamper-Evident Audit Log

- `audit_events` table with `integrity_hash` (SHA-256) and `previous_hash` columns forming a hash chain
- `AuditEvent.ComputeHash(previousHash)` concatenates event_type, entity_type, entity_id, actor_id, timestamp (RFC3339Nano), details JSON, and previous hash, then computes SHA-256
- Any modification to a historical event breaks the chain, enabling detection

## 8. Background Jobs

All background jobs are designed as goroutines with configurable intervals and graceful shutdown via context cancellation. They are defined across `internal/jobs/jobs.go` and `internal/jobs/procurement_jobs.go`.

| Job | Trigger | Action |
|---|---|---|
| **Unpaid Order Auto-Close** | Periodic scan | Finds orders in `created` status where `NOW() >= auto_close_at` (30 min after creation). Transitions to `auto_closed`. |
| **Group-Buy Cutoff Evaluation** | Periodic scan | Finds `active` campaigns past their `cutoff_time`. Calls `Evaluate()` which sets status to `succeeded` or `failed` based on threshold. |
| **Variance Resolution Deadline** | Periodic scan | Finds `open` variance records past their `resolution_due_date` (5 business days). Triggers escalation for overdue items. Escalation thresholds: >$250 absolute or >2% relative. |
| **Biometric Key Rotation** | Scheduled (daily check) | Ensures an active biometric envelope key exists and rotates it when the configured rotation window or expiry threshold is reached. |
| **Nightly Backup** | Scheduled (nightly) | Invokes the injected `DumpFunc` (wired to `pg_dump` in `cmd/api/main.go`) to produce a database archive at `FC_BACKUP_PATH`, encrypts the archive with AES-256-GCM (required key reference), computes SHA-256 checksum of the final encrypted file, and records metadata in `backup_runs`. |
| **Retention Cleanup** | Scheduled | Scans entities past their retention period and hard-deletes only configured non-audit targets (`orders`, `purchase_orders`, `sessions`, `captcha_challenges`). Per-record deletion events are written to the append-only audit log, and `audit_events` is never purged. |

## 9. Backup Strategy

- **Schedule**: Nightly (configurable), can also be triggered manually via `POST /api/v1/admin/backups`
- **Destination**: Admin-configured path (`FC_BACKUP_PATH`, default `/var/backups/fitcommerce`; overridden to named Docker volume in docker-compose.yml)
- **Encryption**: AES-256-GCM encrypted archive is mandatory; backup trigger fails if `FC_BACKUP_ENCRYPTION_KEY_REF` is unset. Key derived via HKDF-SHA256 (`DeriveKeyFromRef`)
- **Integrity**: SHA-256 checksum computed and stored in `backup_runs.checksum` with `checksum_algorithm = 'sha256'`
- **Metadata**: Every backup run is recorded in the `backup_runs` table with `archive_path`, `checksum`, `encryption_key_ref`, `status`, `file_size`, `started_at`, `completed_at`
- **Status tracking**: `running` → `completed` or `failed`
- **Admin visibility**: `GET /api/v1/admin/backups` lists backup history
- **Restore**: Operational procedure (decrypt archive, verify checksum, run pg_restore). Not automated in the application.
- **pg_dump integration note**: `BackupService` accepts a `DumpFunc` dependency injection point, and the implementation wired in `cmd/api/main.go` shells out to `pg_dump` using `exec.CommandContext`. In containerized deployment, `postgresql-client` is installed in the backend image so `pg_dump` is available by default.

## 10. Retention Defaults

Seeded in `backend/database/seeds/seed.sql` via the `retention_policies` table:

| Entity Type | Retention Period | Description |
|---|---|---|
| `financial_records` | 7 years | Orders, payments, refunds |
| `procurement_records` | 7 years | Purchase orders, supplier invoices, landed costs |
| `access_logs` | 2 years | Login attempts, session records, access audit trails |

`audit_events` is intentionally excluded from retention purge targets to preserve append-only audit integrity.

Enforcement logic in `domain/retention.go`: `IsWithinRetention(createdAt, retentionDays, now)` prevents deletion of records still within their retention window. Attempting to delete a retained record returns `ErrRetentionViolation` (HTTP 403, code `RETENTION_VIOLATION`).

## 11. Requirement-to-Module Mapping Table

| Prompt Domain | Frontend Feature | Backend Package(s) | Database Tables |
|---|---|---|---|
| **Roles & Access Control** | `lib/auth.tsx`, `lib/constants.ts` (ROLE_PERMISSIONS), `routes.tsx` (role guards) | `domain/enums.go` (UserRole), `http/middleware.go` (RequireRole), `security/` | `users` |
| **Dashboard & KPIs** | `features/dashboard/`, `routes/DashboardPage.tsx` | `application/services.go` (DashboardService), `reporting/` | `members`, `coaches`, `orders`, `items` |
| **Exports** | `features/reports/`, `routes/ReportsPage.tsx`, `lib/format.ts` (formatExportFilename) | `application/services.go` (ReportService), `reporting/` | `report_definitions`, `export_jobs` |
| **Catalog Management** | `features/catalog/`, `routes/Catalog*.tsx`, `lib/validation.ts` (item schemas) | `domain/item.go`, `domain/policies.go`, `domain/batch_edit.go`, `application/services.go` (ItemService) | `items`, `item_availability_windows`, `item_blackout_windows`, `batch_edit_jobs`, `batch_edit_results` |
| **Group-Buy Campaigns** | `features/group-buy/`, `routes/GroupBuy*.tsx` | `domain/campaign.go`, `domain/state_machines.go`, `application/services.go` (CampaignService) | `group_buy_campaigns`, `group_buy_participants` |
| **Orders** | `features/orders/`, `routes/Order*.tsx`, `lib/validation.ts` (order schemas) | `domain/order.go`, `domain/state_machines.go`, `application/services.go` (OrderService) | `orders`, `order_timeline_entries`, `fulfillment_groups`, `fulfillment_group_orders` |
| **Procurement** | `features/procurement/`, `routes/Procurement*.tsx`, `routes/Suppliers*.tsx`, `routes/PurchaseOrder*.tsx`, `routes/LandedCosts*.tsx` | `domain/purchase_order.go`, `domain/supplier.go`, `domain/variance.go`, `domain/state_machines.go`, `application/services.go` (SupplierService, PurchaseOrderService, VarianceService, LandedCostService) | `suppliers`, `purchase_orders`, `purchase_order_lines`, `variance_records`, `landed_cost_entries` |
| **Security** | `lib/auth.tsx`, `lib/validation.ts`, `lib/format.ts` (maskField) | `security/`, `domain/user.go`, `domain/errors.go`, `http/middleware.go` | `users`, `sessions`, `captcha_challenges` |
| **Biometric** | `routes/BiometricPage.tsx` | `domain/biometric.go`, `security/` | `biometric_enrollments`, `encryption_keys` |
| **Audit & Retention** | `routes/AuditPage.tsx` | `domain/audit.go`, `domain/retention.go`, `application/services.go` (AuditService) | `audit_events`, `retention_policies` |
| **Backup** | `routes/BackupsPage.tsx` | `domain/backup.go`, `jobs/`, `application/services.go` (BackupService) | `backup_runs` |
| **Inventory** | `features/inventory/`, `routes/InventoryPage.tsx` | `domain/inventory.go`, `application/services.go` (InventoryService) | `inventory_snapshots`, `inventory_adjustments`, `warehouse_bins` |
| **Offline Operation** | `lib/api-client.ts` (relative URLs, offline-aware network errors), `lib/offline-cache.ts` (IndexedDB query persistence), `app/providers.tsx` (cache hydration + read-first offline behavior), `frontend/Dockerfile` (nginx serves static) | All packages (no external calls) | Local PostgreSQL |
| **Locations & Organization** | `lib/types.ts` (Location, Member, Coach) | `domain/location.go`, `domain/member.go` | `locations`, `members`, `coaches` |

## Prompt 4: Core Backend Workflows

### Catalog Publish Validation Flow

Items must pass `domain.ValidateItemForPublish` (name, category, brand, condition, billing_model all required; availability/blackout overlap detection via `DetectWindowOverlap`) before status transitions to Published. `ErrPublishBlocked` carries the full list of failure reasons and maps to HTTP 422.

### Batch-Edit Atomicity

Each row in a batch edit is independently validated and committed as its own atomic unit. A row failure stores a reason string in `BatchEditResult.FailureReason` and `Success=false`, and has zero effect on any other row — successful rows remain committed, failed rows remain unchanged.

### Order Cancellation Policy

Members may cancel their own order only if it is in `Created` (unpaid) status. ManageOrders roles (Admin, OpsMgr) may cancel any order in `Created` or `Paid` status. The router applies `authMW` only for `POST /orders/:id/cancel`; the cancellation handler enforces ownership+status checks. Rationale: self-service before payment, staff-controlled after.

### Order State Machine and Inventory Impact

- `Created` → `Paid`: no inventory change (item still reserved)
- `Created` → `Cancelled`: inventory restored (+quantity)
- `Created` → `AutoClosed`: inventory restored (+quantity)
- `Paid` → `Cancelled`: inventory restored (+quantity)
- `Paid` → `Refunded`: inventory restored (+quantity)
- **Split**: original cancelled (record-keeping), child orders inherit the same reservation. Net inventory change = 0. No adjustments.
- **Merge**: originals cancelled (record-keeping), merged order created for same total quantity. Net inventory change = 0. No adjustments.

### Fulfillment Grouping — Deferred to Prompt 7

Split/merge at Prompt 4 is order-record-level only: split creates N child orders from one parent, merge creates one consolidated order from multiple source orders. Grouping by supplier, warehouse bin, and pickup point (`FulfillmentGroup`, `FulfillmentGroupOrder`) is fully deferred to Prompt 7.

### Campaign Lifecycle

Active → `EvaluateAtCutoff` (job or manual trigger) → Succeeded (participating orders remain active) / Failed (participating Created orders auto-closed, inventory restored per order).

### Background Jobs

- **AutoCloseJob**: polls every 1 minute (`ListExpiredUnpaid` → `TransitionOrder(AutoClosed)` → restore inventory). Exits on context cancellation.
- **CutoffEvalJob**: polls every 1 minute (`ListPastCutoff` → `EvaluateAtCutoff` per campaign). Exits on context cancellation.

Both jobs are started as goroutines in `cmd/api/main.go` before the shutdown-signal wait.

## Prompt 5: React Application Shell and Client Architecture

### Frontend Shell Architecture

The frontend is a Vite + React 18 + TypeScript single-page application served on port 5173 (dev) / 3000 (Docker). It uses:
- **React Router v7** with nested routes and a single shared Layout (sidebar + AppBar + `<Outlet />`)
- **TanStack Query v5** for server state (staleTime 30s, retry 1, no refetch on focus)
- **MUI v6** for all UI components; custom theme (primary: #1B5E20, secondary: #FF6F00)
- **React Hook Form + Zod** for form validation
- **react-router-dom `createBrowserRouter`** with lazy-loaded route components

### Route Map

All authenticated routes share `Layout` (sidebar navigation + AppBar) as a parent route element, protected by `ProtectedRoute`. Role-restricted child routes use a second `ProtectedRoute` with `allowedRoles`.

| Path | Component | Role restriction |
|------|-----------|-----------------|
| `/login` | `LoginPage` | Public |
| `/dashboard` | `DashboardPage` | Any authenticated |
| `/catalog`, `/catalog/*` | `CatalogPage`, `CatalogDetailPage`, `CatalogFormPage` | administrator, operations_manager |
| `/inventory` | `InventoryPage` | administrator, operations_manager |
| `/group-buys`, `/group-buys/:id` | `GroupBuysPage`, `GroupBuyDetailPage` | Any authenticated |
| `/orders`, `/orders/:id` | `OrdersPage`, `OrderDetailPage` | Any authenticated |
| `/procurement/*` | Procurement pages | administrator, operations_manager, procurement_specialist |
| `/reports` | `ReportsPage` | Any authenticated (content gated per role) |
| `/admin/*` | Admin pages | administrator only |

The `ProtectedRoute` at the parent level redirects unauthenticated users to `/login` (preserving `state.from` for post-login redirect). Role-restricted routes redirect unauthorized roles to `/dashboard`.

### State Boundaries

- **Server state**: all entity data (items, orders, campaigns, inventory, dashboard KPIs) managed by TanStack Query with domain-specific hook files in `src/lib/hooks/`
- **Auth state**: `AuthContext` (login, logout, session refresh, captcha/lockout states)
- **Notification state**: `NotificationsProvider` (toast queue, Snackbar lifecycle)
- **UI-local state**: toggle selections, filter values, dialog open/close — held in component `useState`
- **No global store (Redux/Zustand)** — server state in query cache, local UI state local

### Dashboard Composition Strategy

`DashboardPage` uses a `DashboardFilters` object (period, location_id, coach_id, category, from, to) as the query key for `useDashboardKPIs`. The page renders:
1. Period toggle buttons (Daily / Weekly / Monthly / Quarterly / Yearly)
2. `FilterBar` for location, coach, category, date range (visible to administrator, operations_manager, coach via `RequireRole`)
3. `StatCard` grid for membership & engagement KPIs (all roles)
4. Operations KPIs section — `class_fill_rate`, `coach_productivity` (administrator, operations_manager, coach only via `RequireRole`)

All sections degrade gracefully: loading → Skeleton in StatCard; API error → Alert warning; no data → `'—'` placeholder values.

### Navigation and Role-Aware Sidebar

`NavSidebar` reads `ROLE_PERMISSIONS[user.role]` from `constants.ts` and filters `ALL_NAV_ITEMS` to show only modules the current role can access. Active route highlighted via `useLocation`. Sidebar is a permanent MUI `Drawer` (240px wide, dark green background).

### Shared UI Primitives

| Component | Purpose |
|-----------|---------|
| `DataTable<T>` | Server-paginated table with columns config, loading skeleton, empty state, error alert |
| `FilterBar` | Reusable filter bar with text, select, and date field types; onChange callback; clear button |
| `PageContainer` | Page wrapper with title, breadcrumbs, and actions slot |
| `StatCard` | KPI card with value, label, loading skeleton, and change indicator |
| `ConfirmDialog` | Blocking confirmation dialog with destructive variant |
| `EmptyState` | Illustrated empty state with optional action button |
| `ErrorBoundary` | React class error boundary wrapping the Layout outlet |

### Data-Fetching Hooks

Domain-specific hooks in `src/lib/hooks/`:
- `useDashboard.ts` — `useDashboardKPIs(filters: DashboardFilters)`
- `useItems.ts` — `useItemList`, `useItem`, `usePublishItem`, `useUnpublishItem`, `useUpdateItem`, `useBatchEdit`
- `useOrders.ts` — `useOrderList`, `useOrder`, `useOrderTimeline`, `useCancelOrder`, `usePayOrder`, `useRefundOrder`, `useAddOrderNote`
- `useCampaigns.ts` — `useCampaignList`, `useCampaign`, `useJoinCampaign`, `useCancelCampaign`, `useEvaluateCampaign`
- `useInventory.ts` — `useInventorySnapshots`, `useInventoryAdjustments`, `useCreateAdjustment`, `useWarehouseBins`, `useWarehouseBin`

All mutations call `queryClient.invalidateQueries` on success to keep the cache consistent.

### Global UX Infrastructure

- **Toasts**: `NotificationsProvider` wraps the app; `useNotify()` returns `{ success, error, warning, info }` functions using MUI Snackbar + Alert (queued, auto-dismiss 5s)
- **Error boundary**: `ErrorBoundary` class component wraps `<Outlet />` in Layout; shows an inline error card with a "Try again" reset button
- **Session expiry**: `apiClient` dispatches `auth:session-expired` CustomEvent on 401; `AuthProvider` listens and clears auth state, triggering redirect to `/login`
- **Unauthorized redirects**: `ProtectedRoute` redirects unauthenticated users to `/login` and wrong-role users to `/dashboard`

## Prompt 6: Primary Operational Screens

### Screen Inventory

| Route | Component | Roles |
|-------|-----------|-------|
| `/catalog` | `CatalogPage` | admin, ops_mgr |
| `/catalog/new` | `CatalogFormPage` | admin, ops_mgr |
| `/catalog/:id` | `CatalogDetailPage` | admin, ops_mgr |
| `/catalog/:id/edit` | `CatalogFormPage` | admin, ops_mgr |
| `/inventory` | `InventoryPage` | admin, ops_mgr |
| `/group-buys` | `GroupBuysPage` | all authenticated |
| `/group-buys/:id` | `GroupBuyDetailPage` | all authenticated |
| `/orders` | `OrdersPage` | all authenticated |
| `/orders/:id` | `OrderDetailPage` | all authenticated |
| `/reports` | `ReportsPage` | role-gated sections |

### Catalog Screens

**CatalogPage** renders a server-paginated `DataTable<Item>` with `FilterBar` (category, brand, condition, status). The "New Item" button is gated by `RequireRole(['administrator','operations_manager'])`. Filtering resets the page to 0.

**CatalogDetailPage** renders an item detail card with all fields. Publish/Unpublish actions use `ConfirmDialog` and call the publish/unpublish mutations. Admin/ops_mgr role gated. After publish/unpublish, cache is invalidated via TanStack Query.

**CatalogFormPage** is both create and edit mode — distinguished by presence of `params.id`. Uses `react-hook-form` + `zodResolver(createItemSchema)`. On 409 conflict, a root error is set with a "reload and try again" message. On success, navigates to the detail page. On edit mode, loads existing item and resets form via `useEffect`.

### Inventory Screen

**InventoryPage** uses MUI `Tabs` to switch between Snapshots and Adjustments views. Snapshots are fetched as a non-paginated list; adjustments are paginated. The "Create Adjustment" action opens a `Dialog` form with `item_id`, `quantity_change`, and `reason` fields, validated with `inventoryAdjustmentSchema`. Gated to admin/ops_mgr.

### Group Buy Screens

**GroupBuysPage** lists campaigns with a `LinearProgress` column showing `current_quantity / min_quantity`. Staff can open the generic "Create Campaign" dialog from the page, while members can start a campaign only from an item-driven entry point that pre-fills and locks the selected catalog item. Status filter uses the `FilterBar` select field.

**GroupBuyDetailPage** displays campaign header, description, a `LinearProgress` bar with labels, and key metadata. The Join form is visible only to `member` role using `RequireRole(['member'])`. Cancel/Evaluate staff actions are visible to admin/ops_mgr only. Both destructive actions use `ConfirmDialog`.

### Order Screens

**OrdersPage** lists orders with a status `FilterBar`. Rows show truncated IDs (first 8 chars). Clicking an order ID navigates to `/orders/:id`. Backend enforces role-based filtering (members see own orders only).

**OrderDetailPage** shows order details and an operation timeline. Action buttons are role-aware:
- Cancel: members can cancel own Created orders; ManageOrders roles can cancel Created or Paid
- Pay (Record Payment): ManageOrders only, opens a Dialog for settlement marker
- Refund: ManageOrders only, opens a ConfirmDialog
- Add Note: ManageOrders only, opens a Dialog with a textarea

### Reports Screen

**ReportsPage** renders the role-filtered report catalog returned by the backend, supports CSV/PDF export initiation, and keeps export/download actions explicitly online-only in the current read-first offline rollout. Completed export jobs are listed in-page with download actions gated behind connectivity.

### Shared Component: StatusChip

`StatusChip` wraps MUI `Chip` with pre-defined `COLOR_MAP` and `LABEL_MAP` for all item, order, campaign, and PO statuses. Used across all listing and detail pages. Falls back to `'default'` color and the raw status string as label for unknown values.

### State Management Pattern

All pages follow the same pattern:
- `page` state: 0-indexed (matches MUI `TablePagination`); pass `page + 1` to hooks
- `filters` state: `Record<string, string>`, reset page to 0 on filter change
- `isLoading`, `error` from TanStack Query propagated directly to `DataTable`
- Mutations: call `mutateAsync`, catch error, show notification via `useNotify()`
