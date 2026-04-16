# Test Coverage Audit

## Scope And Method
- Audit mode: static inspection only
- Execution: no tests/scripts/containers run
- Primary route source: repo/backend/internal/http/router.go (RegisterRoutes + register*Routes)
- Primary API test sources: repo/backend/api_tests/*.go
- Unit test sources: repo/backend/unit_tests/**/*.go

## Backend Endpoint Inventory
Resolved base prefix: `/api/v1`

1. POST /api/v1/auth/login
2. POST /api/v1/auth/logout
3. GET /api/v1/auth/session
4. POST /api/v1/auth/captcha/verify
5. GET /api/v1/dashboard/kpis
6. POST /api/v1/items
7. GET /api/v1/items
8. GET /api/v1/items/:id
9. PUT /api/v1/items/:id
10. POST /api/v1/items/:id/publish
11. POST /api/v1/items/:id/unpublish
12. POST /api/v1/items/batch-edit
13. GET /api/v1/inventory/snapshots
14. POST /api/v1/inventory/adjustments
15. GET /api/v1/inventory/adjustments
16. POST /api/v1/warehouse-bins
17. GET /api/v1/warehouse-bins
18. GET /api/v1/warehouse-bins/:id
19. POST /api/v1/campaigns
20. GET /api/v1/campaigns
21. GET /api/v1/campaigns/:id
22. POST /api/v1/campaigns/:id/join
23. POST /api/v1/campaigns/:id/cancel
24. POST /api/v1/campaigns/:id/evaluate
25. POST /api/v1/orders
26. GET /api/v1/orders
27. GET /api/v1/orders/:id
28. POST /api/v1/orders/:id/pay
29. POST /api/v1/orders/:id/cancel
30. POST /api/v1/orders/:id/refund
31. POST /api/v1/orders/:id/notes
32. GET /api/v1/orders/:id/timeline
33. POST /api/v1/orders/:id/split
34. POST /api/v1/orders/merge
35. POST /api/v1/suppliers
36. GET /api/v1/suppliers
37. GET /api/v1/suppliers/:id
38. PUT /api/v1/suppliers/:id
39. POST /api/v1/purchase-orders
40. GET /api/v1/purchase-orders
41. GET /api/v1/purchase-orders/:id
42. POST /api/v1/purchase-orders/:id/approve
43. POST /api/v1/purchase-orders/:id/receive
44. POST /api/v1/purchase-orders/:id/return
45. POST /api/v1/purchase-orders/:id/void
46. GET /api/v1/variances
47. GET /api/v1/variances/:id
48. POST /api/v1/variances/:id/resolve
49. GET /api/v1/procurement/landed-costs
50. GET /api/v1/procurement/landed-costs/:poId
51. GET /api/v1/reports
52. GET /api/v1/reports/:id/data
53. POST /api/v1/exports
54. GET /api/v1/exports/:id
55. GET /api/v1/exports/:id/download
56. GET /api/v1/admin/audit-log
57. GET /api/v1/admin/audit-log/security
58. POST /api/v1/admin/backups
59. GET /api/v1/admin/backups
60. GET /api/v1/admin/backups/:id/verify
61. POST /api/v1/admin/biometrics/rotate-key
62. GET /api/v1/admin/biometrics/keys
63. POST /api/v1/admin/biometrics
64. GET /api/v1/admin/biometrics/:user_id
65. POST /api/v1/admin/biometrics/:user_id/revoke
66. POST /api/v1/admin/users
67. GET /api/v1/admin/users
68. GET /api/v1/admin/users/:id
69. PUT /api/v1/admin/users/:id
70. POST /api/v1/admin/users/:id/deactivate
71. GET /api/v1/admin/retention-policies
72. GET /api/v1/admin/retention-policies/:entity_type
73. PUT /api/v1/admin/retention-policies/:entity_type
74. POST /api/v1/locations
75. GET /api/v1/locations
76. GET /api/v1/locations/:id
77. POST /api/v1/coaches
78. GET /api/v1/coaches
79. GET /api/v1/coaches/:id
80. POST /api/v1/members
81. GET /api/v1/members
82. GET /api/v1/members/:id

## API Test Mapping Table
Legend:
- `TNM HTTP` = true no-mock HTTP test
- `HTTP+Mock` = HTTP test with mocking/stubbing in execution path

