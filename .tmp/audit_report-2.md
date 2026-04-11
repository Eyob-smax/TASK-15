# FitCommerce Static Delivery Acceptance and Architecture Audit

## 0. Re-review Update (2026-04-11, Static-Only)

- This addendum supersedes the original verdict and findings below.
- Re-review scope: static inspection of current backend/frontend implementation and tests only.
- Runtime execution was still not performed in this update pass.

### 0.1 Updated Verdict

- Overall conclusion: Partial Pass

### 0.2 High Findings From Prior Audit: Current Status

1. Pricing semantics (order price sourced from refundable deposit)
- Status: Resolved
- Evidence: repo/backend/internal/application/order_service.go:63, repo/backend/internal/application/campaign_service.go:120

2. Actor attribution in payment/split/merge/auto-close timeline and audit paths
- Status: Resolved
- Evidence: repo/backend/internal/http/order_handler.go:128, repo/backend/internal/application/order_service.go:138, repo/backend/internal/application/order_service.go:283, repo/backend/internal/application/order_service.go:399, repo/backend/internal/application/order_service.go:572, repo/backend/internal/application/order_service.go:577

3. Missing timeline provenance for campaign-origin orders
- Status: Resolved
- Evidence: repo/backend/internal/application/campaign_service.go:18, repo/backend/internal/application/campaign_service.go:134, repo/backend/internal/bootstrap/app.go:125

4. Export generation silently swallowing report query errors
- Status: Resolved
- Evidence: repo/backend/internal/application/report_service.go:298, repo/backend/internal/application/report_service.go:300

5. Offline-first mutation path unavailable for catalog/report actions
- Status: Partially Resolved
- Evidence (queue added): repo/frontend/src/lib/offline-cache.ts:14, repo/frontend/src/lib/offline-cache.ts:149, repo/frontend/src/lib/hooks/useItems.ts:75, repo/frontend/src/lib/hooks/useReports.ts:43, repo/frontend/src/app/providers.tsx:62, repo/frontend/src/app/providers.tsx:82, repo/frontend/src/routes/CatalogFormPage.tsx:134, repo/frontend/src/routes/ReportsPage.tsx:201

### 0.3 Remaining Risks (Post-fix)

1. Severity: High
- Title: Offline replay can drop queued mutations on non-offline errors without operator visibility
- Evidence: repo/frontend/src/lib/offline-mutations.ts:51, repo/frontend/src/lib/offline-mutations.ts:54
- Why it matters: server-side validation/conflict failures during replay are removed from queue, which risks silent data loss and weakens operational traceability.
- Minimum actionable fix: retain failed entries with terminal error state, surface per-entry replay errors in UI, and require explicit user/operator resolution.

2. Severity: Medium
- Title: Offline queued create flow uses synthetic client IDs without reconciliation mapping
- Evidence: repo/frontend/src/lib/hooks/useItems.ts:126, repo/frontend/src/lib/hooks/useItems.ts:131, repo/frontend/src/routes/CatalogFormPage.tsx:95
- Why it matters: UI may navigate to a temporary item id that does not match server-created id after replay.
- Minimum actionable fix: persist client->server id mapping after replay and redirect/update caches accordingly.

3. Severity: Medium
- Title: Regression tests do not cover new export failure-path behavior
- Evidence: repo/backend/unit_tests/reports/report_service_test.go:171, repo/backend/unit_tests/reports/report_service_test.go:194
- Why it matters: a critical integrity path is now implemented but unguarded against future regressions.
- Minimum actionable fix: add tests that force queryReportData failure and assert ExportStatusFailed + error return.

4. Severity: Medium
- Title: No frontend unit tests for offline mutation queue and replay
- Evidence: repo/frontend/unit_tests/lib/offline-cache.test.ts:1 (query cache only), no matches for queue/replay helpers in frontend unit tests
- Why it matters: new offline queue behavior is high-impact and currently unverified in CI.
- Minimum actionable fix: add tests for enqueue/load/remove/replay success, replay transient offline stop, and replay terminal-error retention behavior.

### 0.4 Static Validation Notes

- Updated backend unit coverage includes new assertions for pricing and actor provenance:
  - repo/backend/unit_tests/orders/order_service_test.go:220
  - repo/backend/unit_tests/orders/order_service_test.go:255
  - repo/backend/unit_tests/orders/order_service_test.go:317
  - repo/backend/unit_tests/campaign/campaign_service_test.go:376

