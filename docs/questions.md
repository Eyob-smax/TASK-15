# questions.md

## 1. Offline-first scope between browser clients and the local backend
**The Gap**  
The prompt says the suite is “offline-first” and that the React UI consumes the Go API over the local network or same machine, but it does not say whether browser clients must continue functioning independently when the backend is unreachable.

**The Interpretation**  
Treat offline-first as **internet-independent local deployment**, not as a fully standalone browser app. The backend remains the source of truth on the local machine or LAN, and the React client is expected to function without internet but still requires reachability to the local Go/PostgreSQL stack for authoritative writes and reconciled reads.

**Proposed Implementation**  
Implement a local-network-first web architecture with resilient client UX: cached read models where appropriate, retry-safe form preservation, unsaved-change protection, and clear degraded-state messaging when the backend is unreachable. Do not invent a second client-side source of truth that could conflict with PostgreSQL.

## 2. Exact KPI formulas for dashboard reporting
**The Gap**  
The prompt requires KPIs for member growth, churn, renewal rate, engagement, class fill rate, and coach productivity, but it does not define the exact formulas, denominator rules, or which records count toward each metric.

**The Interpretation**  
Adopt stable operational definitions that can be documented and audited. Use time-bucketed metrics based on the selected dashboard period (daily/weekly/monthly), and keep formulas explicit in code and docs.

**Proposed Implementation**  
Define and document formulas as follows unless later repository context overrides them: member growth = new active memberships in period; churn = memberships ended or lapsed in period; renewal rate = renewed memberships / memberships due to renew in period; engagement = attended sessions or qualified interactions per active member; class fill rate = occupied spots / scheduled spots; coach productivity = completed coached sessions plus fulfilled operational actions attributed to that coach. Publish the formulas in `docs/design.md` and `docs/api-spec.md` so report behavior is statically reviewable.

## 3. What counts as “committed quantity” for group-buy success at cutoff
**The Gap**  
The prompt says campaigns succeed only if committed quantity meets the threshold at cutoff, while orders use offline settlement markers and unpaid orders auto-close after 30 minutes. It does not explicitly define whether a merely created order contributes to committed quantity.

**The Interpretation**  
Committed quantity should mean quantity attached to orders that are still valid at cutoff and have an accepted offline settlement marker, not simply draft/intake orders.

**Proposed Implementation**  
Model commitment as a derived quantity from orders in an eligible paid/settlement-confirmed state at cutoff time. Created-but-unpaid orders reserve temporary intent only; they do not count toward campaign success once they auto-close or remain unpaid at cutoff. Document this clearly in the campaign/order state machine and related tests.

## 4. Role boundaries for Coach and Member on reporting and operational data
**The Gap**  
The prompt defines five roles, but only partially describes what Coach and Member can see beyond “class readiness and limited reporting” and “browse items and join group-buys.” The exact limits for dashboards, order visibility, procurement data, and audit exposure are not fully spelled out.

**The Interpretation**  
Use a least-privilege model. Coaches get operational views limited to their assigned classes, readiness, and explicitly allowed reports. Members get self-service item browsing, campaign participation, and access only to their own orders and outcomes. They do not access procurement, inventory-control, or security/audit surfaces.

**Proposed Implementation**  
Create an explicit permission matrix in backend middleware and frontend route/action guards. Scope Coach access by assigned location/class/coach identity. Scope Member access to self-owned campaigns, orders, and downloadable documents only. Keep all admin, procurement, backup, audit, biometric, and global reporting functions restricted to staff roles.

## 5. Time-zone policy for availability windows, cutoff times, and nightly jobs
**The Gap**  
The prompt gives human-readable availability windows and cutoff behavior but does not define the authoritative time zone for item availability, group-buy cutoff evaluation, session expiration, or nightly backups.

**The Interpretation**  
Assume a single club-configured local time zone governs business semantics, while persistence stores canonical UTC timestamps plus the configured display zone.

**Proposed Implementation**  
Add an organization/system time-zone setting in configuration. Store timestamps in UTC in PostgreSQL, convert to the configured club zone for UI display and rule evaluation inputs, and use that same configured zone for cutoff jobs, nightly backups, and timestamped export filenames.

## 6. Landed-cost allocation method for procurement reconciliation
**The Gap**  
The prompt requires landed-cost rollups per item and period for reconciliation summaries but does not define how shared freight/fees are allocated across received lines when a purchase order spans multiple items.

