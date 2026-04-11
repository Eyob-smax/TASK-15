# FitCommerce Operations & Inventory Suite -- Requirement Traceability Matrix

This document maps every explicit requirement to the frontend module, backend module, database table(s), and test file(s) that implement and verify it.

Test files span:
- `backend/unit_tests/` — domain (6+ files), security (7 files), catalog, orders, campaign, procurement, backup, admin, reports, jobs, platform (config)
- `backend/api_tests/` — auth, items, orders, campaign, procurement, admin, inventory (error envelopes, RBAC, CRUD flows)
- `frontend/unit_tests/` — auth (3 files), lib (2 files), components (3 files), routes (14 files)

Run all tests: `./run_tests.sh` from `repo/`. Coverage reports written to `backend/coverage_unit.out`, `backend/coverage_api.out`, and `frontend/coverage/`.

---

## 1. Roles & Access Control

| ID | Requirement | Frontend Module | Backend Module | Database | Tests |
|----|-------------|-----------------|----------------|----------|-------|
| RAC-01 | Five roles: Administrator, Operations Manager, Procurement Specialist, Coach, Member | `lib/constants.ts` (USER_ROLES, ROLE_LABELS), `lib/types.ts` (UserRole) | `domain/enums.go` (UserRole, AllUserRoles) | `users.role` (user_role enum in migration 00001) | `backend/unit_tests/domain/` |
| RAC-02 | Route-level RBAC middleware blocks unauthorized roles | `app/routes.tsx` (allowedRoles per route) | `http/middleware.go` (RequireRole) | -- | `backend/api_tests/` |
| RAC-03 | Frontend role-aware route guards prevent navigation | `lib/auth.tsx` (ProtectedRoute, RequireRole) | -- | -- | `frontend/unit_tests/` |
| RAC-04 | Role permission matrix determines module access | `lib/constants.ts` (ROLE_PERMISSIONS) | `http/middleware.go` (RequireRole per group) | -- | `frontend/unit_tests/lib/` |
| RAC-05 | Members see only their own data (scope filtering) | `lib/auth.tsx` (user context) | `application/services.go` (OrderService.List userID filter) | `orders.user_id` FK | `backend/api_tests/` |
| RAC-06 | Coaches see only their location's data | `features/dashboard/types.ts` (KPIFilters.location_id) | `application/services.go` (DashboardService, MemberRepository) | `coaches.location_id`, `members.location_id` | `backend/unit_tests/domain/` |
| RAC-07 | Administrator can manage all users | `routes/UsersPage.tsx` (admin only) | `application/services.go` (UserService) | `users` | `backend/api_tests/` |

---

## 2. Dashboard & KPIs

| ID | Requirement | Frontend Module | Backend Module | Database | Tests |
|----|-------------|-----------------|----------------|----------|-------|
| KPI-01 | Member growth KPI | `features/dashboard/types.ts` (KPIDashboard.member_growth), `routes/DashboardPage.tsx` | `application/services.go` (DashboardService), `reporting/reporting.go` | `members` (count by period) | `backend/unit_tests/domain/` |
| KPI-02 | Churn rate KPI | `features/dashboard/types.ts` (KPIDashboard.churn) | `application/services.go` (DashboardKPIs.Churn) | `members.membership_status` | `backend/unit_tests/domain/` |
| KPI-03 | Renewal rate KPI | `features/dashboard/types.ts` (KPIDashboard.renewal_rate) | `application/services.go` (DashboardKPIs.RenewalRate) | `members.renewal_date` | `backend/unit_tests/domain/` |
| KPI-04 | Engagement score KPI | `features/dashboard/types.ts` (KPIDashboard.engagement) | `application/services.go` (DashboardKPIs.Engagement) | `orders`, `group_buy_participants` | `backend/unit_tests/domain/` |
| KPI-05 | Class fill rate KPI | `features/dashboard/types.ts` (KPIDashboard.class_fill_rate) | `application/services.go` (DashboardKPIs.ClassFillRate) | `items`, `orders` | `backend/unit_tests/domain/` |
| KPI-06 | Coach productivity KPI | `features/dashboard/types.ts` (KPIDashboard.coach_productivity) | `application/services.go` (DashboardKPIs.CoachProductivity) | `coaches`, `members` | `backend/unit_tests/domain/` |
| KPI-07 | Time period toggles (daily/weekly/monthly/quarterly/yearly) | `features/dashboard/types.ts` (KPIFilters.period) | `http/router.go` (GET /dashboard/kpis query params) | -- | `frontend/unit_tests/` |
| KPI-08 | Location, coach, category filters | `features/dashboard/types.ts` (KPIFilters) | `http/router.go` (query params: location, coach, category, from, to) | `locations`, `coaches` | `backend/api_tests/` |
| KPI-09 | Dashboard accessible by admin, ops_mgr, procurement, coach (not member) | `app/routes.tsx` (DashboardPage -- no explicit restriction, but `CanAccessDashboard`) | `domain/enums.go` (UserRole.CanAccessDashboard) | -- | `backend/unit_tests/domain/` |

---

## 3. Exports

| ID | Requirement | Frontend Module | Backend Module | Database | Tests |
|----|-------------|-----------------|----------------|----------|-------|
| EXP-01 | CSV export format | `lib/validation.ts` (createExportSchema: format csv/pdf) | `reporting/reporting.go` (CSV generation), `domain/report.go` (ExportFormatCSV) | `export_jobs.format` | `backend/unit_tests/domain/` |
| EXP-02 | PDF export format | `lib/validation.ts` (createExportSchema) | `reporting/reporting.go` (PDF generation), `domain/report.go` (ExportFormatPDF) | `export_jobs.format` | `backend/unit_tests/domain/` |
| EXP-03 | Timestamped filename (report_type_YYYYMMDD_HHmmss.ext) | `lib/format.ts` (formatExportFilename) | `domain/report.go` (GenerateExportFilename) | `export_jobs.filename` | `backend/unit_tests/domain/`, `frontend/unit_tests/lib/` |
| EXP-04 | Role-aware data masking in exports | `lib/format.ts` (maskField) | `security/security.go` (data masking), `reporting/reporting.go` | -- | `frontend/unit_tests/lib/` |
| EXP-05 | Export job lifecycle (pending -> processing -> completed/failed) | `lib/types.ts` (ExportStatus) | `domain/enums.go` (ExportStatus*) | `export_jobs.status` | `backend/unit_tests/domain/` |
| EXP-06 | Download endpoint with Content-Disposition header | -- (browser native download) | `http/router.go` (GET /exports/:id/download) | `export_jobs.file_path` | `backend/api_tests/` |

