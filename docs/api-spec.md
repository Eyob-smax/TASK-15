# FitCommerce Operations & Inventory Suite -- REST API Specification

## Conventions

### Base URL

```
/api/v1/
```

All endpoints are prefixed with `/api/v1/`. In Docker, the frontend nginx proxies `/api/*` to the backend on port 8080.

### Authentication

Session-based authentication using HTTP cookies.

- `POST /api/v1/auth/login` returns a session cookie.
- All subsequent requests include the cookie via `credentials: 'include'`.
- The `AuthMiddleware` validates the session on every request to a protected endpoint.
- Sessions have a 30-minute idle timeout and a 12-hour absolute timeout.

### Error Envelope

All error responses use the following structure:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable description",
    "details": [
      {
        "field": "field_name",
        "message": "Field-level error description"
      }
    ]
  }
}
```

The `details` array is omitted when there are no field-level details.

### Standard Error Codes

| Code | HTTP Status | Description |
|---|---|---|
| `UNAUTHORIZED` | 401 | Not authenticated -- session missing or expired |
| `FORBIDDEN` | 403 | Authenticated but not authorized for this action |
| `NOT_FOUND` | 404 | Requested entity does not exist |
| `CONFLICT` | 409 | Version conflict or duplicate entity |
| `VALIDATION_ERROR` | 422 | Request body failed validation |
| `ACCOUNT_LOCKED` | 423 | Account locked due to failed login attempts |
| `CAPTCHA_REQUIRED` | 403 | CAPTCHA must be solved before proceeding |
| `INVALID_TRANSITION` | 422 | State machine transition is not allowed |
| `PUBLISH_BLOCKED` | 422 | Item cannot be published (validation failures) |
| `BATCH_PARTIAL_FAILURE` | 207 | Some rows in batch edit failed |
| `VARIANCE_UNRESOLVED` | 422 | Variance record has not been resolved |
| `RETENTION_VIOLATION` | 403 | Entity cannot be deleted within retention period |
| `INTERNAL_ERROR` | 500 | Unexpected server error |
| `NOT_IMPLEMENTED` | 501 | Endpoint stub not yet implemented |

### Success Envelope

Single-entity responses:

```json
{
  "data": { ... }
}
```

List/paginated responses:

```json
{
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_count": 150,
    "total_pages": 8
  }
}
```

### Pagination

Query parameters for list endpoints:

| Parameter | Type | Default | Max | Description |
|---|---|---|---|---|
| `page` | integer | 1 | -- | Page number (1-indexed) |
| `page_size` | integer | 20 | 100 | Items per page |

### Sorting

| Parameter | Type | Default | Description |
|---|---|---|---|
| `sort` | string | varies by endpoint | Field name to sort by |
| `order` | string | `asc` | Sort direction: `asc` or `desc` |

### Filtering

Each list endpoint documents its specific filter parameters below.

### Optimistic Concurrency

Entities with a `version` field (items, purchase orders) support optimistic concurrency control:

- The current `version` is returned in GET responses.
- PUT/update requests must include the `version` field in the request body.
- The server compares the submitted version with the stored version.
- If they differ, a `CONFLICT` (409) error is returned.
- On successful update, the version is incremented.

### Export Downloads

Export file download endpoints return the file directly with:

- `Content-Disposition: attachment; filename="report_type_YYYYMMDD_HHmmss.csv"`
- Appropriate `Content-Type` header (`text/csv` or `application/pdf`)

---

## Endpoint Reference

---

### Auth & Sessions

#### POST /auth/login

Authenticate a user and create a new session.

**Required Role**: Public (no authentication required)

**Request Body**:
```json
{
  "email": "admin@fitcommerce.local",
  "password": "securepassword"
}
```

**Success Response** (200):
```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "admin@fitcommerce.local",
      "role": "administrator",
      "status": "active",
      "display_name": "Admin User",
      "location_id": "uuid | null",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    },
    "session": {
      "idle_expires_at": "2026-04-09T10:30:00Z",
      "absolute_expires_at": "2026-04-09T22:00:00Z"
    }
  }
}
```

A `Set-Cookie` header is returned to establish the session cookie. The session token is intentionally cookie-only and never appears in the JSON response body.

**Error Codes**: `UNAUTHORIZED`, `ACCOUNT_LOCKED`, `CAPTCHA_REQUIRED`, `VALIDATION_ERROR`

---

#### POST /auth/logout

Destroy the current session.

**Required Role**: Any authenticated user

**Request Body**: Empty (`{}`)

**Success Response** (200):
```json
{
  "data": {
    "message": "logged out"
  }
}
```

**Error Codes**: `UNAUTHORIZED`

---

#### POST /auth/captcha/verify

Verify a CAPTCHA challenge answer during locked-account recovery.

**Required Role**: Public

**Request Body**:
```json
{
  "challenge_id": "uuid",
  "answer": "42"
}
```

**Success Response** (200):
```json
{
  "data": {
    "message": "captcha verified"
  }
}
```

Missing, expired, or already-verified challenges return `UNAUTHORIZED` to avoid challenge enumeration.

**Error Codes**: `VALIDATION_ERROR`, `UNAUTHORIZED`

---

#### GET /auth/session

Retrieve the current session and user information.

**Required Role**: Any authenticated user

**Success Response** (200):
```json
{
  "data": {
    "user": { "...UserResponse" },
    "session": { "...SessionResponse" }
  }
}
```

**Error Codes**: `UNAUTHORIZED`

---

### Dashboard KPIs

#### GET /dashboard/kpis

Retrieve aggregated KPI metrics for the operations dashboard.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`, `coach`