**The Interpretation**  
Use a documented allocation strategy that is simple, auditable, and maintainable. A value-weighted or quantity-weighted allocation must be chosen explicitly.

**Proposed Implementation**  
Default to value-weighted landed-cost allocation across received lines within a procurement receipt, with an override hook for direct line-specific charges when the receipt explicitly attributes a cost to a single item. Persist both raw cost components and computed allocations so reconciliation summaries are explainable and testable.

## 7. Biometric module scope and storage boundaries
**The Gap**  
The prompt says biometric data is encrypted “if stored for facility access,” which implies the biometric module is conditional, but it does not specify whether the initial delivery must include active biometric enrollment/capture workflows or only storage/security support.

**The Interpretation**  
Treat biometric handling as an optional but supported module. The initial delivery must include secure data structures, encryption, masking, retention, and admin controls for biometric records, but should not invent hardware-device integration or mandatory enrollment flows unless other repository context requires it.

**Proposed Implementation**  
Implement biometric record models, encrypted storage, key-rotation metadata, role-aware admin views, and clear enable/disable configuration. Keep capture-device specifics abstracted behind interfaces and leave hardware-specific enrollment adapters out of scope unless an existing repo contract demands them.

## 8. Backup encryption key custody and restore verification flow
**The Gap**  
The prompt requires nightly encrypted local backups with checksum verification, but it does not state where backup encryption keys come from, how restore authorization works, or how restore verification should be documented.

**The Interpretation**  
Assume the system manages backup encryption through locally configured admin-controlled key material, with restore restricted to privileged staff and documented as an operational workflow rather than silently automatic behavior.

**Proposed Implementation**  
Add backup configuration for archive path, encryption-key reference or passphrase-derived key material, checksum algorithm, and restore authorization policy. Persist backup run metadata and verification status in PostgreSQL, expose admin-visible history, and document the restore assumptions honestly in `repo/README.md` and `docs/design.md`.

## 9. Group-buy inventory reservation timing
**The Gap**  
The prompt requires that campaigns succeed only when committed quantity meets the threshold at cutoff, but it does not specify when inventory is decremented for group-buy participants — at join time, at payment confirmation, or only at campaign success.

**The Interpretation**  
Inventory should be decremented at join time when a participant commits to a group-buy campaign. If the campaign fails at cutoff, the participating orders are auto-closed and inventory is restored. This is consistent with how finite-quantity items are managed and prevents overselling during active campaigns.

**Proposed Implementation**  
Decrement inventory on participant join. Track reserved quantity separately from available quantity in `inventory_snapshots` so the UI can display both. On campaign failure or order auto-close, issue a compensating adjustment to restore the reserved stock.

## 10. Split/merge order inventory — net-zero adjustment
**The Gap**  
The prompt supports order splitting and merging but does not define whether these operations should trigger inventory adjustments, since they move quantity between order records without changing the total committed stock.

**The Interpretation**  
Split and merge do not change the net quantity committed to an item. Child orders inherit the reservation from the original. The original order's cancellation in a split/merge context is a record-keeping action, not a true inventory release.

**Proposed Implementation**  
Do not create inventory adjustments for split or merge operations. Model them as pure order-record transformations. The existing reservation is transferred to child orders implicitly. Document this in the order state machine and related tests to prevent future confusion with real cancellations.

## 11. Dashboard KPI value format (fractional vs. percentage)
**The Gap**  
The prompt requires KPI metrics for member growth, churn, renewal rate, and engagement, but it does not specify whether the API returns raw fractions (e.g., 0.05 for 5%) or whole percentage values (e.g., 5.0 for 5%), which affects how the frontend formats them.

**The Interpretation**  
Adopt the standard API convention of fractional values (0.0–1.0 range) for all percentage-type KPIs. The frontend `formatPercentage` helper receives the raw fractional value and is responsible for display formatting.

**Proposed Implementation**  
Return all percentage-type KPI fields as fractional floats from the backend. The `formatPercentage` utility in the frontend formats without multiplying by 100 — it expects values already in fractional form. Document this contract in `docs/api-spec.md` when the dashboard endpoint is implemented.

## 12. Dashboard KPI endpoint stub behavior
**The Gap**  
The prompt requires a KPI dashboard but the KPI computations depend on attendance and session-data models that are not yet implemented. It is unclear whether the endpoint should return a placeholder response or signal its incomplete state explicitly.