---

## 4. Catalog Management

> Implementation: `internal/application/item_service.go`, `internal/store/postgres/item_store.go`, `internal/http/item_handler.go`

| ID | Requirement | Frontend Module | Backend Module | Database | Tests |
|----|-------------|-----------------|----------------|----------|-------|
| CAT-01 | Multi-spec items (category, brand, condition, billing model) | `features/catalog/types.ts` (ItemFormData), `lib/validation.ts` (createItemSchema) | `domain/item.go` (Item struct), `domain/enums.go` (ItemCondition, BillingModel) | `items` (category, brand, condition, billing_model columns) | `backend/unit_tests/domain/` |
| CAT-02 | Draft/published/unpublished status lifecycle | `lib/constants.ts` (ITEM_STATUSES), `lib/types.ts` (ItemStatus) | `domain/enums.go` (ItemStatus*), `http/router.go` (publish/unpublish routes) | `items.status` (item_status enum) | `backend/unit_tests/domain/` |
| CAT-03 | Publish validation (all required fields, no window overlaps) | -- (server-side validation) | `domain/policies.go` (ValidateItemForPublish, DetectWindowOverlap) | `items`, `item_availability_windows`, `item_blackout_windows` | `backend/unit_tests/domain/` |
| CAT-04 | Availability windows | `lib/types.ts` (AvailabilityWindow), `lib/validation.ts` (availabilityWindowSchema) | `domain/item.go` (AvailabilityWindow) | `item_availability_windows` | `backend/unit_tests/domain/` |
| CAT-05 | Blackout windows | `lib/types.ts` (BlackoutWindow), `lib/validation.ts` (blackoutWindowSchema) | `domain/item.go` (BlackoutWindow) | `item_blackout_windows` | `backend/unit_tests/domain/` |
| CAT-06 | Batch edit (multiple items, single operation, including availability windows) | `features/catalog/types.ts` (BatchEditFormData), `lib/validation.ts` (batchEditRowSchema), `routes/CatalogPage.tsx` | `domain/batch_edit.go`, `application/services.go` (ItemService.BatchEdit), `http/dto/requests.go` (BatchEditRequest) | `batch_edit_jobs`, `batch_edit_results`, `item_availability_windows` | `backend/unit_tests/domain/`, `backend/api_tests/item_test.go`, `frontend/unit_tests/routes/CatalogPage.test.tsx` |
| CAT-07 | Batch partial failure reporting | -- (API error handling) | `domain/errors.go` (ErrBatchEditPartialFailure), `http/errors.go` (BATCH_PARTIAL_FAILURE, 207) | `batch_edit_results.success`, `batch_edit_results.failure_reason` | `backend/unit_tests/domain/` |
| CAT-08 | Default refundable deposit ($50) | `lib/validation.ts` (refundable_deposit default 50), `lib/constants.ts` (DEFAULT_REFUNDABLE_DEPOSIT) | `domain/item.go` (DefaultRefundableDeposit, ApplyDepositDefault), `domain/policies.go` (ApplyDepositDefault) | `items.refundable_deposit DEFAULT 50.00` | `backend/unit_tests/domain/` |
| CAT-09 | Optimistic concurrency via version field | `lib/types.ts` (Item -- no explicit version; DTO has it) | `domain/item.go` (Item.Version), `http/dto/requests.go` (UpdateItemRequest.Version) | `items.version` | `backend/unit_tests/domain/` |

---

## 4a. Inventory

> Implementation: `internal/application/inventory_service.go`, `internal/store/postgres/inventory_store.go`, `internal/http/inventory_handler.go`

| ID | Requirement | Frontend Module | Backend Module | Database | Tests |
|----|-------------|-----------------|----------------|----------|-------|
| INV-01 | Inventory snapshots (point-in-time quantity records) | `features/inventory/types.ts`, `routes/InventoryPage.tsx` | `domain/inventory.go` (InventorySnapshot), `application/services.go` (InventoryService) | `inventory_snapshots` | `backend/unit_tests/domain/` |
| INV-02 | Manual inventory adjustments with reason | `lib/validation.ts` (adjustmentSchema), `routes/InventoryPage.tsx` | `domain/inventory.go` (InventoryAdjustment), `application/services.go` (InventoryService.Adjust) | `inventory_adjustments` | `backend/unit_tests/domain/` |
| INV-03 | Warehouse bin management | `routes/InventoryPage.tsx` | `domain/inventory.go` (WarehouseBin), `application/services.go` (InventoryService), `http/router.go` (warehouse-bins routes) | `warehouse_bins` | `backend/unit_tests/domain/` |
| INV-04 | Inventory restoration on order cancellation/auto-close | -- (backend logic) | `application/services.go` (OrderService -- restore on cancel/auto-close/refund) | `items.quantity` | `backend/unit_tests/domain/` |

---

## 5. Group-Buy Campaigns

> Implementation: `internal/application/campaign_service.go`, `internal/store/postgres/campaign_store.go`, `internal/http/campaign_handler.go`

