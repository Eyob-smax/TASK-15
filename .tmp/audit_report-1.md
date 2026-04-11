# FitCommerce Static Delivery Acceptance & Architecture Audit

Date: 2026-04-11  
Mode: Static-only (no runtime execution, no Docker, no tests run)

## 0. Re-review Update (Authoritative)

This update supersedes the initial conclusions below and reflects static re-review after user-applied fixes.

- Updated conclusion: **Partial Pass**
- Acceptance status: **Not full pass yet** (one medium requirement-fit gap remains)

### Closed findings from initial report

1. CAPTCHA plaintext storage risk: **Closed**

- Evidence: `repo/backend/database/migrations/00010_captcha_answer_hashing.sql:5`, `repo/backend/database/migrations/00010_captcha_answer_hashing.sql:42`, `repo/backend/internal/store/postgres/captcha_store.go:30`

2. CAPTCHA RNG/verification hardening: **Closed**

- Evidence: `repo/backend/internal/security/captcha.go:4`, `repo/backend/internal/security/captcha.go:17`, `repo/backend/internal/security/captcha.go:57`, `repo/backend/internal/application/auth_service.go:82`, `repo/backend/internal/application/auth_service.go:253`

3. Auth API spec/implementation drift: **Closed**

- Evidence: `docs/api-spec.md:167`, `docs/api-spec.md:173`, `docs/api-spec.md:194`, `docs/api-spec.md:217`, `repo/backend/internal/http/auth_handler.go:28`, `repo/backend/internal/http/auth_handler.go:78`, `repo/backend/internal/http/errors.go:56`

### Remaining open finding

1. Offline-first cache coverage is partial for admin read domains: **Medium**

- Evidence:
  - Offline cache allowlist excludes admin roots such as `users`, `audit`, `backups`, `biometrics`, `encryption-keys`, `retention-policies`: `repo/frontend/src/lib/offline-cache.ts:15`
  - Admin hooks use those roots and therefore are not persisted for offline read continuity: `repo/frontend/src/lib/hooks/useAdmin.ts:88`, `repo/frontend/src/lib/hooks/useAdmin.ts:150`, `repo/frontend/src/lib/hooks/useAdmin.ts:175`, `repo/frontend/src/lib/hooks/useAdmin.ts:199`, `repo/frontend/src/lib/hooks/useAdmin.ts:248`, `repo/frontend/src/lib/hooks/useAdmin.ts:260`
- Impact: offline-first behavior exists for many operational reads, but not consistently across admin read surfaces.

---

## 1. Verdict

- Overall conclusion: **Fail (historical initial pass; superseded by Section 0)**

Rationale:

- A core prompt requirement is **offline-first management**; static evidence for offline-first behavior in the web client is insufficient and current implementation appears online-request-first.
- Security/privacy has a material weakness: CAPTCHA answers are stored in plaintext.
- Documentation-to-implementation contract inconsistencies are material for acceptance hard gate 1.1 (static verifiability).

---

## 2. Scope and Static Verification Boundary

- What was reviewed:
  - Core docs and contracts: `repo/README.md:182`, `docs/api-spec.md:145`, `prompt.md:1`
  - Backend routing/auth/security/services/jobs/migrations/stores/tests
  - Frontend routing/pages/hooks/api client/tests/config
- What was not reviewed:
  - Runtime behavior in browser/API/network/container
  - Actual DB migration execution against a live PostgreSQL instance
  - Real file generation/download integrity in runtime environment
- What was intentionally not executed:
  - Project start, Docker, tests, external services
- Claims requiring manual verification:
  - End-to-end offline behavior under network interruption
  - Real export/download behavior in deployed environment
  - TLS cert deployment details in target environment
  - Backup restore operational procedure validity against real archives

---

## 3. Repository / Requirement Mapping Summary

- Prompt core business goal (mapped):
  - Fitness-club operations and inventory suite with group-buy campaigns, strict ops controls, role-based permissions, reporting/exports, procurement loop, auditability, security/privacy, and offline-first operation.