**The Interpretation**  
`GET /api/v1/dashboard/kpis` should return `501 Not Implemented` rather than a fake empty response. Returning `{}` would mask the incomplete state from consumers and make it harder to detect when the endpoint is genuinely unimplemented.

**Proposed Implementation**  
Register `GET /api/v1/dashboard/kpis` in the router using the `notImplemented` handler. The frontend `useDashboardSummary` hook handles the resulting error state gracefully, showing a warning alert instead of crashing. The endpoint will be wired once attendance and session-data models are implemented.

## 13. PDF export library choice
**The Gap**  
The prompt requires CSV and PDF report exports but does not specify which PDF generation library to use, how complex the output should be, or whether charts and rich formatting are required.

**The Interpretation**  
Use a lightweight, dependency-free PDF library. Charts, rich formatting, and embedded images are deferred scope. The initial delivery requires only tabular data output in a readable PDF layout.

**Proposed Implementation**  
PDF generation uses `github.com/jung-kurt/gofpdf` v1.16.2. Output is a simple tabular layout (Arial font, A4 portrait). The library is declared in `go.mod` and resolves during Docker build — no host-side `go get` required. Charts and rich formatting are deferred.

## 14. Biometric key rotation scope
**The Gap**  
The prompt requires biometric key rotation but does not specify whether rotating the active encryption key should retroactively re-encrypt all existing enrolled biometric templates, or only apply to future registrations.

**The Interpretation**  
Key rotation should create a new active key for future `Register` operations only. Retroactive re-encryption of existing enrolled templates is a migration-level operation that requires careful coordination of all enrolled records and is out of scope for the initial delivery.

**Proposed Implementation**  
`BiometricService.RotateKey` creates a new active encryption key and marks the previous key as rotated. Existing enrolled templates in `biometric_enrollments.encrypted_data` are not re-encrypted. The audit event emitted on rotation documents this explicitly: `"new key active for future operations only; existing templates not re-encrypted"`.

## 15. Retention enforcement scope
**The Gap**  
The prompt requires configurable retention policies with scheduled enforcement, but it does not define whether enforcement means hard delete, soft delete, or archival — and the current schema does not include `deleted_at` columns on affected tables.

**The Interpretation**  
Retention enforcement should be log-only in the current release. The schema lacks the `deleted_at` columns required for safe soft-delete enforcement. Hard-delete enforcement without soft-delete markers would be irreversible and unauditable.

**Proposed Implementation**  
`RetentionCleanupJob` (runs every 24 h) calls `RetentionService.RunCleanup`, which logs which records would be eligible for purge under each policy's `retention_years` threshold. No DELETE statements are issued. Hard-delete enforcement requires a future schema migration to add `deleted_at` columns to all affected tables before the cleanup job can act on them.

## 16. Backup actor tracking for scheduled vs. manual runs
**The Gap**  
The prompt requires both scheduled nightly backups and manually triggered backups from the admin panel, but it does not define how the system should distinguish between a system-initiated backup and a user-initiated one in the audit record.

**The Interpretation**  
The `BackupRun` domain struct should not carry a trigger-actor field — actor identity is an audit concern, not an operational state concern. A nil actor indicates a system-scheduled run; a non-nil actor identifies the user who triggered it manually.

**Proposed Implementation**  
`BackupService.Trigger` accepts `performedBy *uuid.UUID`. Nil means system-scheduled (passed by `BackupJob`); non-nil means a specific user triggered it from the admin panel (passed by `BackupHandler`). The actor is recorded only in the associated `backup.completed` audit event, not on the `BackupRun` struct itself.

## 17. Frontend test coverage tooling
**The Gap**  
The project requires coverage reporting for frontend tests but does not specify which Vitest coverage provider to use. Two providers are available: V8 (built-in) and Istanbul (requires Babel transformation).

**The Interpretation**  
Use the V8 provider. It requires no Babel transformation and works natively with the existing Vite + TypeScript setup. Istanbul was considered but rejected because it adds a build-time dependency that conflicts with the native ESM setup.

**Proposed Implementation**  
Frontend coverage uses `@vitest/coverage-v8`, added as a dev dependency in `package.json` and configured in `vite.config.ts` under `test.coverage`. The `--coverage` flag in `run_tests.sh` activates it. Text and lcov reporters are configured.

## 18. Backend coverage reporting
**The Gap**  
The project requires coverage tracking for backend tests but does not define the output format, tool, or how to summarize coverage across multiple test suites running separately.