| ID | Requirement | Frontend Module | Backend Module | Database | Tests |
|----|-------------|-----------------|----------------|----------|-------|
| GRP-01 | Create campaign with item, min quantity, cutoff time, including member item-driven starts | `lib/validation.ts` (createCampaignSchema), `routes/GroupBuysPage.tsx`, `routes/CatalogDetailPage.tsx` | `domain/campaign.go` (GroupBuyCampaign), `application/services.go` (CampaignService.Create), `security/rbac.go`, `http/router.go` | `group_buy_campaigns` | `backend/unit_tests/domain/`, `backend/api_tests/campaign_test.go`, `frontend/unit_tests/routes/GroupBuysPage.test.tsx` |
| GRP-02 | Join campaign (commits quantity, creates linked order) | `lib/validation.ts` (joinCampaignSchema), `routes/GroupBuyDetailPage.tsx` | `domain/campaign.go` (GroupBuyParticipant), `application/services.go` (CampaignService.Join) | `group_buy_participants`, `orders.campaign_id` | `backend/unit_tests/domain/` |
| GRP-03 | Minimum quantity threshold evaluation | -- (backend logic) | `domain/campaign.go` (MeetsThreshold) | `group_buy_campaigns.min_quantity`, `group_buy_campaigns.current_committed_qty` | `backend/unit_tests/domain/` |
| GRP-04 | Cutoff time evaluation (succeeded/failed outcome) | -- (backend logic + cron) | `domain/campaign.go` (Evaluate, IsAtCutoff), `domain/state_machines.go` (ValidCampaignTransition), `jobs/jobs.go` | `group_buy_campaigns.cutoff_time`, `group_buy_campaigns.status`, `group_buy_campaigns.evaluated_at` | `backend/unit_tests/domain/` |
| GRP-05 | Campaign state machine (active -> succeeded/failed/cancelled) | `lib/constants.ts` (CAMPAIGN_STATUSES) | `domain/state_machines.go` (validCampaignTransitions, TransitionCampaign) | `group_buy_campaigns.status` (campaign_status enum) | `backend/unit_tests/domain/` |
| GRP-06 | Cancel active campaign | `routes/GroupBuyDetailPage.tsx` | `application/services.go` (CampaignService.Cancel), `domain/state_machines.go` | `group_buy_campaigns.status` | `backend/unit_tests/domain/` |

---

## 6. Orders

> Implementation: `internal/application/order_service.go`, `internal/store/postgres/order_store.go`, `internal/http/order_handler.go`
>
> Note: FUL-* (fulfillment grouping by supplier, warehouse bin, pickup point) requirements are tracked as deferred to Prompt 7.

| ID | Requirement | Frontend Module | Backend Module | Database | Tests |
|----|-------------|-----------------|----------------|----------|-------|
| ORD-01 | Order state machine (created -> paid -> {cancelled, refunded}; created -> {cancelled, auto_closed}) | `lib/constants.ts` (ORDER_STATUSES), `lib/types.ts` (OrderStatus) | `domain/state_machines.go` (validOrderTransitions, TransitionOrder) | `orders.status` (order_status enum) | `backend/unit_tests/domain/` |
| ORD-02 | Offline settlement marker (not a payment gateway) | `lib/validation.ts` (payOrderSchema.settlement_marker) | `http/dto/requests.go` (PayOrderRequest.SettlementMarker), `application/services.go` (OrderService.Pay) | `orders.settlement_marker` | `backend/unit_tests/domain/` |
| ORD-03 | Auto-close after 30 minutes for unpaid orders | `lib/constants.ts` (AUTO_CLOSE_TIMEOUT_MINUTES) | `domain/order.go` (AutoCloseTimeout, ShouldAutoClose), `jobs/jobs.go` (auto-close job), `application/services.go` (OrderService.AutoCloseExpired) | `orders.auto_close_at`, `orders.status` | `backend/unit_tests/domain/` |
| ORD-04 | Order timeline (chronological event log) | `lib/types.ts` (OrderTimelineEntry) | `domain/order.go` (OrderTimelineEntry), `application/services.go` (OrderService.AddNote), `store/repositories.go` (TimelineRepository) | `order_timeline_entries` | `backend/unit_tests/domain/` |
| ORD-05 | Fulfillment groups (order splitting/grouping) | `lib/types.ts` (FulfillmentGroup) | `domain/order.go` (FulfillmentGroup, FulfillmentGroupOrder), `store/repositories.go` (FulfillmentRepository) | `fulfillment_groups`, `fulfillment_group_orders` | `backend/unit_tests/domain/` |
| ORD-06 | Order split operation | `routes/OrderDetailPage.tsx` | `application/services.go` (OrderService.Split), `http/router.go` (POST /orders/:id/split) | `orders`, `fulfillment_group_orders` | `backend/api_tests/` |
| ORD-07 | Order merge operation | `routes/OrdersPage.tsx` | `application/services.go` (OrderService.Merge), `http/router.go` (POST /orders/merge) | `orders` | `backend/api_tests/` |
| ORD-08 | Order notes | `routes/OrderDetailPage.tsx` | `application/services.go` (OrderService.AddNote), `http/dto/requests.go` (AddOrderNoteRequest) | `order_timeline_entries` | `backend/api_tests/` |
| ORD-09 | Refund flow | `routes/OrderDetailPage.tsx` | `domain/state_machines.go` (paid -> refunded), `application/services.go` (OrderService.Refund) | `orders.status`, `orders.refunded_at` | `backend/unit_tests/domain/` |

---

## 7. Procurement