- Important boundary remains: this is a static re-review; no runtime confirmation is claimed.

---

## 1. Original Audit Snapshot (Superseded By Section 0)

- Overall conclusion at original audit time: Fail

## 2. Scope and Static Verification Boundary

- What was reviewed:
  - Prompt and constraints: prompt.md:1, prompt.md:3
  - Documentation and run/test guidance: repo/README.md:3, repo/README.md:182, repo/README.md:257, repo/run_tests.sh:27, repo/run_tests.sh:29, repo/run_tests.sh:45
  - Backend architecture, routes, services, security, jobs, seeds/migrations: repo/backend/internal/bootstrap/app.go:242, repo/backend/internal/http/router.go:128, repo/backend/internal/http/router.go:138, repo/backend/internal/application/order_service.go:63, repo/backend/internal/application/campaign_service.go:118, repo/backend/internal/application/report_service.go:298, repo/backend/internal/application/backup_service.go:84
  - Frontend offline behavior and reporting/catalog flows: repo/frontend/src/lib/offline.tsx:11, repo/frontend/src/lib/offline-cache.ts:48, repo/frontend/src/routes/CatalogFormPage.tsx:78, repo/frontend/src/routes/CatalogFormPage.tsx:137, repo/frontend/src/routes/ReportsPage.tsx:203
  - Static tests (unit + API + frontend): repo/backend/api_tests/lockout_test.go:58, repo/backend/api_tests/rbac_test.go:15, repo/backend/api_tests/member_coach_location_test.go:9, repo/backend/api_tests/report_export_test.go:22, repo/backend/unit_tests/orders/order_service_test.go:220, repo/backend/unit_tests/reports/report_service_test.go:158, repo/frontend/unit_tests/lib/offline-cache.test.ts:1
- What was not reviewed:
  - Runtime behavior under real network/browser/container execution
  - Real DB load/performance characteristics
  - Real TLS trust and certificate UX on clients
- What was intentionally not executed:
  - Project startup, Docker, tests, external services
- Claims requiring manual verification:
  - End-to-end offline reconciliation conflict behavior
  - Real backup scheduling and restore integrity across actual filesystem constraints
  - Browser-level UX/a11y and production deployment hardening

## 3. Repository and Requirement Mapping Summary

- Prompt core goal:
  - Offline-first operations and inventory suite for fitness-club group-buy workflows, with strict role controls, operational timelines, procurement loop, reporting/export, security/privacy, retention, and encrypted backups: prompt.md:1, prompt.md:3
- Main implementation areas mapped:
  - Backend API routing and RBAC middleware: repo/backend/internal/http/router.go:128, repo/backend/internal/http/router.go:138
  - Core domain/application services (orders, campaigns, procurement, reports, retention, backups): repo/backend/internal/application/order_service.go:63, repo/backend/internal/application/campaign_service.go:118, repo/backend/internal/application/report_service.go:298, repo/backend/internal/application/backup_service.go:84
  - Frontend role-gated routes with offline cache and mutation restrictions: repo/frontend/src/lib/offline-cache.ts:48, repo/frontend/src/routes/CatalogFormPage.tsx:78, repo/frontend/src/routes/ReportsPage.tsx:203
  - Static test suites and entrypoints: repo/run_tests.sh:27, repo/run_tests.sh:45

## 4. Section-by-section Review

### 4.1 Hard Gates

#### 4.1.1 Documentation and static verifiability

- Conclusion: Pass
- Rationale: Startup, environment, and test guidance are present and statically coherent with repository structure.
- Evidence: repo/README.md:182, repo/README.md:216, repo/README.md:217, repo/README.md:257, repo/run_tests.sh:27, repo/run_tests.sh:45, repo/backend/internal/bootstrap/app.go:242, repo/backend/internal/bootstrap/app.go:249
- Manual verification note: Runtime operability is not confirmed in this static audit.

#### 4.1.2 Material deviation from Prompt

