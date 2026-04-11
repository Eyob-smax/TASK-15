# Delivery Acceptance Fix-Check

## 1) Source Report

- Source: [.tmp/static_audit_report_2026-04-11_static_only.md](.tmp/static_audit_report_2026-04-11_static_only.md)
- This fix-check re-reviews the issues and remaining risks listed in that source report.
- Scope: static verification only (no runtime execution, no Docker startup, no test execution).

## 2) Re-checked Findings From Source Report

### Finding H1 (High)

- Prior finding from source report:
  - Pricing semantics previously used refundable deposit as order unit price.
- Current status: FIXED
- Verification evidence:
  - Order creation uses item unit price and computes total from that value:
    - [repo/backend/internal/application/order_service.go](repo/backend/internal/application/order_service.go#L63)
    - [repo/backend/internal/application/order_service.go](repo/backend/internal/application/order_service.go#L64)
  - Campaign join order creation uses item unit price:
    - [repo/backend/internal/application/campaign_service.go](repo/backend/internal/application/campaign_service.go#L120)
  - Unit tests assert unit-price semantics:
    - [repo/backend/unit_tests/orders/order_service_test.go](repo/backend/unit_tests/orders/order_service_test.go#L255)
    - [repo/backend/unit_tests/campaign/campaign_service_test.go](repo/backend/unit_tests/campaign/campaign_service_test.go#L376)

### Finding H2 (High)

- Prior finding from source report:
  - Actor attribution in payment/split/merge/auto-close timeline and audit paths was incorrect.
- Current status: FIXED
- Verification evidence:
  - Pay path threads authenticated actor through handler to service:
    - [repo/backend/internal/http/order_handler.go](repo/backend/internal/http/order_handler.go#L128)
    - [repo/backend/internal/application/order_service.go](repo/backend/internal/application/order_service.go#L138)
    - [repo/backend/internal/application/order_service.go](repo/backend/internal/application/order_service.go#L159)
  - Split/merge actor-aware flows are implemented:
    - [repo/backend/internal/application/order_service.go](repo/backend/internal/application/order_service.go#L283)
    - [repo/backend/internal/application/order_service.go](repo/backend/internal/application/order_service.go#L399)
    - [repo/backend/internal/application/order_service.go](repo/backend/internal/application/order_service.go#L536)
    - [repo/backend/internal/application/order_service.go](repo/backend/internal/application/order_service.go#L543)
  - Split/merge actor assertions are present in tests:
    - [repo/backend/unit_tests/orders/order_service_test.go](repo/backend/unit_tests/orders/order_service_test.go#L692)
    - [repo/backend/unit_tests/orders/order_service_test.go](repo/backend/unit_tests/orders/order_service_test.go#L741)

### Finding H3 (High)

- Prior finding from source report:
  - Campaign-origin orders were missing timeline provenance.
- Current status: FIXED
- Verification evidence:
  - Campaign join creates order timeline entry:
    - [repo/backend/internal/application/campaign_service.go](repo/backend/internal/application/campaign_service.go#L134)
  - Campaign service unit tests verify timeline write:
    - [repo/backend/unit_tests/campaign/campaign_service_test.go](repo/backend/unit_tests/campaign/campaign_service_test.go#L379)

### Finding H4 (High)

- Prior finding from source report:
  - Export generation swallowed report query failures.
- Current status: FIXED
- Verification evidence:
  - Query failure marks export job as failed and returns error:
    - [repo/backend/internal/application/report_service.go](repo/backend/internal/application/report_service.go#L310)
    - [repo/backend/internal/application/report_service.go](repo/backend/internal/application/report_service.go#L311)
  - Regression tests cover failed status and failed persistence:
    - [repo/backend/unit_tests/reports/report_service_test.go](repo/backend/unit_tests/reports/report_service_test.go#L241)
    - [repo/backend/unit_tests/reports/report_service_test.go](repo/backend/unit_tests/reports/report_service_test.go#L266)
    - [repo/backend/unit_tests/reports/report_service_test.go](repo/backend/unit_tests/reports/report_service_test.go#L291)

### Finding H5 (High)

- Prior finding from source report:
  - Offline-first mutation path was unavailable for catalog/report actions.
- Current status: FIXED
- Verification evidence:
  - Offline queue and replay wiring exists:
    - [repo/frontend/src/lib/offline-cache.ts](repo/frontend/src/lib/offline-cache.ts#L149)
    - [repo/frontend/src/lib/hooks/useItems.ts](repo/frontend/src/lib/hooks/useItems.ts#L148)
    - [repo/frontend/src/lib/hooks/useReports.ts](repo/frontend/src/lib/hooks/useReports.ts#L44)
    - [repo/frontend/src/app/providers.tsx](repo/frontend/src/app/providers.tsx#L62)
    - [repo/frontend/src/app/providers.tsx](repo/frontend/src/app/providers.tsx#L82)
  - Replay now marks non-offline failures as failed and retains entries for later inspection/retry:
    - [repo/frontend/src/lib/offline-mutations.ts](repo/frontend/src/lib/offline-mutations.ts#L70)
    - [repo/frontend/src/lib/offline-mutations.ts](repo/frontend/src/lib/offline-mutations.ts#L87)
  - Failed queued actions are surfaced to operators in-app:
    - [repo/frontend/src/components/FailedOfflineMutationsAlert.tsx](repo/frontend/src/components/FailedOfflineMutationsAlert.tsx#L13)
    - [repo/frontend/src/components/Layout.tsx](repo/frontend/src/components/Layout.tsx#L136)

## 3) Remaining Risks Re-check

### Risk R1 (High)

- Prior risk from source report:
  - Offline replay drops queued mutations on non-offline errors without operator visibility.
- Current status: FIXED
- Verification evidence:
  - Queue entry is no longer removed on generic non-offline failures and is marked as failed:
    - [repo/frontend/src/lib/offline-mutations.ts](repo/frontend/src/lib/offline-mutations.ts#L70)
    - [repo/frontend/src/lib/offline-mutations.ts](repo/frontend/src/lib/offline-mutations.ts#L87)
  - Failed status metadata is persisted in queue storage:
    - [repo/frontend/src/lib/offline-cache.ts](repo/frontend/src/lib/offline-cache.ts#L44)
    - [repo/frontend/src/lib/offline-cache.ts](repo/frontend/src/lib/offline-cache.ts#L190)
  - Operator-facing failed-queue UI is present and wired into app layout:
    - [repo/frontend/src/components/FailedOfflineMutationsAlert.tsx](repo/frontend/src/components/FailedOfflineMutationsAlert.tsx#L19)
    - [repo/frontend/src/components/FailedOfflineMutationsAlert.tsx](repo/frontend/src/components/FailedOfflineMutationsAlert.tsx#L44)
    - [repo/frontend/src/components/Layout.tsx](repo/frontend/src/components/Layout.tsx#L136)
  - UI behavior is covered by component tests:
    - [repo/frontend/unit_tests/components/FailedOfflineMutationsAlert.test.tsx](repo/frontend/unit_tests/components/FailedOfflineMutationsAlert.test.tsx#L19)
    - [repo/frontend/unit_tests/components/FailedOfflineMutationsAlert.test.tsx](repo/frontend/unit_tests/components/FailedOfflineMutationsAlert.test.tsx#L84)

### Risk R2 (Medium)

- Prior risk from source report:
  - Synthetic client IDs for offline create are not reconciled to server IDs.
- Current status: FIXED
- Verification evidence:
  - Offline create stores temporary ID in queued payload for later mapping:
    - [repo/frontend/src/lib/hooks/useItems.ts](repo/frontend/src/lib/hooks/useItems.ts#L148)
  - Replay extracts server ID and stores temp->server mapping:
    - [repo/frontend/src/lib/offline-mutations.ts](repo/frontend/src/lib/offline-mutations.ts#L19)
    - [repo/frontend/src/lib/offline-mutations.ts](repo/frontend/src/lib/offline-mutations.ts#L38)
    - [repo/frontend/src/lib/offline-cache.ts](repo/frontend/src/lib/offline-cache.ts#L243)
  - Item reads resolve mapped ID before API lookup:
    - [repo/frontend/src/lib/hooks/useItems.ts](repo/frontend/src/lib/hooks/useItems.ts#L45)
    - [repo/frontend/src/lib/offline-cache.ts](repo/frontend/src/lib/offline-cache.ts#L263)
  - Offline create path no longer navigates to temporary detail route:
    - [repo/frontend/src/routes/CatalogFormPage.tsx](repo/frontend/src/routes/CatalogFormPage.tsx#L109)

### Risk R3 (Medium)

- Prior risk from source report:
  - Missing regression tests for export query failure path.
- Current status: FIXED
- Verification evidence:
  - Failure-path tests now exist and assert failed export status:
    - [repo/backend/unit_tests/reports/report_service_test.go](repo/backend/unit_tests/reports/report_service_test.go#L241)
    - [repo/backend/unit_tests/reports/report_service_test.go](repo/backend/unit_tests/reports/report_service_test.go#L261)
    - [repo/backend/unit_tests/reports/report_service_test.go](repo/backend/unit_tests/reports/report_service_test.go#L291)

### Risk R4 (Medium)

- Prior risk from source report:
  - No frontend unit tests for offline mutation queue and replay.
- Current status: FIXED
- Verification evidence:
  - Queue/replay-specific unit tests now exist:
    - [repo/frontend/unit_tests/lib/offline-mutations.test.ts](repo/frontend/unit_tests/lib/offline-mutations.test.ts#L1)
    - [repo/frontend/unit_tests/lib/offline-mutations.test.ts](repo/frontend/unit_tests/lib/offline-mutations.test.ts#L34)
  - Tests cover mapping, failed-marking retention, offline-stop, and failed-entry skip behaviors:
    - [repo/frontend/unit_tests/lib/offline-mutations.test.ts](repo/frontend/unit_tests/lib/offline-mutations.test.ts#L61)
    - [repo/frontend/unit_tests/lib/offline-mutations.test.ts](repo/frontend/unit_tests/lib/offline-mutations.test.ts#L88)
    - [repo/frontend/unit_tests/lib/offline-mutations.test.ts](repo/frontend/unit_tests/lib/offline-mutations.test.ts#L119)
    - [repo/frontend/unit_tests/lib/offline-mutations.test.ts](repo/frontend/unit_tests/lib/offline-mutations.test.ts#L147)
  - Failed-queue operator UI behavior is also tested:
    - [repo/frontend/unit_tests/components/FailedOfflineMutationsAlert.test.tsx](repo/frontend/unit_tests/components/FailedOfflineMutationsAlert.test.tsx#L59)
    - [repo/frontend/unit_tests/components/FailedOfflineMutationsAlert.test.tsx](repo/frontend/unit_tests/components/FailedOfflineMutationsAlert.test.tsx#L84)

## 4) Fix-Check Verdict

- Pass
- Summary:
  - Fixed: H1, H2, H3, H4, H5, R1, R2, R3, R4

## 5) Notes

- This is a static fix-check only.
- No project run, no test execution, and no Docker commands were used for this validation.