**The Interpretation**  
Use Go's built-in `-coverprofile` flag to produce coverage data per suite. Print a summary line from `go tool cover -func` after each suite. Do not automate HTML report opening, as that requires a browser and is inappropriate for CI environments.

**Proposed Implementation**  
Backend unit and API tests are run with `-coverprofile` flags, producing `coverage_unit.out` and `coverage_api.out` in `backend/`. A summary line (`go tool cover -func | tail -1`) is printed after each suite. A merged HTML report can be generated manually with `go tool cover -html=coverage_unit.out` but is not automated in `run_tests.sh`.

## 19. VariancesPage overdue indicator without resolution_due_date in the frontend type
**The Gap**  
The `VarianceRecord` TypeScript type does not include a `resolution_due_date` field — the backend returns `escalated: boolean` in the frontend-facing shape. The UI requires an overdue indicator but has no date field to compare against.

**The Interpretation**  
Use the `escalated` field as the overdue proxy. When the backend marks a variance as escalated, it is effectively overdue by the business definition (past its 5-business-day resolution window). This avoids adding a field to the shared type that the API does not currently return.

**Proposed Implementation**  
`isOverdue` in `VariancesPage.tsx` evaluates to `v.escalated === true`. The `Overdue` chip is shown whenever a variance is escalated. `VariancesPage.test.tsx` verifies this by setting `escalated: true` on a mock variance and asserting the chip renders.

## 20. AdminPage test scope
**The Gap**  
`AdminPage.tsx` is a minimal placeholder component that renders only a heading. Writing tests for features it does not implement would test nothing real, but omitting tests entirely would leave the component with zero coverage.

**The Interpretation**  
Test only what exists in the component. Detailed feature tests belong in the sub-page tests (`UsersPage.test.tsx`, `AuditPage.test.tsx`) and API tests (`admin_test.go`), not in `AdminPage.test.tsx`.

**Proposed Implementation**  
`AdminPage.test.tsx` verifies only that the admin page heading renders correctly (2 tests). No additional tests are added to this file because there is no additional behavior to test in the component itself.

## 21. CatalogFormPage edit-mode pre-fill test timing
**The Gap**  
The CatalogFormPage edit-mode test must assert that form fields are pre-filled from the existing item data, but the pre-fill happens inside a `useEffect` that runs after the first render, making synchronous assertions unreliable.

**The Interpretation**  
The test must wait for the `useEffect` to complete before asserting field values. Using `waitFor` from React Testing Library is the correct approach, as it retries the assertion until it passes or times out.

**Proposed Implementation**  
The edit-mode test uses `waitFor` before asserting the name field value and before clicking Save. This is necessary because the form calls `reset(existingItem)` inside a `useEffect`, which fires asynchronously after the initial render. Without `waitFor`, the assertion runs before the reset has taken effect.

## 22. Frontend Dockerfile nginx configuration portability
**The Gap**  
The original `frontend/Dockerfile` used Docker's heredoc syntax (`COPY <<'EOF' ... EOF`) to write the nginx config inline. Heredoc support in COPY requires Docker BuildKit, which is not enabled by default on Docker 20.10 and is not guaranteed in all deployment environments.

**The Interpretation**  
Extract the nginx configuration to a separate file and use a standard `COPY` instruction. This is universally compatible with Docker 20.10+ and does not require BuildKit.

**Proposed Implementation**  
The nginx configuration lives in `frontend/nginx.conf`. The Dockerfile uses `COPY nginx.conf /etc/nginx/conf.d/default.conf`. No heredoc syntax is used anywhere. The change is backward-compatible with all Docker 20.10+ environments.

## 23. TLS termination approach for LAN deployment
**The Gap**  
The backend supports TLS via `FC_TLS_CERT_FILE` and `FC_TLS_KEY_FILE`, but it is unclear whether TLS should be configured in `docker-compose.yml` or left to the operator.

**The Interpretation**  
TLS termination in a LAN deployment is typically handled by the operator at the infrastructure level (e.g., a reverse proxy or load balancer), not inside the application compose stack. Embedding TLS configuration in `docker-compose.yml` would require operators to manage certificate files inside the compose context, which is error-prone.

**Proposed Implementation**  
`FC_TLS_CERT_FILE` and `FC_TLS_KEY_FILE` are absent from `docker-compose.yml`. The backend starts in plain HTTP mode by default, which is appropriate for local/LAN-only use. Operators who need TLS can add the vars to a `.env` file or compose override and mount the certificate files as volumes.