| ID | Requirement | Frontend Module | Backend Module | Database | Tests |
|----|-------------|-----------------|----------------|----------|-------|
| PRO-01 | Supplier CRUD | `lib/validation.ts` (createSupplierSchema), `routes/SuppliersPage.tsx` | `domain/supplier.go`, `application/services.go` (SupplierService), `store/repositories.go` (SupplierRepository) | `suppliers` | `backend/unit_tests/domain/` |
| PRO-02 | Purchase order lifecycle (create -> approve -> receive -> return/void) | `lib/constants.ts` (PO_STATUSES), `routes/PurchaseOrder*.tsx` | `domain/state_machines.go` (validPOTransitions), `domain/purchase_order.go`, `application/services.go` (PurchaseOrderService) | `purchase_orders` (po_status enum) | `backend/unit_tests/domain/` |
| PRO-03 | PO line items with ordered qty/price | `lib/validation.ts` (createPurchaseOrderSchema), `lib/types.ts` (PurchaseOrderLine) | `domain/purchase_order.go` (PurchaseOrderLine), `http/dto/requests.go` (POLineRequest) | `purchase_order_lines` | `backend/unit_tests/domain/` |
| PRO-04 | Receipt recording with received qty/price | `lib/validation.ts` (receivePurchaseOrderSchema) | `http/dto/requests.go` (ReceivePurchaseOrderRequest), `application/services.go` (PurchaseOrderService.Receive) | `purchase_order_lines.received_quantity`, `purchase_order_lines.received_unit_price` | `backend/api_tests/` |
| PRO-05 | Auto-variance generation on receipt (shortage, overage, price difference) | -- (backend auto-generates) | `domain/variance.go` (VarianceRecord, VarianceType), `domain/enums.go` (VarianceType*) | `variance_records` | `backend/unit_tests/domain/` |
| PRO-06 | Variance escalation (>$250 or >2%) | -- (backend logic) | `domain/variance.go` (RequiresEscalation, VarianceEscalationAmountThreshold, VarianceEscalationPercentThreshold) | `variance_records.difference_amount` | `backend/unit_tests/domain/` |
| PRO-07 | Variance resolution deadline (5 business days) | `lib/constants.ts` (VARIANCE_RESOLUTION_BUSINESS_DAYS) | `domain/variance.go` (CalculateResolutionDueDate, VarianceResolutionBusinessDays) | `variance_records.resolution_due_date` | `backend/unit_tests/domain/` |
| PRO-08 | Variance resolution by procurement staff | `lib/validation.ts` (resolveVarianceSchema) | `application/services.go` (VarianceService.Resolve), `http/dto/requests.go` (ResolveVarianceRequest) | `variance_records.resolved_at`, `variance_records.resolution_notes` | `backend/api_tests/` |
| PRO-09 | Landed cost tracking (value-weighted allocation) | `lib/types.ts` (LandedCostEntry), `routes/LandedCostsPage.tsx` | `domain/variance.go` (LandedCostEntry, CalculateValueWeightedAllocation) | `landed_cost_entries` | `backend/unit_tests/domain/` |
| PRO-10 | PO optimistic concurrency via version | `lib/types.ts` (PurchaseOrder) | `domain/purchase_order.go` (PurchaseOrder.Version) | `purchase_orders.version` | `backend/unit_tests/domain/` |
| PRO-11 | Overdue variance detection | -- (backend job) | `domain/variance.go` (IsOverdue), `jobs/jobs.go` (variance deadline job) | `variance_records.status`, `variance_records.resolution_due_date` | `backend/unit_tests/domain/` |

---

## 8. Security

| ID | Requirement | Frontend Module | Backend Module | Database | Tests |
|----|-------------|-----------------|----------------|----------|-------|
| SEC-01 | Argon2id password hashing with per-user salts | -- (server-side only) | `security/security.go` (Argon2id), `domain/user.go` (PasswordHash, Salt fields) | `users.password_hash`, `users.salt` | `backend/unit_tests/domain/` |
| SEC-02 | 5-failure lockout for 15 minutes | `lib/constants.ts` (LOGIN_LOCKOUT_THRESHOLD, LOGIN_LOCKOUT_DURATION_MINUTES) | `domain/user.go` (IncrementFailedLogin, Lock, IsLocked), `platform/config.go` (LoginLockoutThreshold, LoginLockoutDurationMinutes) | `users.failed_login_count`, `users.locked_until` | `backend/unit_tests/domain/` |
| SEC-03 | Server-side CAPTCHA after lockout | -- (API returns challenge_id) | `domain/user.go` (CaptchaChallenge with `answer_hash`/`answer_salt`), `domain/errors.go` (ErrCaptchaRequired), `security/captcha.go` (crypto RNG + constant-time verification) | `captcha_challenges` | `backend/unit_tests/security/captcha_test.go`, `backend/api_tests/lockout_test.go` |
| SEC-04 | Server-side sessions (30 min idle, 12 hr absolute) | `lib/constants.ts` (SESSION_*) | `domain/user.go` (Session, IsExpired, RefreshIdle), `platform/config.go` (SessionIdleTimeoutMinutes, SessionAbsoluteTimeoutHours) | `sessions` (idle_expires_at, absolute_expires_at) | `backend/unit_tests/domain/` |
| SEC-05 | RBAC (route + service + data scope) | `lib/auth.tsx`, `lib/constants.ts`, `app/routes.tsx` | `http/middleware.go` (AuthMiddleware, RequireRole), `domain/enums.go` (IsStaff, CanAccessDashboard) | `users.role` | `backend/api_tests/` |
| SEC-06 | Parameterized queries (SQL injection prevention) | -- (backend only) | `store/repositories.go` (all use pgx parameterized queries) | -- | `backend/api_tests/` |
| SEC-07 | Output encoding for XSS prevention | React JSX auto-escaping | `http/errors.go` (JSON-only responses) | -- | -- |
| SEC-08 | Sensitive field masking by role | `lib/format.ts` (maskField) | `security/security.go` (data masking utilities) | -- | `frontend/unit_tests/lib/` |
| SEC-09 | Request ID on every request | -- (transparent to frontend) | `http/middleware.go` (RequestIDMiddleware, X-Request-ID header) | -- | `backend/api_tests/` |
| SEC-10 | Panic recovery middleware | -- | `http/middleware.go` (RecoverMiddleware) | -- | `backend/unit_tests/domain/` |

---

## 9. Biometric

