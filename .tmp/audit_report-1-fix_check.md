# Delivery Acceptance Fix-Check

## 1) Source Report

- Source: [.tmp/fitcommerce-static-audit.md](.tmp/fitcommerce-static-audit.md)
- This fix-check re-reviews only the issues listed in that report using static evidence.

## 2) Re-checked Issues

### Issue H1 (High)

- Prior issue from source report:
  - KPI business semantics and filter fidelity were not fully aligned with prompt requirements.
- Current status: FIXED
- Verification evidence:
  - Coach filter propagation remains wired through KPI calls and helper queries:
    - [repo/backend/internal/application/dashboard_service.go](repo/backend/internal/application/dashboard_service.go#L52)
    - [repo/backend/internal/application/dashboard_service.go](repo/backend/internal/application/dashboard_service.go#L257)
  - Renewal-rate logic now uses due-for-renewal denominator and renewed numerator for both current and previous windows:
    - [repo/backend/internal/application/dashboard_service.go](repo/backend/internal/application/dashboard_service.go#L154)
    - [repo/backend/internal/application/dashboard_service.go](repo/backend/internal/application/dashboard_service.go#L159)
    - [repo/backend/internal/application/dashboard_service.go](repo/backend/internal/application/dashboard_service.go#L167)
    - [repo/backend/internal/application/dashboard_service.go](repo/backend/internal/application/dashboard_service.go#L172)
  - Class fill rate now uses class-capacity occupancy (current_committed_qty / max_quantity) instead of min-quantity campaign progress:
    - [repo/backend/internal/application/dashboard_service.go](repo/backend/internal/application/dashboard_service.go#L205)
    - [repo/backend/internal/application/dashboard_service.go](repo/backend/internal/application/dashboard_service.go#L430)

### Issue H2 (High)

- Prior issue from source report:
  - Fulfillment grouping dimensions (supplier/bin/pickup) were supported in backend but not exposed in frontend split/merge workflows.
- Current status: FIXED
- Verification evidence:
  - Merge dialog now captures fulfillment grouping fields:
    - [repo/frontend/src/routes/OrdersPage.tsx](repo/frontend/src/routes/OrdersPage.tsx#L81)
    - [repo/frontend/src/routes/OrdersPage.tsx](repo/frontend/src/routes/OrdersPage.tsx#L89)
    - [repo/frontend/src/routes/OrdersPage.tsx](repo/frontend/src/routes/OrdersPage.tsx#L97)
  - Merge request payload now sends supplier/bin/pickup:
    - [repo/frontend/src/routes/OrdersPage.tsx](repo/frontend/src/routes/OrdersPage.tsx#L173)
    - [repo/frontend/src/routes/OrdersPage.tsx](repo/frontend/src/routes/OrdersPage.tsx#L174)
    - [repo/frontend/src/routes/OrdersPage.tsx](repo/frontend/src/routes/OrdersPage.tsx#L175)
  - Split dialog now captures fulfillment grouping fields and submits them:
    - [repo/frontend/src/routes/OrderDetailPage.tsx](repo/frontend/src/routes/OrderDetailPage.tsx#L297)
    - [repo/frontend/src/routes/OrderDetailPage.tsx](repo/frontend/src/routes/OrderDetailPage.tsx#L305)
    - [repo/frontend/src/routes/OrderDetailPage.tsx](repo/frontend/src/routes/OrderDetailPage.tsx#L313)
    - [repo/frontend/src/routes/OrderDetailPage.tsx](repo/frontend/src/routes/OrderDetailPage.tsx#L259)
    - [repo/frontend/src/routes/OrderDetailPage.tsx](repo/frontend/src/routes/OrderDetailPage.tsx#L260)
    - [repo/frontend/src/routes/OrderDetailPage.tsx](repo/frontend/src/routes/OrderDetailPage.tsx#L261)
  - Hooks and offline replay pipeline now preserve these fields:
    - [repo/frontend/src/lib/hooks/useOrders.ts](repo/frontend/src/lib/hooks/useOrders.ts#L157)
    - [repo/frontend/src/lib/hooks/useOrders.ts](repo/frontend/src/lib/hooks/useOrders.ts#L165)
    - [repo/frontend/src/lib/hooks/useOrders.ts](repo/frontend/src/lib/hooks/useOrders.ts#L196)
    - [repo/frontend/src/lib/hooks/useOrders.ts](repo/frontend/src/lib/hooks/useOrders.ts#L204)
    - [repo/frontend/src/lib/offline-mutations.ts](repo/frontend/src/lib/offline-mutations.ts#L97)
    - [repo/frontend/src/lib/offline-mutations.ts](repo/frontend/src/lib/offline-mutations.ts#L105)

### Issue M3 (Medium)

- Prior issue from source report:
  - Backup scheduling was interval-based (24h ticker) instead of explicit nightly scheduling.
- Current status: FIXED
- Verification evidence:
  - Backup job now distinguishes scheduled mode from interval mode:
    - [repo/backend/internal/jobs/procurement_jobs.go](repo/backend/internal/jobs/procurement_jobs.go#L19)
  - Production constructor now uses scheduled (clock-based) mode:
    - [repo/backend/internal/jobs/procurement_jobs.go](repo/backend/internal/jobs/procurement_jobs.go#L22)
  - Run loop now computes next midnight trigger:
    - [repo/backend/internal/jobs/procurement_jobs.go](repo/backend/internal/jobs/procurement_jobs.go#L54)
    - [repo/backend/internal/jobs/procurement_jobs.go](repo/backend/internal/jobs/procurement_jobs.go#L58)
    - [repo/backend/internal/jobs/procurement_jobs.go](repo/backend/internal/jobs/procurement_jobs.go#L59)

### Issue M4 (Medium)

- Prior issue from source report:
  - Biometric production path had limited static verification evidence in API tests.
- Current status: FIXED
- Verification evidence:
  - Dedicated biometric API test suite exists:
    - [repo/backend/api_tests/biometric_test.go](repo/backend/api_tests/biometric_test.go#L18)
    - [repo/backend/api_tests/biometric_test.go](repo/backend/api_tests/biometric_test.go#L52)
    - [repo/backend/api_tests/biometric_test.go](repo/backend/api_tests/biometric_test.go#L84)
    - [repo/backend/api_tests/biometric_test.go](repo/backend/api_tests/biometric_test.go#L118)
    - [repo/backend/api_tests/biometric_test.go](repo/backend/api_tests/biometric_test.go#L145)
    - [repo/backend/api_tests/biometric_test.go](repo/backend/api_tests/biometric_test.go#L168)
  - Test config explicitly enables biometric module and key ref:
    - [repo/backend/api_tests/biometric_test.go](repo/backend/api_tests/biometric_test.go#L13)
    - [repo/backend/api_tests/biometric_test.go](repo/backend/api_tests/biometric_test.go#L14)
  - Integration harness supports per-test config override used by biometric tests:
    - [repo/backend/api_tests/integration_helpers_test.go](repo/backend/api_tests/integration_helpers_test.go#L101)

## 3) Fix-Check Verdict

- Pass
- Summary:
  - Fixed: H1, H2, M3, M4
  - Partially fixed: none
  - Not fixed: none

## 4) Notes

- This is a static fix-check only.
- No project run, no test execution, and no Docker commands were used.
- No unresolved findings from the prior source report remain in this static re-check.