- Core flows mapped statically:
  - Auth/session/lockout/CAPTCHA (`repo/backend/internal/application/auth_service.go:55`)
  - RBAC and route protection (`repo/backend/internal/http/router.go:71`, `repo/backend/internal/security/rbac.go:41`)
  - Catalog publish validation and batch edits (`repo/backend/internal/application/item_service.go:184`, `repo/backend/internal/domain/policies.go:9`)
  - Campaign/order/procurement state logic (`repo/backend/internal/application/campaign_service.go:145`, `repo/backend/internal/application/order_service.go:496`, `repo/backend/internal/application/procurement_service.go:141`)
  - Reporting/export and masking (`repo/backend/internal/application/report_service.go:245`, `repo/backend/internal/application/report_service.go:437`)
  - Backups/retention/biometric key rotation jobs (`repo/backend/internal/jobs/procurement_jobs.go:8`, `repo/backend/internal/jobs/procurement_jobs.go:114`)

---

## 4. Section-by-section Review

### 4.1 Hard Gates

#### 4.1.1 Documentation and static verifiability

- Conclusion: **Fail**
- Rationale:
  - Startup/test/config docs exist and are extensive, but there are material API contract mismatches between documentation and implementation that reduce static verifiability reliability.
- Evidence:
  - Docs present: `repo/README.md:182`, `repo/README.md:253`, `repo/README.md:203`
  - Doc says login response contains `session.token`: `docs/api-spec.md:160`; implementation intentionally does not include token in body: `repo/backend/internal/http/auth_handler.go:67`
  - Doc says logout success is 204: `docs/api-spec.md:181`; implementation returns 200 envelope: `repo/backend/internal/http/auth_handler.go:78`
  - Doc lists captcha verify `NOT_FOUND`: `docs/api-spec.md:211`; service maps missing challenge to unauthorized: `repo/backend/internal/application/auth_service.go:233`
- Manual verification note:
  - Manual contract reconciliation needed before consumer conformance checks.

#### 4.1.2 Material deviation from Prompt

- Conclusion: **Fail**
- Rationale:
  - Prompt requires offline-first management; frontend static evidence indicates online request/response dependency without offline persistence/sync mechanisms.
- Evidence:
  - Client request model is direct fetch with credentials: `repo/frontend/src/lib/api-client.ts:75`, `repo/frontend/src/lib/api-client.ts:92`
  - No service worker registration in app bootstrap: `repo/frontend/src/main.tsx:1`
  - Query behavior is reconnect/refetch oriented: `repo/frontend/src/app/providers.tsx:16`
- Manual verification note:
  - Manual offline mode test required (LAN disconnect / backend unavailable) to confirm actual behavior.

### 4.2 Delivery Completeness

#### 4.2.1 Core explicit requirements coverage

- Conclusion: **Partial Pass**
- Rationale:
  - Most core domains are implemented (catalog/inventory/campaigns/orders/procurement/reports/audit/security jobs).
  - Offline-first requirement remains materially unproven/likely unmet.
- Evidence:
  - Domain coverage in routes and services: `repo/backend/internal/http/router.go:39`, `repo/backend/internal/bootstrap/app.go:98`
  - Prompt-required business rules examples implemented: `repo/backend/internal/domain/policies.go:9`, `repo/backend/internal/application/order_service.go:496`, `repo/backend/internal/application/procurement_service.go:177`
  - Offline-first gap evidence: `repo/frontend/src/main.tsx:1`, `repo/frontend/src/lib/api-client.ts:75`

#### 4.2.2 0-to-1 end-to-end deliverable vs partial/demo

- Conclusion: **Pass**
- Rationale:
  - Full project structure exists with backend/frontend, migrations, tests, Docker artifacts, and documentation.
- Evidence:
  - Structure/docs: `repo/README.md:1`, `repo/run_tests.sh:1`, `repo/docker-compose.yml:1`
  - Backend+frontend tests exist: `repo/backend/api_tests/auth_test.go:1`, `repo/frontend/unit_tests/setup.ts:1`

### 4.3 Engineering and Architecture Quality

#### 4.3.1 Engineering structure and module decomposition

- Conclusion: **Pass**
- Rationale:
  - Clear layered decomposition (domain/application/http/store/platform/jobs) and substantial route/service segregation.
- Evidence:
  - Wiring and composition root: `repo/backend/internal/bootstrap/app.go:42`
  - Route decomposition: `repo/backend/internal/http/router.go:39`