| ID | Requirement | Frontend Module | Backend Module | Database | Tests |
|----|-------------|-----------------|----------------|----------|-------|
| BIO-01 | AES-256 envelope encryption for biometric data | `routes/BiometricPage.tsx` | `domain/biometric.go` (BiometricEnrollment.EncryptedData), `security/security.go` (AES-256-GCM) | `biometric_enrollments.encrypted_data` (BYTEA) | `backend/unit_tests/domain/` |
| BIO-02 | 90-day key rotation | `lib/constants.ts` (BIOMETRIC_KEY_ROTATION_DAYS) | `domain/biometric.go` (EncryptionKey.NeedsRotation), `platform/config.go` (BiometricKeyRotationDays), `jobs/procurement_jobs.go` (key rotation check), `cmd/api/main.go` (job scheduling) | `encryption_keys` (activated_at, rotated_at, expires_at, status) | `backend/unit_tests/domain/`, `backend/unit_tests/jobs/jobs_test.go` |
| BIO-03 | Optional module (enable/disable) | -- | `platform/config.go` (BiometricModuleEnabled) | -- | `backend/unit_tests/domain/` |
| BIO-04 | TLS transport security enforced by default with explicit insecure override | -- | `cmd/api/main.go` (StartTLS), `platform/config.go` (TLSCertFile, TLSKeyFile, AllowInsecureHTTP), `platform/tls.go` | -- | `backend/unit_tests/platform/config_test.go` |
| BIO-05 | Key lifecycle (active -> rotated -> revoked) | `routes/BiometricPage.tsx` | `domain/enums.go` (EncryptionKeyStatus*), `domain/biometric.go` (EncryptionKey) | `encryption_keys.status` (encryption_key_status enum) | `backend/unit_tests/domain/` |
| BIO-06 | Biometric data masking in API responses | -- | `security/security.go` (masking) | -- | `backend/unit_tests/domain/` |

---

## 10. Audit & Retention

| ID | Requirement | Frontend Module | Backend Module | Database | Tests |
|----|-------------|-----------------|----------------|----------|-------|
| AUD-01 | Tamper-evident audit log with SHA-256 hash chaining | `routes/AuditPage.tsx` | `domain/audit.go` (AuditEvent.ComputeHash, IntegrityHash, PreviousHash) | `audit_events` (integrity_hash VARCHAR(64), previous_hash VARCHAR(64)) | `backend/unit_tests/domain/` |
| AUD-02 | Append-only audit events (no updates/deletes) | -- (read-only UI) | `store/repositories.go` (AuditRepository.Create, AuditRepository.List, AuditRepository.GetLatestHash -- no Update/Delete) | `audit_events` | `backend/unit_tests/domain/` |
| AUD-03 | Audit event fields: event_type, entity_type, entity_id, actor_id, details (JSONB) | `lib/types.ts` (AuditEvent) | `domain/audit.go` (AuditEvent struct) | `audit_events` (all columns) | `backend/unit_tests/domain/` |
| AUD-04 | 7-year financial record retention | -- | `domain/retention.go` (RetentionFinancialDays derived from 7 years) | `retention_policies` (seeded: financial_records -> 2555 days) | `backend/unit_tests/domain/`, `backend/api_tests/admin_test.go` |
| AUD-05 | 2-year access log retention | -- | `domain/retention.go` (RetentionAccessLogDays derived from 2 years) | `retention_policies` (seeded: access_logs -> 730 days) | `backend/unit_tests/domain/`, `backend/api_tests/admin_test.go` |
| AUD-06 | 7-year audit event retention | -- | `domain/retention.go` | `retention_policies` (seeded: audit_events -> 2555 days) | `backend/unit_tests/domain/`, `backend/api_tests/admin_test.go` |
| AUD-07 | 7-year procurement record retention | -- | `domain/retention.go` | `retention_policies` (seeded: procurement_records -> 2555 days) | `backend/unit_tests/domain/`, `backend/api_tests/admin_test.go` |
| AUD-08 | Retention violation prevention (cannot delete within period) | -- | `domain/retention.go` (IsWithinRetention), `domain/errors.go` (ErrRetentionViolation), `http/errors.go` (RETENTION_VIOLATION) | `retention_policies` | `backend/unit_tests/domain/` |
| AUD-09 | Traceable deletions (audit trail for any removal) | -- | `application/services.go` (AuditService.Log called on destructive actions) | `audit_events` | `backend/api_tests/` |

---

## 11. Backup

| ID | Requirement | Frontend Module | Backend Module | Database | Tests |
|----|-------------|-----------------|----------------|----------|-------|
| BAK-01 | Nightly local backup | `routes/BackupsPage.tsx` (admin backup history view) | `jobs/jobs.go` (backup job), `application/services.go` (BackupService) | `backup_runs` | `backend/unit_tests/domain/` |
| BAK-02 | Encrypted archive (AES-256) | -- | `domain/backup.go` (BackupRun.EncryptionKeyRef), `jobs/jobs.go` | `backup_runs.encryption_key_ref` | `backend/unit_tests/domain/` |
| BAK-03 | SHA-256 checksum verification | -- | `domain/backup.go` (BackupRun.Checksum, ChecksumAlgorithm = "sha256") | `backup_runs.checksum`, `backup_runs.checksum_algorithm` | `backend/unit_tests/domain/` |
| BAK-04 | Admin-configured backup path | -- (FC_BACKUP_PATH env var) | `platform/config.go` (Config.BackupPath), `docker-compose.yml` (FC_BACKUP_PATH) | -- | -- |
| BAK-05 | Backup metadata stored in PostgreSQL | -- | `domain/backup.go` (BackupRun), `store/repositories.go` (BackupRepository) | `backup_runs` | `backend/unit_tests/domain/` |
| BAK-06 | Admin-visible backup history | `routes/BackupsPage.tsx` | `http/router.go` (GET /admin/backups), `application/services.go` (BackupService.List) | `backup_runs` | `backend/api_tests/` |
| BAK-07 | Manual backup trigger | `routes/BackupsPage.tsx` | `http/router.go` (POST /admin/backups), `application/services.go` (BackupService.Trigger) | `backup_runs` | `backend/api_tests/` |
| BAK-08 | Backup status tracking (running -> completed/failed) | `lib/types.ts` (BackupStatus) | `domain/enums.go` (BackupStatus*) | `backup_runs.status` (backup_status enum) | `backend/unit_tests/domain/` |