| Endpoint | Covered | Type | Test Files | Evidence (function refs) |
|---|---|---|---|---|
| POST /api/v1/auth/login | yes | TNM HTTP | api_tests/auth_test.go, api_tests/lockout_test.go | TestAuth_LoginSessionLogoutFlow, TestAuth_LockoutAfterFiveFailures |
| POST /api/v1/auth/logout | yes | TNM HTTP | api_tests/auth_test.go | TestAuth_LoginSessionLogoutFlow |
| GET /api/v1/auth/session | yes | TNM HTTP | api_tests/auth_test.go, api_tests/http_branch_coverage_test.go | TestAuth_LoginSessionLogoutFlow, TestHTTP_ItemCampaignAuthAndLocationBranches |
| POST /api/v1/auth/captcha/verify | yes | TNM HTTP | api_tests/lockout_test.go | TestAuth_CaptchaFlowAfterLockoutExpiry |
| GET /api/v1/dashboard/kpis | yes | TNM HTTP | api_tests/dashboard_test.go | TestDashboard_AdminAndCoachCanReadKPIs |
| POST /api/v1/items | yes | TNM HTTP | api_tests/item_test.go, api_tests/http_branch_coverage_test.go | TestItems_CreateGetAndUpdateRespectVersioning |
| GET /api/v1/items | yes | TNM HTTP | api_tests/item_test.go | TestItems_PublishAndBatchEditUseRealValidation |
| GET /api/v1/items/:id | yes | TNM HTTP | api_tests/item_test.go, api_tests/http_branch_coverage_test.go | TestItems_CreateGetAndUpdateRespectVersioning |
| PUT /api/v1/items/:id | yes | TNM HTTP | api_tests/item_test.go, api_tests/http_branch_coverage_test.go | TestItems_CreateGetAndUpdateRespectVersioning |
| POST /api/v1/items/:id/publish | yes | TNM HTTP | api_tests/item_test.go | TestItems_PublishAndBatchEditUseRealValidation |
| POST /api/v1/items/:id/unpublish | yes | TNM HTTP | api_tests/item_test.go | TestItems_PublishAndBatchEditUseRealValidation |
| POST /api/v1/items/batch-edit | yes | TNM HTTP | api_tests/item_test.go | TestItems_PublishAndBatchEditUseRealValidation |
| GET /api/v1/inventory/snapshots | yes | TNM HTTP | api_tests/inventory_test.go, api_tests/http_branch_coverage_test.go | TestInventory_RoleAccessAndSnapshotListing |
| POST /api/v1/inventory/adjustments | yes | TNM HTTP | api_tests/inventory_test.go, api_tests/http_branch_coverage_test.go | TestInventory_CreateAdjustmentAndWarehouseBin |
| GET /api/v1/inventory/adjustments | yes | TNM HTTP | api_tests/inventory_test.go | TestInventory_CreateAdjustmentAndWarehouseBin |
| POST /api/v1/warehouse-bins | yes | TNM HTTP | api_tests/inventory_test.go | TestInventory_CreateAdjustmentAndWarehouseBin |
| GET /api/v1/warehouse-bins | yes | TNM HTTP | api_tests/inventory_test.go | TestInventory_CreateAdjustmentAndWarehouseBin |
| GET /api/v1/warehouse-bins/:id | yes | TNM HTTP | api_tests/inventory_test.go | TestInventory_CreateAdjustmentAndWarehouseBin |
| POST /api/v1/campaigns | yes | TNM HTTP | api_tests/campaign_test.go, api_tests/report_export_test.go, api_tests/http_branch_coverage_test.go | TestCampaigns_MemberCanStartCampaign |
| GET /api/v1/campaigns | yes | TNM HTTP | api_tests/campaign_test.go | TestCampaigns_ListCancelAndEvaluateEndpoints |
| GET /api/v1/campaigns/:id | yes | TNM HTTP | api_tests/campaign_test.go | TestCampaigns_OperationsManagerCanCreateAndMemberCanJoin |
| POST /api/v1/campaigns/:id/join | yes | TNM HTTP | api_tests/campaign_test.go, api_tests/order_test.go, api_tests/http_branch_coverage_test.go | TestCampaigns_OperationsManagerCanCreateAndMemberCanJoin |
| POST /api/v1/campaigns/:id/cancel | yes | TNM HTTP | api_tests/campaign_test.go | TestCampaigns_ListCancelAndEvaluateEndpoints |
| POST /api/v1/campaigns/:id/evaluate | yes | TNM HTTP | api_tests/campaign_test.go | TestCampaigns_ListCancelAndEvaluateEndpoints |
| POST /api/v1/orders | yes | TNM HTTP | api_tests/order_test.go, api_tests/report_export_test.go | TestOrders_MemberCanCreateListAndViewOnlyOwnOrders |
| GET /api/v1/orders | yes | TNM HTTP | api_tests/order_test.go | TestOrders_MemberCanCreateListAndViewOnlyOwnOrders |
| GET /api/v1/orders/:id | yes | TNM HTTP | api_tests/order_test.go | TestOrders_MemberCanCreateListAndViewOnlyOwnOrders |
| POST /api/v1/orders/:id/pay | yes | TNM HTTP | api_tests/order_test.go | TestOrders_ManagerCanPayNoteRefundAndReadTimeline |
| POST /api/v1/orders/:id/cancel | yes | TNM HTTP | api_tests/order_test.go | TestOrders_MemberCanCancelOwnOrderButNotOthers |
| POST /api/v1/orders/:id/refund | yes | TNM HTTP | api_tests/order_test.go | TestOrders_ManagerCanPayNoteRefundAndReadTimeline |
| POST /api/v1/orders/:id/notes | yes | TNM HTTP | api_tests/order_test.go | TestOrders_ManagerCanPayNoteRefundAndReadTimeline |
| GET /api/v1/orders/:id/timeline | yes | TNM HTTP | api_tests/order_test.go | TestOrders_ManagerCanPayNoteRefundAndReadTimeline |
| POST /api/v1/orders/:id/split | yes | TNM HTTP | api_tests/order_test.go | TestOrders_ManagerCanSplitOrder |
| POST /api/v1/orders/merge | yes | TNM HTTP | api_tests/order_test.go | TestOrders_ManagerCanMergeOrders |
| POST /api/v1/suppliers | yes | TNM HTTP | api_tests/procurement_test.go, api_tests/http_branch_coverage_test.go | TestProcurement_MemberCannotCreateSupplier |
| GET /api/v1/suppliers | yes | TNM HTTP | api_tests/procurement_test.go | TestProcurement_SupplierListGetAndUpdate |
| GET /api/v1/suppliers/:id | yes | TNM HTTP | api_tests/procurement_test.go | TestProcurement_SupplierListGetAndUpdate |
| PUT /api/v1/suppliers/:id | yes | TNM HTTP | api_tests/procurement_test.go, api_tests/http_branch_coverage_test.go | TestProcurement_SupplierListGetAndUpdate |
| POST /api/v1/purchase-orders | yes | TNM HTTP | api_tests/procurement_test.go, api_tests/http_branch_coverage_test.go | TestProcurement_CreateApproveReceiveAndResolveVarianceFlow |
| GET /api/v1/purchase-orders | yes | TNM HTTP | api_tests/procurement_test.go | TestProcurement_POListGetReturnVoidAndLandedCostSummary |
| GET /api/v1/purchase-orders/:id | yes | TNM HTTP | api_tests/procurement_test.go | TestProcurement_CreateApproveReceiveAndResolveVarianceFlow |
| POST /api/v1/purchase-orders/:id/approve | yes | TNM HTTP | api_tests/procurement_test.go, api_tests/http_branch_coverage_test.go | TestProcurement_CreateApproveReceiveAndResolveVarianceFlow |
| POST /api/v1/purchase-orders/:id/receive | yes | TNM HTTP | api_tests/procurement_test.go, api_tests/http_branch_coverage_test.go | TestProcurement_CreateApproveReceiveAndResolveVarianceFlow |
| POST /api/v1/purchase-orders/:id/return | yes | TNM HTTP | api_tests/procurement_test.go | TestProcurement_POListGetReturnVoidAndLandedCostSummary |
| POST /api/v1/purchase-orders/:id/void | yes | TNM HTTP | api_tests/procurement_test.go | TestProcurement_POListGetReturnVoidAndLandedCostSummary |
| GET /api/v1/variances | yes | TNM HTTP | api_tests/procurement_test.go | TestProcurement_CreateApproveReceiveAndResolveVarianceFlow |
| GET /api/v1/variances/:id | yes | TNM HTTP | api_tests/procurement_test.go, api_tests/http_branch_coverage_test.go | TestProcurement_CreateApproveReceiveAndResolveVarianceFlow |
| POST /api/v1/variances/:id/resolve | yes | TNM HTTP | api_tests/procurement_test.go, api_tests/http_branch_coverage_test.go | TestProcurement_CreateApproveReceiveAndResolveVarianceFlow |
| GET /api/v1/procurement/landed-costs | yes | TNM HTTP | api_tests/procurement_test.go | TestProcurement_POListGetReturnVoidAndLandedCostSummary |
| GET /api/v1/procurement/landed-costs/:poId | yes | TNM HTTP | api_tests/procurement_test.go | TestProcurement_CreateApproveReceiveAndResolveVarianceFlow |
| GET /api/v1/reports | yes | TNM HTTP | api_tests/integration_helpers_test.go | reportIDByType helper (called by report tests) |
| GET /api/v1/reports/:id/data | yes | TNM HTTP | api_tests/report_export_test.go | TestReports_CoachCannotReadAdminOnlyReportData |
| POST /api/v1/exports | yes | TNM HTTP | api_tests/report_export_test.go, api_tests/http_branch_coverage_test.go | TestReports_ExportDownloadUsesRealFilesAndAccessControl |
| GET /api/v1/exports/:id | yes | TNM HTTP | api_tests/report_export_test.go, api_tests/http_branch_coverage_test.go | TestReports_ExportDownloadUsesRealFilesAndAccessControl |
| GET /api/v1/exports/:id/download | yes | TNM HTTP | api_tests/report_export_test.go, api_tests/http_branch_coverage_test.go | TestReports_ExportDownloadUsesRealFilesAndAccessControl |
| GET /api/v1/admin/audit-log | yes | TNM HTTP | api_tests/admin_test.go | TestAdmin_CanCreateUserAndReadSecurityAudit |
| GET /api/v1/admin/audit-log/security | yes | TNM HTTP | api_tests/admin_test.go | TestAdmin_CanCreateUserAndReadSecurityAudit |
| POST /api/v1/admin/backups | yes | HTTP+Mock | api_tests/admin_test.go | TestAdmin_CanTriggerAndListBackups |
| GET /api/v1/admin/backups | yes | HTTP+Mock | api_tests/admin_test.go | TestAdmin_CanTriggerAndListBackups |
| GET /api/v1/admin/backups/:id/verify | yes | HTTP+Mock | api_tests/admin_test.go | TestAdmin_CanTriggerAndListBackups |
| POST /api/v1/admin/biometrics/rotate-key | yes | TNM HTTP | api_tests/biometric_test.go | TestBiometric_RotateKey |
| GET /api/v1/admin/biometrics/keys | yes | TNM HTTP | api_tests/biometric_test.go | TestBiometric_ListKeys |
| POST /api/v1/admin/biometrics | yes | TNM HTTP | api_tests/biometric_test.go | TestBiometric_Register_And_Get |
| GET /api/v1/admin/biometrics/:user_id | yes | TNM HTTP | api_tests/biometric_test.go | TestBiometric_Register_And_Get |
| POST /api/v1/admin/biometrics/:user_id/revoke | yes | TNM HTTP | api_tests/biometric_test.go | TestBiometric_Revoke |
| POST /api/v1/admin/users | yes | TNM HTTP | api_tests/admin_test.go | TestAdmin_UserGetUpdateDeactivateFlow |
| GET /api/v1/admin/users | yes | TNM HTTP | api_tests/rbac_test.go, api_tests/admin_test.go | TestRBAC_AdminCanAccessAdminUsersRoute |
| GET /api/v1/admin/users/:id | yes | TNM HTTP | api_tests/admin_test.go | TestAdmin_UserGetUpdateDeactivateFlow |
| PUT /api/v1/admin/users/:id | yes | TNM HTTP | api_tests/admin_test.go | TestAdmin_UserGetUpdateDeactivateFlow |
| POST /api/v1/admin/users/:id/deactivate | yes | TNM HTTP | api_tests/admin_test.go | TestAdmin_UserGetUpdateDeactivateFlow |
| GET /api/v1/admin/retention-policies | yes | TNM HTTP | api_tests/admin_test.go | TestAdmin_CanUpdateRetentionPolicyInDays |
| GET /api/v1/admin/retention-policies/:entity_type | yes | TNM HTTP | api_tests/admin_test.go | TestAdmin_CanUpdateRetentionPolicyInDays |
| PUT /api/v1/admin/retention-policies/:entity_type | yes | TNM HTTP | api_tests/admin_test.go, api_tests/http_branch_coverage_test.go | TestAdmin_CanUpdateRetentionPolicyInDays |
| POST /api/v1/locations | yes | TNM HTTP | api_tests/member_coach_location_test.go | TestPersonnel_AdminCanCreateLocationMemberAndCoach |
| GET /api/v1/locations | yes | TNM HTTP | api_tests/member_coach_location_test.go | TestPersonnel_AdminCanCreateLocationMemberAndCoach |
| GET /api/v1/locations/:id | yes | TNM HTTP | api_tests/member_coach_location_test.go, api_tests/http_branch_coverage_test.go | TestPersonnel_LocationGetIsForbiddenForOtherAssignedLocation |
| POST /api/v1/coaches | yes | TNM HTTP | api_tests/member_coach_location_test.go | TestPersonnel_AdminCanCreateLocationMemberAndCoach |
| GET /api/v1/coaches | yes | TNM HTTP | api_tests/member_coach_location_test.go | TestPersonnel_AdminCanAccessCrossLocationMemberAndCoachRecords |
| GET /api/v1/coaches/:id | yes | TNM HTTP | api_tests/member_coach_location_test.go | TestPersonnel_AdminCanCreateLocationMemberAndCoach |
| POST /api/v1/members | yes | TNM HTTP | api_tests/member_coach_location_test.go | TestPersonnel_AdminCanCreateLocationMemberAndCoach |
| GET /api/v1/members | yes | TNM HTTP | api_tests/member_coach_location_test.go | TestPersonnel_OperationsManagerIsScopedToAssignedLocation |
| GET /api/v1/members/:id | yes | TNM HTTP | api_tests/member_coach_location_test.go | TestPersonnel_AdminCanCreateLocationMemberAndCoach |