#### 4.3.2 Maintainability/extensibility

- Conclusion: **Partial Pass**
- Rationale:
  - Overall maintainable architecture; however, some critical security implementation details reduce production robustness.
- Evidence:
  - Positive: interfaces/services/repositories split: `repo/backend/internal/application/services.go:12`, `repo/backend/internal/store/repositories.go:1`
  - Concern: plaintext CAPTCHA answer persistence: `repo/backend/database/migrations/00002_users_and_auth.sql:40`, `repo/backend/internal/store/postgres/captcha_store.go:38`

### 4.4 Engineering Details and Professionalism

#### 4.4.1 Error handling, logging, validation, API design

- Conclusion: **Partial Pass**
- Rationale:
  - Strong structured error envelope and logging exist; major docs-vs-code contract drift and CAPTCHA handling weakness reduce professionalism score.
- Evidence:
  - Error mapping envelope: `repo/backend/internal/http/errors.go:39`
  - Structured logs with masking intent: `repo/backend/internal/platform/logger.go:14`, `repo/backend/internal/platform/logger.go:33`
  - DTO validation in handlers: `repo/backend/internal/http/campaign_handler.go:37`, `repo/backend/internal/http/order_handler.go:32`
  - Contract drift: `docs/api-spec.md:181` vs `repo/backend/internal/http/auth_handler.go:78`

#### 4.4.2 Real product/service organization vs demo

- Conclusion: **Pass**
- Rationale:
  - The repository is structured and implemented as a real service stack, not a toy sample.
- Evidence:
  - Multi-domain migrations: `repo/backend/database/migrations/00001_enums_and_base.sql:1`, `repo/backend/database/migrations/00008_variance_resolution_and_retention_cleanup.sql:1`
  - Production-style app composition and jobs: `repo/backend/cmd/api/main.go:56`

### 4.5 Prompt Understanding and Requirement Fit

#### 4.5.1 Business goal and implicit constraints fit

- Conclusion: **Partial Pass**
- Rationale:
  - Most business and operational constraints are represented in code.
  - Offline-first is the major fit deficit.
- Evidence:
  - Security/session/lockout parameters: `repo/backend/internal/platform/config.go:35`
  - Report permission and export naming: `repo/backend/internal/http/report_handler.go:54`, `repo/backend/internal/domain/report.go:36`
  - Offline-first gap: `repo/frontend/src/main.tsx:1`, `repo/frontend/src/lib/api-client.ts:75`

### 4.6 Aesthetics (frontend-only/full-stack)

- Conclusion: **Cannot Confirm Statistically**
- Rationale:
  - Visual rendering quality, interaction feel, and responsive behavior cannot be fully validated without runtime UI execution.
- Evidence:
  - Route/page structure exists: `repo/frontend/src/app/routes.tsx:79`
- Manual verification note:
  - Manual browser review required for visual hierarchy, interaction states, and responsive quality.

---

## 5. Issues / Suggestions (Severity-Rated)

### 5.1 High

#### Issue 1

- Severity: **High**
- Title: Offline-first requirement is not statically evidenced and appears unimplemented in web client behavior
- Conclusion: **Fail**
- Evidence:
  - `repo/frontend/src/main.tsx:1`
  - `repo/frontend/src/lib/api-client.ts:75`
  - `repo/frontend/src/app/providers.tsx:16`
- Impact:
  - Core prompt objective can be missed; severe UX/operational degradation when connectivity to backend drops.
- Minimum actionable fix:
  - Add explicit offline-first client layer: local persistence (e.g., IndexedDB), cached reads, queued mutations with replay/conflict handling, and deterministic UI state for offline/online transitions.

#### Issue 2

- Severity: **High**
- Title: CAPTCHA answers are stored in plaintext
- Conclusion: **Fail**
- Evidence:
  - `repo/backend/database/migrations/00002_users_and_auth.sql:40`
  - `repo/backend/internal/store/postgres/captcha_store.go:38`
  - `repo/backend/internal/application/auth_service.go:87`
- Impact:
  - DB exposure can permit challenge-answer disclosure and lockout bypass facilitation.