- Conclusion: Fail
- Rationale: Core business semantics are materially weakened in multiple places: offline-first scope is explicitly reduced to online-only for key operations, order pricing uses deposit rather than sale/rental unit price, and operation logs misattribute actors.
- Evidence: repo/frontend/src/lib/offline.tsx:11, repo/frontend/src/routes/CatalogFormPage.tsx:137, repo/frontend/src/routes/ReportsPage.tsx:203, repo/backend/internal/application/order_service.go:63, repo/backend/internal/application/campaign_service.go:118, repo/backend/internal/application/order_service.go:164

### 4.2 Delivery Completeness

#### 4.2.1 Core explicit requirements coverage

- Conclusion: Partial Pass
- Rationale: Many core requirements are implemented (auth lockout/CAPTCHA, KPI/report/export, procurement lifecycle, backup encryption/checksum). However, critical requirement fidelity gaps remain in pricing semantics, complete timeline provenance, and offline-first behavior.
- Evidence: repo/backend/internal/application/auth_service.go:118, repo/backend/internal/security/password.go:29, repo/backend/internal/domain/report.go:35, repo/backend/internal/application/backup_service.go:109, repo/backend/internal/application/order_service.go:63, repo/backend/internal/application/report_service.go:298

#### 4.2.2 End-to-end 0 to 1 deliverable status

- Conclusion: Partial Pass
- Rationale: Repository is full-stack and structurally complete with docs/tests. Static quality risks still allow severe business defects despite buildable structure.
- Evidence: repo/README.md:3, repo/backend/internal/bootstrap/app.go:242, repo/frontend/package.json:10, repo/backend/api_tests/report_export_test.go:22

### 4.3 Engineering and Architecture Quality

#### 4.3.1 Structure and decomposition

- Conclusion: Pass
- Rationale: Layering is reasonable (bootstrap, http, application, domain, store, frontend routes/hooks).
- Evidence: repo/backend/internal/bootstrap/app.go:58, repo/backend/internal/http/router.go:34, repo/backend/internal/application/order_service.go:15, repo/frontend/src/app/providers.tsx:1

#### 4.3.2 Maintainability and extensibility

- Conclusion: Partial Pass
- Rationale: Structure is extensible, but key reliability anti-patterns reduce maintainability (silent error swallowing in export generation; actor provenance inconsistencies across operations).
- Evidence: repo/backend/internal/application/report_service.go:298, repo/backend/internal/application/order_service.go:164, repo/backend/internal/application/order_service.go:328

### 4.4 Engineering Details and Professionalism

#### 4.4.1 Error handling, logging, validation, API design

- Conclusion: Partial Pass
- Rationale: Validation and standardized envelopes are mostly good, but export pipeline suppresses query errors and returns completed jobs, which is non-professional for audit/reporting workflows.
- Evidence: repo/backend/internal/http/errors.go:65, repo/backend/internal/application/report_service.go:298, repo/backend/internal/application/report_service.go:170

#### 4.4.2 Product-like service vs demo shape

- Conclusion: Partial Pass
- Rationale: Product-like breadth exists, but critical behaviors still read as rollout placeholders (online-only notices for offline-first scope).
- Evidence: repo/frontend/src/routes/CatalogFormPage.tsx:137, repo/frontend/src/routes/ReportsPage.tsx:203

### 4.5 Prompt Understanding and Requirement Fit

#### 4.5.1 Business objective and constraints fit

- Conclusion: Fail
- Rationale: The implementation departs from core prompt semantics in high-impact areas: pricing model semantics, timeline actor traceability, and offline-first operational expectation.
- Evidence: prompt.md:1, prompt.md:3, repo/backend/internal/application/order_service.go:63, repo/backend/internal/application/campaign_service.go:118, repo/backend/internal/application/order_service.go:164, repo/frontend/src/lib/offline.tsx:11

### 4.6 Aesthetics (frontend-only/full-stack)

#### 4.6.1 Visual/interaction quality

- Conclusion: Cannot Confirm Statistically
- Rationale: Static code review cannot validate rendered spacing, typography consistency, hover/focus behavior, or visual coherence.
- Evidence: repo/frontend/src/routes/DashboardPage.tsx:1
- Manual verification note: Browser-based UI inspection required.

## 5. Issues and Suggestions (Severity-Rated)

### Blocker and High