## 24. Backup and export paths in Docker volumes
**The Gap**  
`FC_BACKUP_PATH` and `FC_EXPORT_PATH` are configured as filesystem paths inside the container, but it is unclear whether these paths should be host-relative or container-internal, and how operators access the files they contain.

**The Interpretation**  
These paths should be container-internal, backed by named Docker volumes for persistence across container restarts. Operators who need host access should bind-mount a host directory instead of relying on named volumes.

**Proposed Implementation**  
`FC_BACKUP_PATH=/var/backups/fitcommerce` and `FC_EXPORT_PATH=/var/exports/fitcommerce` are set in `docker-compose.yml` and backed by named Docker volumes (`backups` and `exports`). These paths exist only inside the container. To access files on the host, operators should inspect the named volume (`docker volume inspect`) or bind-mount a host path instead.

## 25. Inventory snapshot RBAC correction
**The Gap**  
`GET /inventory/snapshots` was guarded by middleware that included `ActionViewCatalog`, which is granted to all five roles including Coach and Member. This meant any authenticated user could read raw stock-level data, which contradicts the documented RBAC matrix restricting inventory access to operational roles.

**The Interpretation**  
Inventory snapshots are operational data. Access should be restricted to roles that actively manage stock or make procurement decisions: Administrator, Operations Manager, and Procurement Specialist. Coach and Member must not see raw inventory quantities.

**Proposed Implementation**  
The snapshot route is now guarded by `NewRequireRole(ActionManageInventory, ActionManageProcurement)`. This grants access to Administrator and OperationsManager (via `manage_inventory`) and ProcurementSpecialist (via `manage_procurement`). The test server in `inventory_test.go` was updated to match, and two new tests were added: `TestInventorySnapshots_ProcurementSpecialist_200` and `TestInventorySnapshots_Coach_403`.

## 26. Authentication requirement for deferred stub routes
**The Gap**  
The `/locations/*`, `/coaches/*`, and `/members/*` route groups were registered without authentication middleware, making them publicly accessible even though every other route group in the API requires a valid session.

**The Interpretation**  
All routes in the API should require authentication by default, even if they currently return 501. Unauthenticated access to stubs is inconsistent and creates a future hardening gap when these routes are implemented.

**Proposed Implementation**  
`registerLocationRoutes`, `registerCoachRoutes`, and `registerMemberRoutes` now accept `authMW echo.MiddlewareFunc` and call `g.Use(authMW)` before registering their routes. They still return 501 for all methods. Individual RBAC guards will be added when the routes are implemented.

## 27. pg_dump integration is a placeholder in the current implementation
**The Gap**  
The project requires nightly encrypted database backups, but integrating `pg_dump` as a subprocess requires the binary to be present in the container and a shell-out implementation. The backup service must still function during development without a real `pg_dump`.

**The Interpretation**  
The `BackupService` should accept a `DumpFunc` dependency injection point so the implementation can be swapped without changing the service. The default wired in `cmd/api/main.go` is a placeholder that must be clearly documented as non-functional for production use.

**Proposed Implementation**  
`pgDumpFn` in `cmd/api/main.go` creates an empty file and logs `"backup dump placeholder"`. The backup archive file is created and its SHA-256 checksum is recorded, but the file contains no database data. For production deployment, replace `pgDumpFn` with a real shell-out: `exec.CommandContext(ctx, "pg_dump", "-Fc", connStr, "-f", destPath).Run()`. The checksum, encryption, and audit pipeline are fully functional regardless of the DumpFunc implementation. Documented in README.md and docs/design.md.

## 28. FC_EXPORT_PATH default vs. Docker volume override
**The Gap**  
`config.go` defines a default value for `FC_EXPORT_PATH` suitable for local development, while `docker-compose.yml` overrides it to a path backed by a named Docker volume. This creates an apparent mismatch that could confuse operators comparing config defaults to the compose file.

**The Interpretation**  
Both values are correct for their respective contexts. The `config.go` default is a safe local-dev path; the compose override ensures persistence inside the container. The distinction should be documented explicitly to avoid operator confusion.

**Proposed Implementation**  
`config.go` default for `FC_EXPORT_PATH` is `/tmp/fitcommerce-exports` (writable on any Linux host, suitable for local development without Docker). `docker-compose.yml` overrides this to `/var/exports/fitcommerce` backed by a named Docker volume, ensuring exports persist across container restarts. The README env-var table documents both values.