- Minimum actionable fix:
  - Store only a hashed CAPTCHA answer (or HMAC) with TTL; verify using constant-time comparison against derived value, never persist plaintext answer.

#### Issue 3

- Severity: **High**
- Title: API documentation materially diverges from implementation contracts
- Conclusion: **Fail**
- Evidence:
  - Login token documented in body: `docs/api-spec.md:160`; implementation omits token body and uses cookie: `repo/backend/internal/http/auth_handler.go:67`
  - Logout documented 204: `docs/api-spec.md:181`; implementation returns 200 body: `repo/backend/internal/http/auth_handler.go:78`
  - CAPTCHA verify documented NOT_FOUND: `docs/api-spec.md:211`; missing challenge returns unauthorized path: `repo/backend/internal/application/auth_service.go:233`
- Impact:
  - Client implementations and acceptance verification can fail despite code functioning as implemented.
- Minimum actionable fix:
  - Reconcile docs and handlers to a single source-of-truth contract; enforce via contract tests generated from spec.

### 5.2 Medium

#### Issue 4

- Severity: **Medium**
- Title: CAPTCHA challenge generation uses non-cryptographic RNG
- Conclusion: **Partial Pass**
- Evidence:
  - `repo/backend/internal/security/captcha.go:5`
  - `repo/backend/internal/security/captcha.go:13`
- Impact:
  - Predictability risk in challenge generation model (defense-in-depth reduction).
- Minimum actionable fix:
  - Replace `math/rand` with cryptographically secure randomness from `crypto/rand`.

#### Issue 5

- Severity: **Medium**
- Title: Some object-level authorization relies mainly on handler checks; service-layer consistency is not universal
- Conclusion: **Partial Pass**
- Evidence:
  - Handler-level ownership enforcement present for orders: `repo/backend/internal/http/order_handler.go:102`, `repo/backend/internal/http/order_handler.go:224`
  - Location scoping in handlers for members/coaches: `repo/backend/internal/http/member_handler.go:71`, `repo/backend/internal/http/coach_handler.go:73`
- Impact:
  - Future route additions/refactors may bypass object constraints if not consistently enforced deeper.
- Minimum actionable fix:
  - Add explicit service-layer object authorization policies for cross-cutting sensitive resources and enforce in all entry paths.

### 5.3 Low

#### Issue 6

- Severity: **Low**
- Title: API spec includes response examples that can be misunderstood as authoritative over stronger implementation security choices
- Conclusion: **Partial Pass**
- Evidence:
  - Token in login response example: `docs/api-spec.md:160`
  - Cookie-only operational behavior in handler: `repo/backend/internal/http/auth_handler.go:42`
- Impact:
  - Integration confusion, slower onboarding.
- Minimum actionable fix:
  - Clarify in spec that session token is cookie-only and intentionally omitted from response payload.

---

## 6. Security Review Summary

### authentication entry points

- Conclusion: **Partial Pass**
- Evidence:
  - Auth endpoints and session validation: `repo/backend/internal/http/router.go:71`, `repo/backend/internal/http/middleware.go:20`
  - Argon2id password hashing: `repo/backend/internal/security/password.go:20`
  - Lockout/session timeout enforcement: `repo/backend/internal/application/auth_service.go:110`, `repo/backend/internal/application/auth_service.go:146`
  - Weakness: plaintext CAPTCHA answer: `repo/backend/database/migrations/00002_users_and_auth.sql:40`

### route-level authorization

- Conclusion: **Pass**
- Evidence:
  - Route groups protected by auth and role middleware: `repo/backend/internal/http/router.go:79`, `repo/backend/internal/http/router.go:200`
  - Role permission matrix: `repo/backend/internal/security/rbac.go:41`

### object-level authorization

- Conclusion: **Partial Pass**
- Evidence:
  - Orders owner checks: `repo/backend/internal/http/order_handler.go:102`, `repo/backend/internal/http/order_handler.go:160`
  - Member/coach location-based object checks: `repo/backend/internal/http/member_handler.go:106`, `repo/backend/internal/http/coach_handler.go:108`

### function-level authorization