1. Severity: High
- Title: Order pricing uses refundable deposit as transactional unit price
- Conclusion: Fail
- Evidence: repo/backend/internal/application/order_service.go:63, repo/backend/internal/application/campaign_service.go:118, prompt.md:1
- Impact: Revenue/procurement accounting and customer pricing semantics can be materially wrong; billing-model intent is undermined.
- Minimum actionable fix: Use item unit_price (or billing-model-specific price field) as order unit price, keep refundable_deposit as separate liability field; update DTOs and accounting/reporting rollups accordingly.

2. Severity: High
- Title: Operation timeline and audit actor attribution is incorrect in key flows
- Conclusion: Fail
- Evidence: repo/backend/internal/http/order_handler.go:123, repo/backend/internal/application/order_service.go:159, repo/backend/internal/application/order_service.go:164, repo/backend/internal/application/order_service.go:328, repo/backend/internal/application/order_service.go:348
- Impact: Tamper-evident audit purpose is weakened because actions by managers/system can be recorded as if done by order owners.
- Minimum actionable fix: Thread performedBy actor through Pay/Split/Merge/AutoClose paths and write the actual actor to timeline + audit events.

3. Severity: High
- Title: Campaign-created orders miss creation timeline provenance
- Conclusion: Fail
- Evidence: repo/backend/internal/application/campaign_service.go:124, prompt.md:1
- Impact: Staff cannot rely on a complete visible operation timeline for campaign-origin orders; explainability is degraded.
- Minimum actionable fix: Add timeline entry creation in campaign join flow (and include campaign linkage metadata).

4. Severity: High
- Title: Export generation silently ignores report query failures
- Conclusion: Fail
- Evidence: repo/backend/internal/application/report_service.go:298
- Impact: Exports can be marked completed with partial/empty content while underlying data retrieval fails, causing silent reporting integrity defects.
- Minimum actionable fix: Handle queryReportData error explicitly, fail export job status with reason, and return error envelope.

5. Severity: High
- Title: Offline-first requirement is materially narrowed by online-only mutation policy
- Conclusion: Fail
- Evidence: repo/frontend/src/lib/offline.tsx:11, repo/frontend/src/routes/CatalogFormPage.tsx:78, repo/frontend/src/routes/CatalogFormPage.tsx:137, repo/frontend/src/routes/ReportsPage.tsx:203, prompt.md:1
- Impact: Core management actions are unavailable offline, conflicting with prompt’s offline-first operational expectation.
- Minimum actionable fix: Implement offline mutation queue + reconciliation for critical flows, or formally narrow and justify requirement scope in prompt-aligned docs.

### Medium

6. Severity: Medium
- Title: Biometric security path is disabled by default and untested in API suite
- Conclusion: Partial Pass
- Evidence: repo/docker-compose.yml:36, repo/backend/internal/http/biometric_handler.go:30, repo/backend/api_tests/integration_helpers_test.go:146
- Impact: A core security/privacy path is not validated in default delivery posture; production readiness risk remains.
- Minimum actionable fix: Provide a tested enabled-biometric profile and add API tests for register/get/revoke/rotate/list flows.

7. Severity: Medium
- Title: Report service tests do not cover failure-path correctness for export data retrieval
- Conclusion: Partial Pass
- Evidence: repo/backend/unit_tests/reports/report_service_test.go:158, repo/backend/unit_tests/reports/report_service_test.go:169, repo/backend/internal/application/report_service.go:298
- Impact: Severe export integrity defects can pass static test gates.
- Minimum actionable fix: Add unit/integration tests that force query failures and assert failed export status + error propagation.

8. Severity: Medium
- Title: Order tests reinforce deposit-based pricing instead of catching semantic mismatch
- Conclusion: Partial Pass
- Evidence: repo/backend/unit_tests/orders/order_service_test.go:220, repo/backend/unit_tests/orders/order_service_test.go:231
- Impact: Existing tests can lock in incorrect pricing behavior.
- Minimum actionable fix: Add tests asserting unit_price semantics by billing model and explicit deposit separation.

## 6. Security Review Summary

- Authentication entry points:
  - Conclusion: Pass
  - Evidence: repo/backend/internal/http/auth_handler.go:25, repo/backend/internal/application/auth_service.go:118, repo/backend/internal/security/password.go:29
  - Reasoning: Lockout/CAPTCHA/session lifecycle and Argon2id hashing are implemented.

