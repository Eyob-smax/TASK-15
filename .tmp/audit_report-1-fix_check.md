# Delivery Acceptance Fix-Check

## 1) Source Report

- Source: [.tmp/audit_report-4.md](.tmp/audit_report-4.md)
- This fix-check re-reviews the issues recorded in that report using static evidence only.

## 2) Re-checked Issues

### Issue H1 (High)

- Prior issue from source report:
  - Offline-first requirement was not statically evidenced and appeared unimplemented.
- Current status: FIXED
- Verification evidence:
  - Offline persistence/hydration pipeline is present:
    - [repo/frontend/src/app/providers.tsx](repo/frontend/src/app/providers.tsx#L61)
    - [repo/frontend/src/app/providers.tsx](repo/frontend/src/app/providers.tsx#L63)
    - [repo/frontend/src/app/providers.tsx](repo/frontend/src/app/providers.tsx#L49)
  - IndexedDB-backed cache layer is implemented:
    - [repo/frontend/src/lib/offline-cache.ts](repo/frontend/src/lib/offline-cache.ts#L48)
    - [repo/frontend/src/lib/offline-cache.ts](repo/frontend/src/lib/offline-cache.ts#L108)
    - [repo/frontend/src/lib/offline-cache.ts](repo/frontend/src/lib/offline-cache.ts#L115)
  - Admin read domains are now included in cacheable roots:
    - [repo/frontend/src/lib/offline-cache.ts](repo/frontend/src/lib/offline-cache.ts#L32)
    - [repo/frontend/src/lib/offline-cache.ts](repo/frontend/src/lib/offline-cache.ts#L33)
    - [repo/frontend/src/lib/offline-cache.ts](repo/frontend/src/lib/offline-cache.ts#L34)
    - [repo/frontend/src/lib/offline-cache.ts](repo/frontend/src/lib/offline-cache.ts#L35)
    - [repo/frontend/src/lib/offline-cache.ts](repo/frontend/src/lib/offline-cache.ts#L36)
    - [repo/frontend/src/lib/offline-cache.ts](repo/frontend/src/lib/offline-cache.ts#L37)
  - Offline behavior tests exist:
    - [repo/frontend/unit_tests/lib/offline-cache.test.ts](repo/frontend/unit_tests/lib/offline-cache.test.ts#L35)
    - [repo/frontend/unit_tests/routes/CatalogPage.offline.test.tsx](repo/frontend/unit_tests/routes/CatalogPage.offline.test.tsx#L34)

### Issue H2 (High)

- Prior issue from source report:
  - CAPTCHA answers were stored in plaintext.
- Current status: FIXED
- Verification evidence:
  - Migration adds hashed/salted columns and drops plaintext column:
    - [repo/backend/database/migrations/00010_captcha_answer_hashing.sql](repo/backend/database/migrations/00010_captcha_answer_hashing.sql#L5)
    - [repo/backend/database/migrations/00010_captcha_answer_hashing.sql](repo/backend/database/migrations/00010_captcha_answer_hashing.sql#L42)
  - Store layer uses answer_hash and answer_salt:
    - [repo/backend/internal/store/postgres/captcha_store.go](repo/backend/internal/store/postgres/captcha_store.go#L30)
    - [repo/backend/internal/store/postgres/captcha_store.go](repo/backend/internal/store/postgres/captcha_store.go#L53)
  - Auth flow verifies via derived hash path:
    - [repo/backend/internal/application/auth_service.go](repo/backend/internal/application/auth_service.go#L253)

### Issue H3 (High)

- Prior issue from source report:
  - API documentation materially diverged from implementation contracts.
- Current status: FIXED
- Verification evidence:
  - Spec now states cookie-only login token behavior:
    - [docs/api-spec.md](docs/api-spec.md#L167)
  - Spec and implementation align on logout 200 message:
    - [docs/api-spec.md](docs/api-spec.md#L173)
    - [docs/api-spec.md](docs/api-spec.md#L185)
    - [repo/backend/internal/http/auth_handler.go](repo/backend/internal/http/auth_handler.go#L78)
  - Spec and implementation align on captcha unauthorized semantics:
    - [docs/api-spec.md](docs/api-spec.md#L217)
    - [repo/backend/internal/http/errors.go](repo/backend/internal/http/errors.go#L56)

### Issue M4 (Medium)

- Prior issue from source report:
  - CAPTCHA challenge generation used non-cryptographic randomness.
- Current status: FIXED
- Verification evidence:
  - Security implementation uses crypto random and constant-time verification:
    - [repo/backend/internal/security/captcha.go](repo/backend/internal/security/captcha.go#L4)
    - [repo/backend/internal/security/captcha.go](repo/backend/internal/security/captcha.go#L17)
    - [repo/backend/internal/security/captcha.go](repo/backend/internal/security/captcha.go#L57)

### Issue M5 (Medium)

- Prior issue from source report:
  - Object-level authorization relied mainly on handler checks and service-layer consistency was not universal.
- Current status: FIXED
- Verification evidence:
  - Service-layer actor-aware authorization exists for key read/cancel and split/merge paths:
    - [repo/backend/internal/application/order_service.go](repo/backend/internal/application/order_service.go#L99)
    - [repo/backend/internal/application/order_service.go](repo/backend/internal/application/order_service.go#L114)
    - [repo/backend/internal/application/order_service.go](repo/backend/internal/application/order_service.go#L202)
    - [repo/backend/internal/application/order_service.go](repo/backend/internal/application/order_service.go#L514)
    - [repo/backend/internal/application/order_service.go](repo/backend/internal/application/order_service.go#L521)
    - [repo/backend/internal/application/member_service.go](repo/backend/internal/application/member_service.go#L51)
    - [repo/backend/internal/application/coach_service.go](repo/backend/internal/application/coach_service.go#L50)
  - HTTP handlers now call actor-scoped split/merge service entry points:
    - [repo/backend/internal/http/order_handler.go](repo/backend/internal/http/order_handler.go#L244)
    - [repo/backend/internal/http/order_handler.go](repo/backend/internal/http/order_handler.go#L282)

### Issue L6 (Low)

- Prior issue from source report:
  - API spec examples could mislead integrators on login token handling.
- Current status: FIXED
- Verification evidence:
  - Spec explicitly documents token as cookie-only:
    - [docs/api-spec.md](docs/api-spec.md#L167)
  - Handler comment and response behavior are consistent:
    - [repo/backend/internal/http/auth_handler.go](repo/backend/internal/http/auth_handler.go#L28)
    - [repo/backend/internal/http/auth_handler.go](repo/backend/internal/http/auth_handler.go#L67)

## 3) Additional Gaps Re-check from Source Report

### Gap G1 (Security Coverage Note)

- Prior gap from source report:
  - CAPTCHA storage-strength validation gap.
- Current status: FIXED
- Verification evidence:
  - Security unit tests cover salted hashing and include a persistence-shape guard:
    - [repo/backend/unit_tests/security/captcha_test.go](repo/backend/unit_tests/security/captcha_test.go#L27)
    - [repo/backend/unit_tests/security/captcha_test.go](repo/backend/unit_tests/security/captcha_test.go#L36)
    - [repo/backend/unit_tests/security/captcha_test.go](repo/backend/unit_tests/security/captcha_test.go#L94)
  - Postgres-boundary API test now asserts DB-row hash/salt storage and no plaintext `answer` column:
    - [repo/backend/api_tests/lockout_test.go](repo/backend/api_tests/lockout_test.go#L109)
    - [repo/backend/api_tests/lockout_test.go](repo/backend/api_tests/lockout_test.go#L116)
    - [repo/backend/api_tests/lockout_test.go](repo/backend/api_tests/lockout_test.go#L143)

### Gap G2 (Offline Test Coverage Note)

- Prior gap from source report:
  - No static evidence of frontend tests validating offline behavior.
- Current status: FIXED
- Verification evidence:
  - Offline cache key coverage tests:
    - [repo/frontend/unit_tests/lib/offline-cache.test.ts](repo/frontend/unit_tests/lib/offline-cache.test.ts#L35)
  - Offline route behavior tests:
    - [repo/frontend/unit_tests/routes/CatalogPage.offline.test.tsx](repo/frontend/unit_tests/routes/CatalogPage.offline.test.tsx#L53)
    - [repo/frontend/unit_tests/routes/GroupBuysPage.test.tsx](repo/frontend/unit_tests/routes/GroupBuysPage.test.tsx#L113)

## 4) Fix-Check Verdict

- Pass
- Summary:
  - Fixed: H1, H2, H3, M4, M5, L6, G1, G2
  - Partially fixed: none
  - Not fixed: none

## 5) Notes

- This is a static fix-check only.
- No project run, no test execution, and no Docker commands were used for this validation.
- All issues from the source report are now resolved by static evidence.