## API Test Classification
1. True No-Mock HTTP
- Files: auth_test.go, lockout_test.go, rbac_test.go, dashboard_test.go, item_test.go, inventory_test.go, campaign_test.go, order_test.go, procurement_test.go, report_export_test.go, biometric_test.go, member_coach_location_test.go, http_branch_coverage_test.go
- Evidence for real HTTP bootstrapping: integration_helpers_test.go -> newIntegrationAppWithConfig (bootstrap.NewApp + httptest.NewServer + real http.Client)

2. HTTP with Mocking
- Backup endpoints only in admin_test.go -> TestAdmin_CanTriggerAndListBackups
- Mock/stub source: integration_helpers_test.go -> newIntegrationAppWithConfig injects `testDump` function into bootstrap.NewApp instead of real dump implementation

3. Non-HTTP (unit/integration without HTTP)
- All repo/backend/unit_tests/**/*.go files

## Mock Detection Rules Findings
- `testDump` stub injected in API integration app bootstrap
  - What is mocked/stubbed: backup dump provider dependency
  - Where: api_tests/integration_helpers_test.go -> newIntegrationAppWithConfig
  - Affected endpoint tests: admin_test.go -> TestAdmin_CanTriggerAndListBackups
- No evidence in API tests of `jest.mock`, `vi.mock`, `sinon.stub`, GoMock expectations, or handler/service bypass for route execution.