- Route-level authorization:
  - Conclusion: Partial Pass
  - Evidence: repo/backend/internal/http/router.go:128, repo/backend/internal/http/router.go:137, repo/backend/internal/http/router.go:138
  - Reasoning: Most endpoints are permission-gated; cancel route uses handler/service-level checks after auth middleware.

- Object-level authorization:
  - Conclusion: Partial Pass
  - Evidence: repo/backend/internal/application/order_service.go:95, repo/backend/internal/application/order_service.go:586, repo/backend/api_tests/order_test.go:50
  - Reasoning: Order ownership checks exist, but coverage is concentrated; broader object-level checks should be expanded.

- Function-level authorization:
  - Conclusion: Partial Pass
  - Evidence: repo/backend/internal/application/order_service.go:525, repo/backend/internal/application/order_service.go:164
  - Reasoning: Function guards exist for some operations; actor provenance bug weakens effective audit security controls.

- Tenant/user data isolation:
  - Conclusion: Partial Pass
  - Evidence: repo/backend/api_tests/member_coach_location_test.go:9, repo/backend/internal/security/rbac.go:63
  - Reasoning: Static evidence of location-scoped personnel behavior exists; broader multi-tenant isolation still needs runtime and wider test confirmation.

- Admin/internal/debug protection:
  - Conclusion: Pass
  - Evidence: repo/backend/internal/http/router.go:200, repo/backend/internal/http/router.go:209, repo/backend/api_tests/rbac_test.go:15
  - Reasoning: Admin routes are behind auth and role checks with explicit forbidden-path tests.

## 7. Tests and Logging Review

- Unit tests:
  - Conclusion: Partial Pass
  - Evidence: repo/backend/unit_tests/orders/order_service_test.go:220, repo/backend/unit_tests/reports/report_service_test.go:158, repo/frontend/unit_tests/lib/offline-cache.test.ts:1
  - Rationale: Broad presence, but important failure-path and semantic checks are missing.

- API and integration tests:
  - Conclusion: Partial Pass
  - Evidence: repo/backend/api_tests/lockout_test.go:58, repo/backend/api_tests/rbac_test.go:15, repo/backend/api_tests/report_export_test.go:22
  - Rationale: Good coverage for auth/RBAC/main happy paths; gaps remain for biometric enabled flows and audit actor attribution.

- Logging categories and observability:
  - Conclusion: Pass
  - Evidence: repo/backend/internal/platform/logger.go:1, repo/backend/internal/jobs/jobs.go:72, repo/backend/internal/jobs/procurement_jobs.go:39
  - Rationale: Structured logging and job-level diagnostics are present.

- Sensitive-data leakage risk in logs/responses:
  - Conclusion: Partial Pass
  - Evidence: repo/backend/internal/application/order_service.go:254, repo/backend/internal/security/audit_helper.go:58, repo/backend/internal/application/report_service.go:423
  - Rationale: Note redaction and safe-details handling exist; export masking is field-name based and should be expanded with explicit sensitive-field policy tests.

## 8. Test Coverage Assessment (Static Audit)

### 8.1 Test Overview

- Unit tests and API/integration tests exist:
  - Backend unit: repo/run_tests.sh:27, repo/backend/unit_tests/orders/order_service_test.go:220
  - Backend API/integration: repo/run_tests.sh:29, repo/backend/api_tests/lockout_test.go:58
  - Frontend unit: repo/run_tests.sh:45, repo/frontend/package.json:10
- Test frameworks and entry points:
  - Go test: repo/run_tests.sh:27, repo/run_tests.sh:29
  - Vitest: repo/run_tests.sh:45, repo/frontend/package.json:10
- Documentation provides test commands:
  - repo/README.md:257

### 8.2 Coverage Mapping Table