- Conclusion: **Partial Pass**
- Evidence:
  - Service-level report role check as second layer: `repo/backend/internal/application/report_service.go:66`, `repo/backend/internal/application/report_service.go:254`
  - Not all domain services show explicit actor-based policy checks.

### tenant / user data isolation

- Conclusion: **Partial Pass**
- Evidence:
  - Per-user order visibility controls: `repo/backend/internal/http/order_handler.go:67`
  - Location scoping for personnel directories: `repo/backend/internal/http/member_handler.go:71`, `repo/backend/internal/http/coach_handler.go:73`
- Note:
  - Full multi-tenant isolation model is not explicit in architecture; manual threat modeling recommended.

### admin / internal / debug protection

- Conclusion: **Pass**
- Evidence:
  - Admin routes are auth-protected and role-gated: `repo/backend/internal/http/router.go:206`, `repo/backend/internal/http/router.go:214`, `repo/backend/internal/http/router.go:219`

---

## 7. Tests and Logging Review

### Unit tests

- Conclusion: **Pass**
- Evidence:
  - Domain/policy/state/security/service tests are present: `repo/backend/unit_tests/domain/policies_test.go:36`, `repo/backend/unit_tests/security/rbac_test.go:19`, `repo/backend/unit_tests/orders/order_service_test.go:220`

### API / integration tests

- Conclusion: **Partial Pass**
- Evidence:
  - Auth/lockout/order/rbac/report/procurement tests exist: `repo/backend/api_tests/auth_test.go:8`, `repo/backend/api_tests/lockout_test.go:10`, `repo/backend/api_tests/order_test.go:9`, `repo/backend/api_tests/rbac_test.go:8`, `repo/backend/api_tests/report_export_test.go:9`
- Gap:
  - No static evidence of tests validating offline-first client behavior.

### Logging categories / observability

- Conclusion: **Pass**
- Evidence:
  - Structured JSON logger and request middleware: `repo/backend/internal/platform/logger.go:14`, `repo/backend/internal/platform/logger.go:77`
  - Audit event chaining and security event views: `repo/backend/internal/application/audit_service.go:24`, `repo/backend/internal/http/admin_handler.go:53`

### Sensitive-data leakage risk in logs / responses

- Conclusion: **Partial Pass**
- Evidence:
  - Email masking in login logging: `repo/backend/internal/platform/logger.go:33`, `repo/backend/internal/platform/logger.go:108`
  - Sensitive key stripping in audit helper: `repo/backend/internal/security/audit_helper.go:74`
  - Counterexample risk: plaintext CAPTCHA answer at rest: `repo/backend/database/migrations/00002_users_and_auth.sql:40`

---

## 8. Test Coverage Assessment (Static Audit)

### 8.1 Test Overview

- Unit tests and API/integration tests exist: **Yes**
- Test frameworks:
  - Backend: Go test (`repo/run_tests.sh:22`)
  - Frontend: Vitest (`repo/run_tests.sh:41`, `repo/frontend/package.json:10`)
- Test entry points:
  - `repo/run_tests.sh:1`
  - backend suites: `repo/run_tests.sh:22`
  - frontend suite: `repo/run_tests.sh:41`
- Documentation provides test commands:
  - `repo/README.md:253`, `repo/README.md:261`

### 8.2 Coverage Mapping Table