## Coverage Summary
- Total endpoints: 82
- Endpoints with HTTP tests: 82
- Endpoints with TRUE no-mock tests: 79
- HTTP coverage: 100.00%
- True API coverage: 96.34%

## Unit Test Summary
- Unit test files discovered: 46
- Covered modules (by file/test naming and imports):
  - Services/application: auth, backup, biometric, campaign, coach, inventory, item, landed cost, location, member, order, procurement, report, retention, supplier, user, variance
  - Security/auth utilities: rbac, session, captcha, crypto, password, masking, audit helper
  - Domain objects/enums/errors/state machines/policies
  - HTTP helpers: validation and error envelope handling
  - Jobs and platform config
- Important modules not clearly unit-tested (direct, dedicated unit scope not evident from static naming):
  - Dashboard service logic (internal/application/dashboard_service.go)
  - HTTP middleware routing behavior (internal/http/middleware.go)
  - Router registration contract tests (internal/http/router.go)
  - DTO contract-focused tests (internal/http/dto/*)

## API Observability Check
- Strong overall observability:
  - Endpoint method/path visible in nearly all tests (`app.get`, `app.post`, `app.put` with explicit `/api/v1/...`)
  - Request input visible (maps/query/body strings)
  - Response content assertions present (status, decodeSuccess/decodeError, payload shape/value checks)
- Weak spots:
  - Some branch tests focus mostly on status code for malformed input with minimal payload assertions (e.g., subsets in http_branch_coverage_test.go)

## Tests Check
- Success paths: strong coverage across auth, catalog, orders, procurement, admin, reports
- Failure/validation paths: strong coverage, including invalid UUIDs, forbidden/unauthorized, validation errors
- Auth/permissions: strong coverage in rbac_test.go, dashboard_test.go, personnel and report tests
- Integration boundaries: strong for API flow with real DB and HTTP stack
- Assertions depth: generally meaningful, not auto-generated patterns
- run_tests.sh compliance:
  - Docker-based orchestration used (`docker compose --profile test run ...`) -> OK
  - No host-local dependency install instructions in script invocation contract -> OK

## End-to-End Expectations
- Project type is fullstack.
- No explicit real FE<->BE end-to-end test suite detected (no Playwright/Cypress e2e files under frontend).
- Partial compensation exists via:
  - strong backend API coverage
  - broad frontend unit test suite (route/component/hook tests)

## Test Coverage Score (0-100)
- **92/100**

## Score Rationale
- + Endpoint HTTP coverage is complete (82/82)
- + True no-mock API coverage is very high (79/82)
- + Broad scenario depth (success, failure, authorization, validation)
- + Strong unit-test breadth (46 files across core modules)
- - Backup API path relies on stubbed dump provider in test execution path
- - Missing true FE<->BE end-to-end tests for fullstack expectation
- - A few core transport/routing modules have limited direct unit tests

## Key Gaps
1. Backup API integration tests are not fully no-mock due to injected dump stub.
2. No browser-level fullstack end-to-end tests.
3. Direct unit tests for middleware/router/dashboard internals are less explicit than other domains.

## Confidence & Assumptions
- Confidence: high
- Assumptions:
  - Endpoint inventory is entirely derived from static route registration in router.go.
  - `reportIDByType` helper in integration_helpers_test.go is executed by report tests and therefore covers `GET /api/v1/reports`.
  - Only backup-related API paths are affected by the injected `testDump` dependency.

## Test Coverage Verdict
- **PASS (with strict caveat: backup endpoints are HTTP+Mock, not true no-mock)**

---

# README Audit

## README Location Check
- Required file exists: repo/README.md
- Result: PASS

## Project Type Detection
- Declared in README: `Project Type: Fullstack (frontend + backend + PostgreSQL)`
- Static repo structure confirms fullstack composition.

## Hard Gate Evaluation

### Formatting
- Markdown is structured and readable (headings, tables, code blocks)
- Result: PASS

### Startup Instructions (Fullstack)
- Includes `docker-compose up`
- Also includes `docker compose up --build`
- Result: PASS

### Access Method
- Explicit URLs/ports provided for frontend/backend/postgres
- Result: PASS

### Verification Method
- API verification via curl login/session/RBAC examples
- UI verification flow included
- Result: PASS

### Environment Rules (Docker-contained, strict)
- README does not instruct host-level runtime installs (`npm install`, `pip install`, `apt-get`, manual DB setup)
- Run path is docker-compose and `./run_tests.sh`
- Result: PASS

### Demo Credentials (auth exists)
- Authentication explicitly required
- Credentials provided for roles: Administrator, Operations Manager, Procurement Specialist, Coach, Member
- Result: PASS

## Engineering Quality
- Tech stack clarity: moderate (stack implied, not deeply explained)
- Architecture explanation: limited (quick-start oriented, little architectural context)
- Testing instructions: present and concise (`./run_tests.sh`, coverage flag)
- Security/roles: good practical guidance via demo credentials + RBAC verification steps
- Workflow clarity: good for boot/verify/stop/reset
- Presentation quality: clean, practical, concise

## High Priority Issues
- None.

## Medium Priority Issues
1. Architecture is minimally described; lacks concise component interaction overview.
2. README warns demo users may not be present in seed process, but does not provide deterministic local provisioning steps in the same document.

## Low Priority Issues
1. No explicit troubleshooting section for common local TLS/certificate or docker networking issues.
2. No quick matrix mapping roles to allowed feature areas; this would improve QA repeatability.

## Hard Gate Failures
- None.

## README Verdict (PASS / PARTIAL PASS / FAIL)
- **PASS**

## README Final Verdict
- **PASS**

---

# Final Combined Verdict
- Test Coverage Audit: **PASS with caveats**
- README Audit: **PASS**
- Overall strict-mode outcome: **PASS**