**Query Parameters**:

| Parameter | Type | Description |
|---|---|---|
| `period` | string | Time period: `daily`, `weekly`, `monthly`, `quarterly`, `yearly` |
| `location_id` | uuid | Filter by location ID |
| `coach_id` | uuid | Filter by coach ID |
| `category` | string | Filter by item category |
| `from` | date | Start date (YYYY-MM-DD) |
| `to` | date | End date (YYYY-MM-DD) |

**Success Response** (200):
```json
{
  "data": {
    "member_growth": {
      "value": 120.0,
      "previous_value": 100.0,
      "change_percent": 20.0,
      "period": "monthly"
    },
    "churn": {
      "value": 5.2,
      "previous_value": 6.1,
      "change_percent": -14.75,
      "period": "monthly"
    },
    "renewal_rate": {
      "value": 87.5,
      "previous_value": 85.0,
      "change_percent": 2.94,
      "period": "monthly"
    },
    "engagement": {
      "value": 72.3,
      "previous_value": 68.0,
      "change_percent": 6.32,
      "period": "monthly"
    },
    "class_fill_rate": {
      "value": 81.0,
      "previous_value": 78.5,
      "change_percent": 3.18,
      "period": "monthly"
    },
    "coach_productivity": {
      "value": 15.4,
      "previous_value": 14.2,
      "change_percent": 8.45,
      "period": "monthly"
    }
  }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`

---

### Catalog / Items

#### POST /items

Create a new catalog item (created in `draft` status).

**Required Role**: `administrator`, `operations_manager`

**Request Body**:
```json
{
  "name": "Treadmill Pro 5000",
  "description": "Commercial-grade treadmill",
  "category": "cardio",
  "brand": "ProFit",
  "condition": "new",
  "billing_model": "monthly_rental",
  "refundable_deposit": 50.00,
  "quantity": 10,
  "location_id": "uuid",
  "availability_windows": [
    { "start_time": "2026-04-10T06:00:00Z", "end_time": "2026-04-10T22:00:00Z" }
  ],
  "blackout_windows": [
    { "start_time": "2026-04-15T00:00:00Z", "end_time": "2026-04-16T00:00:00Z" }
  ]
}
```