---

## 12. Offline Operation

| ID | Requirement | Frontend Module | Backend Module | Database | Tests |
|----|-------------|-----------------|----------------|----------|-------|
| OFF-01 | No internet dependency at runtime | `lib/api-client.ts` (relative BASE_URL `/api/v1`), `frontend/Dockerfile` (nginx serves built assets) | All packages (no external HTTP calls) | Local PostgreSQL only | -- |
| OFF-02 | Frontend served via LAN/localhost | `app/providers.tsx` (read-first query cache hydration, `refetchOnWindowFocus: false`, `refetchOnReconnect: false`), `frontend/Dockerfile` (nginx on port 3000), `vite.config.ts` (dev port 5173, proxy to localhost:8080) | `cmd/api/main.go` (binds to 0.0.0.0:8080) | -- | `frontend/unit_tests/components/Layout.test.tsx` |
| OFF-03 | All services run locally via Docker Compose | -- | `docker-compose.yml` (postgres, backend, frontend) | `docker-compose.yml` (pgdata volume) | -- |
| OFF-04 | No CDN or external package registry at runtime | `frontend/Dockerfile` (npm install at build time, nginx serves dist) | `backend/Dockerfile` (go mod download at build time) | -- | -- |
| OFF-05 | Backup to local filesystem (not cloud) | -- | `platform/config.go` (FC_BACKUP_PATH default /var/backups/fitcommerce) | `backup_runs.archive_path` | -- |
| OFF-06 | Cached reads remain available while offline | `lib/offline-cache.ts`, `components/OfflineDataNotice.tsx`, route pages for dashboard/catalog/orders/group-buys/inventory/procurement/reports | -- | IndexedDB client cache + local PostgreSQL origin data | `frontend/unit_tests/routes/CatalogPage.offline.test.tsx` |
| OFF-07 | Offline mode is read-only and mutations require reconnect | `lib/offline.tsx`, mutation-heavy route pages (catalog, campaigns, orders, procurement, reports) | -- | -- | `frontend/unit_tests/routes/GroupBuysPage.test.tsx` |

---

## Prompt 5: React Application Shell

> Implementation: `src/app/`, `src/components/`, `src/lib/hooks/`, `src/routes/LoginPage.tsx`, `src/routes/DashboardPage.tsx`

| ID | Requirement | Frontend Module | Backend Module | Tests |
|----|-------------|-----------------|----------------|-------|
| SHELL-01 | App bootstrap with providers (routing, query cache, auth, theme, notifications, error boundary) | `app/App.tsx`, `app/providers.tsx` (QueryClient, ThemeProvider, AuthProvider, NotificationsProvider, ErrorBoundary) | -- | `unit_tests/auth/protected-route.test.tsx` |
| SHELL-02 | Nested route structure with shared Layout (sidebar + AppBar + Outlet) | `app/routes.tsx` (createBrowserRouter, nested children), `components/Layout.tsx` | -- | `unit_tests/components/Layout.test.tsx` |
| SHELL-03 | Role-aware sidebar navigation | `components/NavSidebar.tsx` (ROLE_PERMISSIONS filter), `lib/constants.ts` | -- | `unit_tests/components/Layout.test.tsx` |
| SHELL-04 | Protected routes with role guards and post-login redirect | `lib/auth.tsx` (ProtectedRoute, withRole), `app/routes.tsx` | -- | `unit_tests/auth/protected-route.test.tsx` |
| SHELL-05 | Session-expired redirect to login | `lib/api-client.ts` (dispatches auth:session-expired on 401), `lib/auth.tsx` (AuthProvider listener) | -- | `unit_tests/auth/session-timeout.test.ts` |
| SHELL-06 | Login page with form validation, lockout display, captcha gate | `routes/LoginPage.tsx` (react-hook-form + Zod), `lib/auth.tsx` | `http/auth_handler.go` | `unit_tests/routes/LoginPage.test.tsx` |
| SHELL-07 | Dashboard shell with period toggles and KPI StatCards | `routes/DashboardPage.tsx` (ToggleButtonGroup, KPIFilters), `lib/hooks/useDashboard.ts` | `http/router.go` (GET /dashboard/kpis stub) | `unit_tests/routes/DashboardPage.test.tsx` |
| SHELL-08 | Permission-aware dashboard sections (RequireRole gates) | `routes/DashboardPage.tsx` (RequireRole per section) | -- | `unit_tests/routes/DashboardPage.test.tsx` |
| SHELL-09 | Reusable DataTable with server-side pagination, loading, empty, error states | `components/DataTable.tsx` | -- | `unit_tests/components/DataTable.test.tsx` |
| SHELL-10 | Reusable FilterBar with text, select, date fields | `components/FilterBar.tsx` | -- | `unit_tests/components/FilterBar.test.tsx` |
| SHELL-11 | Shared UI primitives: StatCard, PageContainer, EmptyState, ConfirmDialog, ErrorBoundary | `components/StatCard.tsx`, `components/PageContainer.tsx`, `components/EmptyState.tsx`, `components/ConfirmDialog.tsx`, `components/ErrorBoundary.tsx` | -- | (integrated in page tests) |
| SHELL-12 | Toast notification system (success/error/warning/info) | `lib/notifications.tsx` (NotificationsProvider, useNotify) | -- | (integrated in mutation hook usage) |
| SHELL-13 | Data-fetching hooks for items, orders, campaigns, inventory, dashboard | `lib/hooks/useItems.ts`, `lib/hooks/useOrders.ts`, `lib/hooks/useCampaigns.ts`, `lib/hooks/useInventory.ts`, `lib/hooks/useDashboard.ts` | All Prompt 4 API endpoints | -- |
| SHELL-14 | Dashboard filter bar gated behind RequireRole | `routes/DashboardPage.tsx` (FilterBar inside RequireRole) | -- | `unit_tests/routes/DashboardPage.test.tsx` |