| Requirement / Risk Point             | Mapped Test Case(s)                                                                                                                      | Key Assertion / Fixture / Mock                                | Coverage Assessment | Gap                                              | Minimum Test Addition                                                         |
| ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------- | ------------------- | ------------------------------------------------ | ----------------------------------------------------------------------------- |
| Auth login/session/logout            | `repo/backend/api_tests/auth_test.go:8`                                                                                                  | Session cookie set + session invalid after logout             | sufficient          | None material                                    | Add contract assertion for exact status/body per spec once reconciled         |
| Lockout + CAPTCHA flow               | `repo/backend/api_tests/lockout_test.go:10`, `repo/backend/api_tests/lockout_test.go:30`                                                 | 5 failures -> locked; captcha verify flow                     | basically covered   | Does not assert secure storage of CAPTCHA answer | Add persistence-layer test to confirm hashed/derived CAPTCHA answers only     |
| Route authorization 401/403          | `repo/backend/api_tests/rbac_test.go:8`, `repo/backend/api_tests/dashboard_test.go:9`, `repo/backend/api_tests/procurement_test.go:10`   | Unauthorized/forbidden route checks by role                   | sufficient          | None major                                       | Extend matrix for all high-risk admin/export endpoints                        |
| Object-level order isolation         | `repo/backend/api_tests/order_test.go:9`                                                                                                 | Member can view own order, forbidden on others                | sufficient          | Merge/split owner-scope tests limited            | Add API tests for split/merge object authorization boundaries                 |
| Item publish blocking rules          | `repo/backend/unit_tests/catalog/item_service_test.go:233`, `repo/backend/unit_tests/domain/policies_test.go:178`                        | Publish blocked on missing fields/window overlap              | sufficient          | None major                                       | Add API-level publish blocked response-detail test                            |
| Campaign cutoff success/failure      | `repo/backend/unit_tests/campaign/campaign_service_test.go:360`, `repo/backend/unit_tests/campaign/campaign_service_test.go:403`         | Succeeded vs failed + auto-close behavior                     | sufficient          | Runtime scheduler integration not executed       | Add deterministic integration test for cutoff job trigger path                |
| Order 30-minute auto-close logic     | `repo/backend/unit_tests/orders/order_service_test.go:220`, `repo/backend/unit_tests/jobs/jobs_test.go:92`                               | AutoCloseAt behavior and job calls                            | basically covered   | Real DB + scheduler timing not statically proven | Add integration test with fixed clock + persisted expired orders              |
| Procurement receive/variance/resolve | `repo/backend/unit_tests/procurement/procurement_service_test.go:387`, `repo/backend/unit_tests/procurement/variance_service_test.go:94` | Quantity/price variance creation and resolution transitions   | sufficient          | Limited API-level negative-path matrix           | Add API tests for invalid resolution action and due-date escalation edges     |
| Report access/export permissions     | `repo/backend/api_tests/report_export_test.go:9`                                                                                         | Coach forbidden for admin-only report; export/download checks | basically covered   | Contract mismatch not tested                     | Add API contract tests for response codes/body schema against spec            |
| Offline-first behavior               | None found                                                                                                                               | N/A                                                           | missing             | Core prompt requirement not represented in tests | Add frontend tests for offline reads, queued writes, replay/conflict behavior |

### 8.3 Security Coverage Audit

- authentication: **Basically covered**
  - Evidence: `repo/backend/api_tests/auth_test.go:8`, `repo/backend/api_tests/lockout_test.go:10`
  - Gap: storage-strength of CAPTCHA data not validated.
- route authorization: **Covered**
  - Evidence: `repo/backend/api_tests/rbac_test.go:8`, `repo/backend/api_tests/procurement_test.go:10`
- object-level authorization: **Basically covered**
  - Evidence: `repo/backend/api_tests/order_test.go:9`, `repo/backend/api_tests/member_coach_location_test.go:28`
  - Gap: split/merge and some cross-resource object checks can still hide defects.
- tenant / data isolation: **Insufficient**
  - Evidence: partial location scoping tests in personnel area: `repo/backend/api_tests/member_coach_location_test.go:28`
  - Gap: no broad tenant-isolation test strategy across all domain entities.
- admin / internal protection: **Covered**
  - Evidence: `repo/backend/api_tests/rbac_test.go:15`, `repo/backend/api_tests/admin_test.go:8`

### 8.4 Final Coverage Judgment

- **Partial Pass**

Boundary explanation:

- Major auth/RBAC/domain-rule paths are covered by static tests.
- However, offline-first behavior (core prompt risk) is missing from tests, and some security-hardening risks (e.g., CAPTCHA answer storage model) could remain undetected while current tests still pass.

---

## 9. Final Notes

- This audit is static-only and evidence-based; no runtime claims are made.
- Section 0 is the authoritative re-review outcome for current state.
- Current status after fixes: major security and contract issues are closed; one medium offline-first coverage gap remains.
- Updated priority remediation order:
  1. Expand offline persistence allowlist and tests to include admin read domains (`users`, `audit`, `backups`, `biometrics`, `encryption-keys`, `retention-policies`).
  2. Add frontend static tests proving offline hydration/persistence behavior for those domains.