**Success Response** (201):
```json
{
  "data": { "...ItemResponse" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `VALIDATION_ERROR`

---

#### GET /items

List catalog items with filtering, sorting, and pagination.

**Required Role**: Any authenticated user (member sees published only; staff sees all)

**Query Parameters**:

| Parameter | Type | Description |
|---|---|---|
| `page` | integer | Page number |
| `page_size` | integer | Items per page |
| `sort` | string | Sort field: `name`, `category`, `brand`, `created_at`, `updated_at` |
| `order` | string | `asc` or `desc` |
| `status` | string | Filter by item status: `draft`, `published`, `unpublished` |
| `category` | string | Filter by category |
| `brand` | string | Filter by brand |
| `condition` | string | Filter by condition: `new`, `open_box`, `used` |
| `billing_model` | string | Filter by billing model: `one_time`, `monthly_rental` |
| `location_id` | uuid | Filter by location |
| `search` | string | Free-text search across name and description |

**Success Response** (200):
```json
{
  "data": [ { "...ItemResponse" } ],
  "pagination": { "page": 1, "page_size": 20, "total_count": 50, "total_pages": 3 }
}
```

**Error Codes**: `UNAUTHORIZED`

---

#### GET /items/:id

Get a single item by ID.

**Required Role**: Any authenticated user

**Success Response** (200):
```json
{
  "data": {
    "id": "uuid",
    "name": "Treadmill Pro 5000",
    "description": "Commercial-grade treadmill",
    "category": "cardio",
    "brand": "ProFit",
    "condition": "new",
    "refundable_deposit": 50.00,
    "billing_model": "monthly_rental",
    "status": "draft",
    "quantity": 10,
    "location_id": "uuid",
    "created_by": "uuid",
    "created_at": "2026-04-01T00:00:00Z",
    "updated_at": "2026-04-01T00:00:00Z",
    "version": 1,
    "availability_windows": [
      { "id": "uuid", "start_time": "2026-04-10T06:00:00Z", "end_time": "2026-04-10T22:00:00Z" }
    ],
    "blackout_windows": [
      { "id": "uuid", "start_time": "2026-04-15T00:00:00Z", "end_time": "2026-04-16T00:00:00Z" }
    ]
  }
}
```

**Error Codes**: `UNAUTHORIZED`, `NOT_FOUND`

---

#### PUT /items/:id

Update a catalog item. Requires optimistic concurrency via the `version` field.

**Required Role**: `administrator`, `operations_manager`

**Request Body**:
```json
{
  "name": "Treadmill Pro 5000 v2",
  "description": "Updated description",
  "category": "cardio",
  "brand": "ProFit",
  "condition": "new",
  "refundable_deposit": 75.00,
  "billing_model": "monthly_rental",
  "quantity": 12,
  "location_id": "uuid",
  "version": 1,
  "availability_windows": [],
  "blackout_windows": []
}
```

**Success Response** (200):
```json
{
  "data": { "...ItemResponse (version incremented)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `VALIDATION_ERROR`

---

#### POST /items/:id/publish

Publish a draft item, making it visible to members. Validates all required fields, checks for window overlaps.

**Required Role**: `administrator`, `operations_manager`

**Request Body**: Empty

**Success Response** (200):
```json
{
  "data": { "...ItemResponse (status: published)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `PUBLISH_BLOCKED`

---

#### POST /items/:id/unpublish

Unpublish a published item, removing it from member visibility.

**Required Role**: `administrator`, `operations_manager`

**Request Body**: Empty

**Success Response** (200):
```json
{
  "data": { "...ItemResponse (status: unpublished)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `INVALID_TRANSITION`

---

#### POST /items/batch-edit

Apply batch edits to multiple items in a single operation.

**Required Role**: `administrator`, `operations_manager`

**Request Body**:
```json
{
  "edits": [
    { "item_id": "uuid", "field": "refundable_deposit", "new_value": "75.00" },
    {
      "item_id": "uuid",
      "field": "availability_windows",
      "availability_windows": [
        { "start_time": "2026-04-10T06:00:00Z", "end_time": "2026-04-10T22:00:00Z" }
      ]
    }
  ]
}
```

Supported `field` values include scalar item fields plus `availability_windows`. When `field` is `availability_windows`, the submitted array replaces the item's current availability windows; sending an empty array clears them. Each row is validated independently and publish-time overlap rules still apply.

**Success Response** (200 or 207 for partial failure):
```json
{
  "data": {
    "job_id": "uuid",
    "total_rows": 2,
    "success_count": 1,
    "failure_count": 1,
    "results": [
      { "item_id": "uuid", "field": "refundable_deposit", "old_value": "50.00", "new_value": "75.00", "success": true },
      { "item_id": "uuid", "field": "category", "old_value": "cardio", "new_value": "strength", "success": false, "failure_reason": "item is published" }
    ]
  }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `VALIDATION_ERROR`, `BATCH_PARTIAL_FAILURE`

---

### Inventory

#### GET /inventory/snapshots

List inventory snapshots.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`

**Query Parameters**:

| Parameter | Type | Description |
|---|---|---|
| `item_id` | uuid | Filter by item |
| `location_id` | uuid | Filter by location |

**Success Response** (200):
```json
{
  "data": [
    {
      "id": "uuid",
      "item_id": "uuid",
      "quantity": 10,
      "location_id": "uuid",
      "recorded_at": "2026-04-09T00:00:00Z"
    }
  ]
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`

---

#### POST /inventory/adjustments

Create a manual inventory adjustment.

**Required Role**: `administrator`, `operations_manager`

**Request Body**:
```json
{
  "item_id": "uuid",
  "quantity_change": -2,
  "reason": "Damaged during delivery"
}
```

**Success Response** (201):
```json
{
  "data": {
    "id": "uuid",
    "item_id": "uuid",
    "quantity_change": -2,
    "reason": "Damaged during delivery",
    "created_by": "uuid",
    "created_at": "2026-04-09T10:00:00Z"
  }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `VALIDATION_ERROR`

---

#### GET /inventory/adjustments

List inventory adjustments.

**Required Role**: `administrator`, `operations_manager`

**Query Parameters**:

| Parameter | Type | Description |
|---|---|---|
| `item_id` | uuid | Filter by item |
| `page` | integer | Page number |
| `page_size` | integer | Items per page |

**Success Response** (200):
```json
{
  "data": [ { "...InventoryAdjustmentResponse" } ],
  "pagination": { "..." }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`

---

#### POST /warehouse-bins

Create a warehouse bin.

**Required Role**: `administrator`, `operations_manager`

**Request Body**:
```json
{
  "location_id": "uuid",
  "name": "BIN-A-01",
  "description": "Aisle A, Rack 1"
}
```

**Success Response** (201):
```json
{
  "data": { "...WarehouseBinResponse" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `VALIDATION_ERROR`, `CONFLICT`

---

#### GET /warehouse-bins

List warehouse bins.

**Required Role**: `administrator`, `operations_manager`

**Query Parameters**: `location_id`, `page`, `page_size`

**Success Response** (200): Paginated list of `WarehouseBinResponse`

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`

---

### Group-Buy Campaigns

#### POST /campaigns

Create a new group-buy campaign.

**Required Role**: `administrator`, `operations_manager`, or `member` when starting a campaign from a published item context

**Request Body**:
```json
{
  "item_id": "uuid",
  "min_quantity": 10,
  "cutoff_time": "2026-04-20T18:00:00Z"
}
```

Staff can open the generic create flow from the campaigns page. Members can only start a campaign from an item-driven flow tied to a published catalog item; the frontend locks the `item_id` for that path.

**Success Response** (201):
```json
{
  "data": {
    "id": "uuid",
    "item_id": "uuid",
    "min_quantity": 10,
    "current_committed_qty": 0,
    "cutoff_time": "2026-04-20T18:00:00Z",
    "status": "active",
    "created_by": "uuid",
    "created_at": "2026-04-09T10:00:00Z",
    "evaluated_at": null
  }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `VALIDATION_ERROR`

---

#### GET /campaigns

List campaigns with pagination.

**Required Role**: Any authenticated user

**Query Parameters**: `page`, `page_size`, `status`

**Success Response** (200): Paginated list of `CampaignResponse`

---

#### GET /campaigns/:id

Get a single campaign with participants.

**Required Role**: Any authenticated user

**Success Response** (200):
```json
{
  "data": { "...CampaignResponse" }
}
```

**Error Codes**: `NOT_FOUND`

---

#### POST /campaigns/:id/join

Join a group-buy campaign (creates an order linked to the campaign).

**Required Role**: Any authenticated user

**Request Body**:
```json
{
  "quantity": 2
}
```

**Success Response** (201):
```json
{
  "data": { "...ParticipantResponse" }
}
```

**Error Codes**: `UNAUTHORIZED`, `NOT_FOUND`, `VALIDATION_ERROR`, `INVALID_TRANSITION` (campaign not active)

---

#### POST /campaigns/:id/cancel

Cancel an active campaign.

**Required Role**: `administrator`, `operations_manager`

**Request Body**: Empty

**Success Response** (200):
```json
{
  "data": { "...CampaignResponse (status: cancelled)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `INVALID_TRANSITION`

---

#### POST /campaigns/:id/evaluate

Manually trigger campaign evaluation at/after cutoff time.

**Required Role**: `administrator`, `operations_manager`

**Request Body**: Empty

**Success Response** (200):
```json
{
  "data": { "...CampaignResponse (status: succeeded|failed)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `INVALID_TRANSITION`

---

### Orders

#### POST /orders

Create a new order.

**Required Role**: Any authenticated user

**Request Body**:
```json
{
  "item_id": "uuid",
  "campaign_id": "uuid (optional, null for standalone orders)",
  "quantity": 1
}
```

**Success Response** (201):
```json
{
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "item_id": "uuid",
    "campaign_id": null,
    "quantity": 1,
    "unit_price": 50.00,
    "total_amount": 50.00,
    "status": "created",
    "settlement_marker": "",
    "notes": "",
    "auto_close_at": "2026-04-09T10:30:00Z",
    "created_at": "2026-04-09T10:00:00Z",
    "updated_at": "2026-04-09T10:00:00Z",
    "paid_at": null,
    "cancelled_at": null,
    "refunded_at": null
  }
}
```

**Error Codes**: `UNAUTHORIZED`, `VALIDATION_ERROR`, `NOT_FOUND` (item)

---

#### GET /orders

List orders with filtering and pagination.

**Required Role**: Any authenticated user (members see own orders only)

**Query Parameters**:

| Parameter | Type | Description |
|---|---|---|
| `page` | integer | Page number |
| `page_size` | integer | Items per page |
| `status` | string | Filter by order status |
| `user_id` | uuid | Filter by user (staff only) |
| `item_id` | uuid | Filter by item |
| `sort` | string | Sort field |
| `order` | string | `asc` or `desc` |

**Success Response** (200): Paginated list of `OrderResponse`

**Error Codes**: `UNAUTHORIZED`

---

#### GET /orders/:id

Get order details.

**Required Role**: Any authenticated user (members can only access own orders)

**Success Response** (200):
```json
{
  "data": { "...OrderResponse" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`

---

#### POST /orders/:id/pay

Mark an order as paid with an offline settlement marker.

**Required Role**: `administrator`, `operations_manager`

**Request Body**:
```json
{
  "settlement_marker": "CASH-RECEIPT-2026-0409-001"
}
```

**Success Response** (200):
```json
{
  "data": { "...OrderResponse (status: paid)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `INVALID_TRANSITION`

---

#### POST /orders/:id/cancel

Cancel an order.

**Required Role**: `administrator`, `operations_manager` (or the order owner for `created` orders)

**Request Body**: Empty

**Success Response** (200):
```json
{
  "data": { "...OrderResponse (status: cancelled)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `INVALID_TRANSITION`

---

#### POST /orders/:id/refund

Refund a paid order.

**Required Role**: `administrator`, `operations_manager`

**Request Body**: Empty

**Success Response** (200):
```json
{
  "data": { "...OrderResponse (status: refunded)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `INVALID_TRANSITION`

---

#### POST /orders/:id/notes

Add a note to the order's timeline.

**Required Role**: `administrator`, `operations_manager`

**Request Body**:
```json
{
  "note": "Customer contacted about delay"
}
```

**Success Response** (201):
```json
{
  "data": { "...TimelineEntryResponse" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `VALIDATION_ERROR`

---

#### GET /orders/:id/timeline

Get the chronological timeline of events for an order.

**Required Role**: Any authenticated user (members can only access own orders)

**Success Response** (200):
```json
{
  "data": [
    {
      "id": "uuid",
      "order_id": "uuid",
      "action": "created",
      "description": "Order created",
      "performed_by": "uuid",
      "created_at": "2026-04-09T10:00:00Z"
    }
  ]
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`

---

#### POST /orders/:id/split

Split an order into multiple orders with specified quantities.

**Required Role**: `administrator`, `operations_manager`

**Request Body**:
```json
{
  "quantities": [3, 2]
}
```

**Success Response** (200):
```json
{
  "data": [ { "...OrderResponse" }, { "...OrderResponse" } ]
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `VALIDATION_ERROR`

---

#### POST /orders/merge

Merge multiple orders for the same item into a single order.

**Required Role**: `administrator`, `operations_manager`

**Request Body**:
```json
{
  "order_ids": ["uuid", "uuid"]
}
```

**Success Response** (200):
```json
{
  "data": { "...OrderResponse (merged)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `VALIDATION_ERROR`

---

### Procurement

#### POST /suppliers

Create a new supplier.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`

**Request Body**:
```json
{
  "name": "FitEquip Global",
  "contact_name": "Jane Doe",
  "contact_email": "jane@fitequip.com",
  "contact_phone": "+1-555-0100",
  "address": "123 Equipment Lane"
}
```

**Success Response** (201):
```json
{
  "data": { "...SupplierResponse" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `VALIDATION_ERROR`, `CONFLICT`

---

#### GET /suppliers

List suppliers with pagination.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`

**Query Parameters**: `page`, `page_size`, `is_active`

**Success Response** (200): Paginated list of `SupplierResponse`

---

#### GET /suppliers/:id

Get a single supplier.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`

**Success Response** (200):
```json
{
  "data": { "...SupplierResponse" }
}
```

**Error Codes**: `NOT_FOUND`

---

#### PUT /suppliers/:id

Update a supplier's details.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`

**Request Body**:
```json
{
  "name": "FitEquip Global Inc.",
  "contact_name": "Jane Doe",
  "contact_email": "jane@fitequip.com",
  "contact_phone": "+1-555-0100",
  "address": "456 Equipment Blvd"
}
```

**Success Response** (200):
```json
{
  "data": { "...SupplierResponse" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `VALIDATION_ERROR`, `CONFLICT`

---

#### POST /purchase-orders

Create a new purchase order.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`

**Request Body**:
```json
{
  "supplier_id": "uuid",
  "lines": [
    {
      "item_id": "uuid",
      "ordered_quantity": 20,
      "ordered_unit_price": 450.00
    }
  ]
}
```

**Success Response** (201):
```json
{
  "data": { "...PurchaseOrderResponse (status: created, with lines)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `VALIDATION_ERROR`, `NOT_FOUND` (supplier/item)

---

#### GET /purchase-orders

List purchase orders with pagination.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`

**Query Parameters**: `page`, `page_size`, `status`, `supplier_id`

**Success Response** (200): Paginated list of `PurchaseOrderResponse`

---

#### GET /purchase-orders/:id

Get a single purchase order with lines.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`

**Success Response** (200):
```json
{
  "data": {
    "id": "uuid",
    "supplier_id": "uuid",
    "status": "created",
    "total_amount": 9000.00,
    "created_by": "uuid",
    "approved_by": null,
    "created_at": "2026-04-01T00:00:00Z",
    "approved_at": null,
    "received_at": null,
    "version": 1,
    "lines": [
      {
        "id": "uuid",
        "purchase_order_id": "uuid",
        "item_id": "uuid",
        "ordered_quantity": 20,
        "ordered_unit_price": 450.00,
        "received_quantity": null,
        "received_unit_price": null
      }
    ]
  }
}
```

**Error Codes**: `NOT_FOUND`

---

#### POST /purchase-orders/:id/approve

Approve a purchase order (transitions `created` -> `approved`).

**Required Role**: `administrator`, `operations_manager`

**Request Body**: Empty

**Success Response** (200):
```json
{
  "data": { "...PurchaseOrderResponse (status: approved)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `INVALID_TRANSITION`

---

#### POST /purchase-orders/:id/receive

Record receipt of goods against a purchase order. Compares ordered vs. received quantities and prices, auto-generating variance records for discrepancies.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`

**Request Body**:
```json
{
  "lines": [
    {
      "po_line_id": "uuid",
      "received_quantity": 18,
      "received_unit_price": 460.00
    }
  ]
}
```

**Success Response** (200):
```json
{
  "data": { "...PurchaseOrderResponse (status: received)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `INVALID_TRANSITION`, `VALIDATION_ERROR`

---

#### POST /purchase-orders/:id/return

Mark a received purchase order as returned.

**Required Role**: `administrator`, `operations_manager`

**Request Body**: Empty

**Success Response** (200):
```json
{
  "data": { "...PurchaseOrderResponse (status: returned)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `INVALID_TRANSITION`

---

#### POST /purchase-orders/:id/void

Void an approved or received purchase order.

**Required Role**: `administrator`, `operations_manager`

**Request Body**: Empty

**Success Response** (200):
```json
{
  "data": { "...PurchaseOrderResponse (status: voided)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `INVALID_TRANSITION`

---

#### GET /variances

List variance records with filtering and pagination.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`

**Query Parameters**:

| Parameter | Type | Description |
|---|---|---|
| `page` | integer | Page number |
| `page_size` | integer | Items per page |
| `status` | string | Filter: `open` or `resolved` |

**Success Response** (200):
```json
{
  "data": [
    {
      "id": "uuid",
      "po_line_id": "uuid",
      "type": "shortage",
      "expected_value": 20.0,
      "actual_value": 18.0,
      "difference_amount": -2.0,
      "status": "open",
      "resolution_due_date": "2026-04-16",
      "resolved_at": null,
      "resolution_notes": "",
      "requires_escalation": false,
      "is_overdue": false,
      "created_at": "2026-04-09T10:00:00Z"
    }
  ],
  "pagination": { "..." }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`

---

#### GET /variances/:id

Get a single variance record.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`

**Success Response** (200):
```json
{
  "data": { "...VarianceResponse" }
}
```

**Error Codes**: `NOT_FOUND`

---

#### POST /variances/:id/resolve

Resolve an open variance record.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`

**Request Body**:
```json
{
  "resolution_notes": "Supplier confirmed shortage, credit note issued for 2 units."
}
```

**Success Response** (200):
```json
{
  "data": { "...VarianceResponse (status: resolved)" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `VALIDATION_ERROR`, `VARIANCE_UNRESOLVED`

---

#### GET /procurement/landed-costs

List landed cost entries with filtering.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`

**Query Parameters**: `item_id`, `period`, `purchase_order_id`

**Success Response** (200):
```json
{
  "data": [
    {
      "id": "uuid",
      "item_id": "uuid",
      "purchase_order_id": "uuid",
      "po_line_id": "uuid",
      "period": "2026-Q1",
      "cost_component": "unit_cost",
      "raw_amount": 450.00,
      "allocated_amount": 450.00,
      "allocation_method": "direct",
      "created_at": "2026-04-09T00:00:00Z"
    }
  ]
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`

---

#### GET /procurement/landed-costs/:itemId

Get landed cost summary for a specific item.

**Required Role**: `administrator`, `operations_manager`, `procurement_specialist`

**Query Parameters**: `period`

**Success Response** (200):
```json
{
  "data": [ { "...LandedCostResponse" } ]
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`

---

### Reports & Exports

#### GET /reports

List available report definitions (filtered by the caller's role).

**Required Role**: Any authenticated user (reports filtered by `allowed_roles`)

**Success Response** (200):
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "member_growth",
      "report_type": "member_growth",
      "description": "Tracks new member sign-ups and total membership growth over time",
      "allowed_roles": ["administrator", "operations_manager"],
      "created_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

---

#### GET /reports/:id/data

Execute a report and return the data.

**Required Role**: Determined by the report's `allowed_roles` field

**Query Parameters**: Varies per report type (location, period, date range, etc.)

**Success Response** (200):
```json
{
  "data": { "...report-specific data structure" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`

---

#### POST /exports

Initiate an export job (CSV or PDF generation).

**Required Role**: Determined by the underlying report's `allowed_roles`

**Request Body**:
```json
{
  "report_id": "uuid",
  "format": "csv"
}
```

**Success Response** (202):
```json
{
  "data": {
    "id": "uuid",
    "report_id": "uuid",
    "format": "csv",
    "filename": "member_growth_20260409_100000.csv",
    "status": "pending",
    "created_by": "uuid",
    "created_at": "2026-04-09T10:00:00Z",
    "completed_at": null
  }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `VALIDATION_ERROR`

---

#### GET /exports/:id

Check the status of an export job.

**Required Role**: The export's creator or an administrator

**Success Response** (200):
```json
{
  "data": { "...ExportResponse" }
}
```

**Error Codes**: `UNAUTHORIZED`, `NOT_FOUND`

---

#### GET /exports/:id/download

Download a completed export file.

**Required Role**: The export's creator or an administrator

**Success Response** (200): Binary file with headers:
- `Content-Disposition: attachment; filename="member_growth_20260409_100000.csv"`
- `Content-Type: text/csv` or `application/pdf`

**Error Codes**: `UNAUTHORIZED`, `NOT_FOUND`, `VALIDATION_ERROR` (if status is not `completed`)

---

### Admin / Audit

#### POST /admin/users

Create a new user account.

**Required Role**: `administrator`

**Request Body**:
```json
{
  "email": "coach@fitcommerce.local",
  "password": "securepassword",
  "role": "coach",
  "display_name": "Coach Smith",
  "location_id": "uuid"
}
```

**Success Response** (201):
```json
{
  "data": { "...UserResponse" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `VALIDATION_ERROR`, `CONFLICT` (email already exists)

---

#### GET /admin/users

List all users with pagination.

**Required Role**: `administrator`

**Query Parameters**: `page`, `page_size`, `role`, `status`

**Success Response** (200): Paginated list of `UserResponse`

---

#### GET /admin/users/:id

Get a single user.

**Required Role**: `administrator`

**Success Response** (200):
```json
{
  "data": { "...UserResponse" }
}
```

**Error Codes**: `NOT_FOUND`

---

#### PUT /admin/users/:id

Update a user account (role, status, display name, location).

**Required Role**: `administrator`

**Request Body**:
```json
{
  "role": "operations_manager",
  "status": "active",
  "display_name": "Updated Name",
  "location_id": "uuid"
}
```

**Success Response** (200):
```json
{
  "data": { "...UserResponse" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `VALIDATION_ERROR`

---

#### GET /admin/audit-log

Query the tamper-evident audit log.

**Required Role**: `administrator`

**Query Parameters**:

| Parameter | Type | Description |
|---|---|---|
| `page` | integer | Page number |
| `page_size` | integer | Items per page |
| `entity_type` | string | Filter by entity type |
| `entity_id` | uuid | Filter by entity ID |
| `actor_id` | uuid | Filter by actor |
| `event_type` | string | Filter by event type |
| `from` | datetime | Start of time range |
| `to` | datetime | End of time range |

**Success Response** (200):
```json
{
  "data": [
    {
      "id": "uuid",
      "event_type": "order.created",
      "entity_type": "order",
      "entity_id": "uuid",
      "actor_id": "uuid",
      "details": { "quantity": 5 },
      "integrity_hash": "sha256-hex-string",
      "created_at": "2026-04-09T10:00:00Z"
    }
  ],
  "pagination": { "..." }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`

---

#### POST /admin/backups

Trigger a manual backup.

**Required Role**: `administrator`

**Request Body**: Empty (`{}`)

**Success Response** (202):
```json
{
  "data": {
    "id": "uuid",
    "archive_path": "/var/backups/fitcommerce/backup_20260409_100000.enc",
    "checksum": "",
    "checksum_algorithm": "sha256",
    "status": "running",
    "file_size": 0,
    "started_at": "2026-04-09T10:00:00Z",
    "completed_at": null
  }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`

---

#### GET /admin/backups

List backup history.

**Required Role**: `administrator`

**Query Parameters**: `page`, `page_size`

**Success Response** (200): Paginated list of `BackupResponse`

---

#### GET /admin/biometric/keys

List encryption keys for the biometric module.

**Required Role**: `administrator`

**Success Response** (200):
```json
{
  "data": [
    {
      "id": "uuid",
      "key_reference": "biometric-key-2026-01",
      "purpose": "biometric",
      "status": "active",
      "activated_at": "2026-01-01T00:00:00Z",
      "rotated_at": null,
      "expires_at": "2026-04-01T00:00:00Z"
    }
  ]
}
```

---

#### POST /admin/biometric/keys/rotate

Trigger rotation of the active biometric encryption key.

**Required Role**: `administrator`

**Request Body**: Empty

**Success Response** (200):
```json
{
  "data": {
    "old_key_id": "uuid",
    "new_key_id": "uuid",
    "message": "Key rotated successfully"
  }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`

---

#### GET /admin/retention-policies

List all retention policies.

**Required Role**: `administrator`

**Success Response** (200):
```json
{
  "data": [
    {
      "id": "uuid",
      "entity_type": "financial_records",
      "retention_days": 2555,
      "description": "Financial transaction records including orders, payments, and refunds"
    }
  ]
}
```

---

### Locations & Organization

#### POST /locations

Create a new location.

**Required Role**: `administrator`

**Request Body**:
```json
{
  "name": "Downtown Club",
  "address": "100 Main St",
  "timezone": "America/New_York"
}
```

**Success Response** (201):
```json
{
  "data": { "...LocationResponse" }
}
```

**Error Codes**: `UNAUTHORIZED`, `FORBIDDEN`, `VALIDATION_ERROR`

---

#### GET /locations

List all locations.

**Required Role**: Any authenticated user

**Query Parameters**: `page`, `page_size`

**Success Response** (200): Paginated list of `LocationResponse`

---

#### GET /locations/:id

Get a single location.

**Required Role**: Any authenticated user

**Success Response** (200):
```json
{
  "data": { "...LocationResponse" }
}
```

**Error Codes**: `NOT_FOUND`

---

#### POST /coaches

Create a coach profile.

**Required Role**: `administrator`

**Request Body**:
```json
{
  "user_id": "uuid",
  "location_id": "uuid",
  "specialization": "Strength Training"
}
```

**Success Response** (201):
```json
{
  "data": { "...CoachResponse" }
}
```

---

#### GET /coaches

List coaches.

**Required Role**: `administrator`, `operations_manager`

**Query Parameters**: `location_id`, `page`, `page_size`

**Success Response** (200): Paginated list of `CoachResponse`

---

#### POST /members

Create a member profile.

**Required Role**: `administrator`, `operations_manager`

**Request Body**:
```json
{
  "user_id": "uuid",
  "location_id": "uuid"
}
```

**Success Response** (201):
```json
{
  "data": { "...MemberResponse" }
}
```

---

#### GET /members

List members.

**Required Role**: `administrator`, `operations_manager`

**Query Parameters**: `location_id`, `page`, `page_size`

**Success Response** (200): Paginated list of `MemberResponse`