## Prompt 6: Primary Operational Screens

| ID | Requirement | Implementation | API Endpoint | Tests |
|----|-------------|----------------|-------------|-------|
| UI-01 | Item listing with filter and pagination | `routes/CatalogPage.tsx` | `GET /items` | `unit_tests/routes/CatalogPage.test.tsx` |
| UI-02 | Item detail view with status and publish/unpublish | `routes/CatalogDetailPage.tsx` | `GET /items/:id`, `POST /items/:id/publish`, `POST /items/:id/unpublish` | -- |
| UI-03 | Item create/edit form with validation and conflict handling | `routes/CatalogFormPage.tsx` | `POST /items`, `PUT /items/:id` | `unit_tests/routes/CatalogFormPage.test.tsx` |
| UI-04 | StatusChip for item/order/campaign/PO statuses | `components/StatusChip.tsx` | -- | -- |
| UI-05 | Inventory snapshots and adjustments view | `routes/InventoryPage.tsx` | `GET /inventory/snapshots`, `GET /inventory/adjustments`, `POST /inventory/adjustments` | `unit_tests/routes/InventoryPage.test.tsx`, `backend/api_tests/inventory_test.go` |
| UI-06 | Group buy campaign listing with progress bar | `routes/GroupBuysPage.tsx` | `GET /campaigns`, `POST /campaigns` | -- |
| UI-07 | Group buy detail with join form (member) and staff actions | `routes/GroupBuyDetailPage.tsx` | `GET /campaigns/:id`, `POST /campaigns/:id/join`, `POST /campaigns/:id/cancel`, `POST /campaigns/:id/evaluate` | `unit_tests/routes/GroupBuyDetailPage.test.tsx` |
| UI-08 | Order listing with status filter | `routes/OrdersPage.tsx` | `GET /orders` | -- |
| UI-09 | Order detail with timeline and role-aware actions | `routes/OrderDetailPage.tsx` | `GET /orders/:id`, `GET /orders/:id/timeline`, `POST /orders/:id/cancel`, `POST /orders/:id/pay`, `POST /orders/:id/refund`, `POST /orders/:id/notes` | `unit_tests/routes/OrderDetailPage.test.tsx` |
| UI-10 | Reports page with role-gated export buttons | `routes/ReportsPage.tsx` | `GET /reports`, `POST /exports` | `unit_tests/routes/ReportsPage.test.tsx` |
| UI-11 | useCreateItem and useCreateCampaign mutations | `lib/hooks/useItems.ts`, `lib/hooks/useCampaigns.ts` | `POST /items`, `POST /campaigns` | -- |

## Prompt 7: Procurement Closure, Reporting/Export, Audit/Admin, Biometrics, Backup

| ID | Requirement | Implementation | Tests |
|----|-------------|----------------|-------|
| PRO-EXT-01 | Supplier management (create, list, get, update) | `application/supplier_service.go`, `http/supplier_handler.go`, `routes/SuppliersPage.tsx` | `backend/unit_tests/procurement/`, `unit_tests/routes/SuppliersPage.test.tsx` |
| PRO-EXT-02 | PO lifecycle (Created→Approved→Received→Returned/Voided) | `application/procurement_service.go`, `http/procurement_handler.go` | `backend/unit_tests/procurement/procurement_service_test.go`, `backend/api_tests/procurement_test.go` |
| PRO-EXT-03 | Variance detection and resolution | `application/variance_service.go`, `domain/variance.go` | `backend/unit_tests/procurement/variance_service_test.go`, `unit_tests/routes/VariancesPage.test.tsx` |
| PRO-EXT-04 | Landed-cost rollup | `internal/store/postgres/variance_store.go` (LandedCostStore) | `backend/unit_tests/domain/landed_cost_test.go` |
| RPT-EXT-01 | Predefined report definitions with role-aware list | `application/report_service.go` | `backend/unit_tests/reports/report_service_test.go` |
| RPT-EXT-02 | CSV/PDF export generation with masking | `application/report_service.go`, `security/masking.go` | `backend/unit_tests/reports/report_service_test.go` |
| RPT-EXT-03 | Export job lifecycle and download | `http/report_handler.go` | `backend/api_tests/report_export_test.go` |
| RPT-EXT-04 | Export filename format (report_type_YYYYMMDD_HHmmss.ext) | `domain/report.go` (GenerateExportFilename) | `backend/unit_tests/domain/report_test.go` |
| ADM-EXT-01 | Audit log (paginated, security-event filter) | `application/audit_service.go`, `http/admin_handler.go` | `backend/api_tests/admin_test.go` |
| ADM-EXT-02 | User management with masking | `application/user_service.go`, `http/user_handler.go` | `backend/unit_tests/admin/user_service_test.go`, `unit_tests/routes/UsersPage.test.tsx` |
| BAK-EXT-01 | Backup trigger (manual and scheduled) | `application/backup_service.go`, `jobs/procurement_jobs.go` | `backend/unit_tests/backup/backup_service_test.go`, `backend/unit_tests/jobs/jobs_test.go` |
| BAK-EXT-02 | AES-256-GCM encryption + HKDF key derivation | `security/crypto.go` (DeriveKeyFromRef, EncryptAESGCM) | `backend/unit_tests/security/crypto_test.go` |
| RET-EXT-01 | Retention policy administration (hard-delete cleanup by retention window) | `application/retention_service.go`, `jobs/procurement_jobs.go` | `backend/unit_tests/jobs/jobs_test.go` |
| BIO-EXT-01 | Biometric enrollment, revocation, key rotation | `application/biometric_service.go`, `http/biometric_handler.go` | `backend/unit_tests/security/crypto_test.go` |
| CFG-EXT-01 | Platform config loading from FC_ env vars with defaults | `platform/config.go` | `backend/unit_tests/platform/config_test.go` |