| Requirement or Risk Point | Mapped Test Case(s) | Key Assertion / Fixture / Mock | Coverage Assessment | Gap | Minimum Test Addition |
|---|---|---|---|---|---|
| 5-failure lockout + CAPTCHA flow | repo/backend/api_tests/lockout_test.go:58, repo/backend/api_tests/lockout_test.go:79 | Expects 423 ACCOUNT_LOCKED and captcha verify behavior | sufficient | None material in static scope | Add brute-force/backoff checks for repeated wrong CAPTCHA |
| 401 and 403 route guards | repo/backend/api_tests/rbac_test.go:7, repo/backend/api_tests/rbac_test.go:15 | Unauthenticated 401 and member forbidden on admin users route | sufficient | Coverage concentrated on selected admin route | Add representative 401/403 tests for procurement/report/export/admin subsets |
| Object-level order isolation | repo/backend/api_tests/order_test.go:50 | Member cannot GET other member order (403) | basically covered | Split/merge/pay ownership and cross-user edge-cases not covered | Add object-level tests for split/merge/cancel/pay permutations |
| Personnel location isolation | repo/backend/api_tests/member_coach_location_test.go:9 | Ops manager forced to own location records | basically covered | Isolation coverage not generalized across all location-bound entities | Add location isolation tests for inventory/procurement/report filters |
| Item publish validation + batch partial failures | repo/backend/api_tests/item_test.go:72, repo/backend/api_tests/item_test.go:140 | Expects PUBLISH_BLOCKED and 207 partial failure counts | sufficient | Inline feedback UX only statically inferred | Add frontend assertions for row-level validation reason rendering |
| Procurement closed-loop + variance resolution | repo/backend/api_tests/procurement_test.go:22 | Create/approve/receive/resolve and landed cost retrieval | basically covered | Missing explicit 5-business-day deadline enforcement tests at API boundary | Add tests asserting overdue escalation and resolve constraints by due date |
| Export ACL and download path | repo/backend/api_tests/report_export_test.go:22 | Download returns attachment; coach forbidden for admin export access | basically covered | No failure-path coverage for data query errors during export | Add export tests with forced DB query failure expecting failed job/status |
| Offline-first behavior for reads | repo/frontend/unit_tests/lib/offline-cache.test.ts:1 | Cacheable roots include dashboard/items/orders/admin reads | basically covered | Mutation offline queue/replay not implemented | Add tests for queued offline writes and replay conflict policy |
| Pricing semantics (unit price vs deposit) | repo/backend/unit_tests/orders/order_service_test.go:220 | Tests currently validate deposit-driven price path | insufficient | Tests reinforce wrong semantic behavior | Add tests asserting order.unit_price equals item.unit_price by billing model |
| Timeline/audit actor provenance | repo/backend/api_tests/order_test.go:104 | Asserts timeline count and redacted note text only | insufficient | No performed_by actor correctness checks | Add assertions that manager/system actor IDs are recorded on pay/split/merge |

### 8.3 Security Coverage Audit

- Authentication:
  - Conclusion: sufficient static coverage
  - Evidence: repo/backend/api_tests/lockout_test.go:58, repo/backend/api_tests/auth_test.go:7
  - Residual risk: CAPTCHA abuse-rate limiting not explicitly tested.

- Route authorization:
  - Conclusion: basically covered
  - Evidence: repo/backend/api_tests/rbac_test.go:15, repo/backend/api_tests/dashboard_test.go:9
  - Residual risk: Not all route groups have explicit deny-path tests.

- Object-level authorization:
  - Conclusion: insufficient
  - Evidence: repo/backend/api_tests/order_test.go:50
  - Residual risk: Severe cross-user action defects in split/merge/pay could remain undetected.

- Tenant/data isolation:
  - Conclusion: basically covered
  - Evidence: repo/backend/api_tests/member_coach_location_test.go:9
  - Residual risk: Incomplete coverage outside personnel/location domain.

- Admin/internal protection:
  - Conclusion: basically covered
  - Evidence: repo/backend/api_tests/rbac_test.go:15, repo/backend/api_tests/admin_test.go:7
  - Residual risk: Biometric-enabled path not covered because test config disables it.

### 8.4 Final Coverage Judgment

- Conclusion: Partial Pass
- Boundary explanation:
  - Major risks covered: auth lockout/CAPTCHA, representative RBAC, key procurement/report happy paths.
  - Major uncovered risks: pricing semantics correctness, timeline/audit actor provenance, export failure integrity, biometric-enabled security path.
  - Result: existing suites can pass while severe business/audit defects remain.

## 9. Final Notes

- This report is static-only and evidence-based; runtime success was not inferred.
- Root-cause findings were prioritized over repeated symptoms.
- Manual verification remains required for runtime/offline replay and production deployment hardening.