## Prompt 8: Test Suite Hardening

| ID | Requirement | New Tests Added |
|----|-------------|-----------------|
| TST-01 | Frontend: item-form validation (create/edit) | `unit_tests/routes/CatalogFormPage.test.tsx` |
| TST-02 | Frontend: inventory page tabs and adjustment dialog | `unit_tests/routes/InventoryPage.test.tsx` |
| TST-03 | Frontend: variance list, overdue indicator, resolve dialog | `unit_tests/routes/VariancesPage.test.tsx` |
| TST-04 | Frontend: supplier list, create dialog, RBAC gates | `unit_tests/routes/SuppliersPage.test.tsx` |
| TST-05 | Frontend: reports page export buttons and job tracking | `unit_tests/routes/ReportsPage.test.tsx` |
| TST-06 | Frontend: admin page heading | `unit_tests/routes/AdminPage.test.tsx` |
| TST-07 | Backend: config defaults, env overrides, invalid-value fallback | `backend/unit_tests/platform/config_test.go` |
| TST-08 | Backend: export filename format and UTC normalization | `backend/unit_tests/domain/report_test.go` |
| TST-09 | Backend API: inventory snapshot/adjustment auth and RBAC | `backend/api_tests/inventory_test.go` |
| TST-10 | Frontend coverage config | `vite.config.ts` (coverage.provider=v8), `package.json` (@vitest/coverage-v8) |
| TST-11 | run_tests.sh coverage flags | `run_tests.sh` (-coverprofile, go tool cover, --coverage) |

## Prompt 9: Docker and Config Hardening

| ID | Requirement | Implementation |
|----|-------------|----------------|
| DCK-01 | Frontend Dockerfile: COPY nginx.conf instead of heredoc | `frontend/Dockerfile`, `frontend/nginx.conf` |
| DCK-02 | Nginx config in separate file for portability | `frontend/nginx.conf` |
| DCK-03 | Port consistency: frontend 3000, backend 8080, postgres 5432 | `docker-compose.yml`, `backend/Dockerfile`, `frontend/Dockerfile`, `frontend/nginx.conf`, `vite.config.ts` |
| DCK-04 | All FC_ env vars documented in README and docker-compose.yml | `repo/README.md`, `docker-compose.yml` |
| DCK-05 | Backend/frontend images built from public registries only | `backend/Dockerfile` (golang:1.22-alpine, alpine:3.20), `frontend/Dockerfile` (node:20-alpine, nginx:1.27-alpine) |
| DCK-06 | Export and backup volumes declared in docker-compose.yml | `docker-compose.yml` (backups, exports named volumes) |

## Prompt 10: Static Readiness Audit

| ID | Finding / Requirement | Action Taken | Files Changed |
|----|----------------------|--------------|---------------|
| AUD-01 | Inventory snapshot RBAC granted access to all roles (incl. Coach/Member) via `ActionViewCatalog` | Corrected to `NewRequireRole(ActionManageInventory, ActionManageProcurement)` — operational roles only | `backend/internal/http/router.go`, `backend/api_tests/inventory_test.go` |
| AUD-02 | Location/Coach/Member stub routes lacked authentication middleware | Added `authMW` to all three stub route registration functions | `backend/internal/http/router.go` |
| AUD-03 | Dashboard endpoint name inconsistent across docs (`/dashboard/summary` vs actual `/dashboard/kpis`) | Corrected all references to `GET /dashboard/kpis` | `questions.md`, `docs/requirement-traceability.md` |
| AUD-04 | pg_dump placeholder undocumented externally (creates empty file, no real DB data) | Added honest note to README.md, docs/design.md, and questions.md | `repo/README.md`, `docs/design.md`, `questions.md` |
| AUD-05 | Session expiry enforcement verified (static review) | `session.go`: HttpOnly, SameSite=Strict, MaxAge from absolute expiry. Session validation in AuthService enforces idle + absolute timeout via DB fields. No fix needed. | -- |
| AUD-06 | Audit hash chain verified (static review) | `audit_helper.go`: `BuildAuditEvent` sets `IntegrityHash = event.ComputeHash(previousHash)`. `RedactSensitiveFields` strips password_hash, salt, token, key, encrypted_data. No fix needed. | -- |
| AUD-07 | PII masking verified (static review) | `masking.go`: `MaskFieldByRole` and `ApplyUserResponseMask` scope email/phone by role and ownership. `RedactBiometric` returns constant placeholder. No fix needed. | -- |
| AUD-08 | Backup encryption flow verified (static review) | `backup_service.go`: `DeriveKeyFromRef` + `EncryptAESGCM` called when `FC_BACKUP_ENCRYPTION_KEY_REF` is set. AES-256-GCM nonce prepended. SHA-256 checksum captured before encryption. No fix needed. | -- |
| AUD-09 | RBAC matrix verified (static review) | `rbac.go` `rolePermissions` map matches documented matrix. `NewRequireRole` uses OR semantics. Admin routes use `NewRequireRole(ActionX)` — no raw role comparisons. No fix needed. | -- |
| AUD-10 | FC_ config params — ExportPath default mismatch documented | `config.go` default `/tmp/fitcommerce-exports` is correct for local dev; docker-compose overrides to named volume. README updated to document both values. | `repo/README.md`, `questions.md` |
| AUD-11 | Inventory snapshot API spec roles updated to reflect RBAC fix | Added `procurement_specialist` to snapshot required roles in api-spec.md | `docs/api-spec.md` |
| AUD-12 | Retention cleanup job documented as log-only | Design doc updated: no DELETE issued; schema lacks `deleted_at`; eligible records are logged. | `docs/design.md` |
| AUD-13 | New inventory API tests for corrected RBAC | `TestInventorySnapshots_ProcurementSpecialist_200`, `TestInventorySnapshots_Coach_403` added | `backend/api_tests/inventory_test.go` |
