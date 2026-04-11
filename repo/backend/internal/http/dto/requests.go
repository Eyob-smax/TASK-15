package dto

import "time"

// LoginRequest contains credentials for user authentication.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// CreateItemRequest contains fields for creating a new catalog item.
type CreateItemRequest struct {
	SKU                 string                       `json:"sku"`
	Name                string                       `json:"name" validate:"required"`
	Description         string                       `json:"description"`
	Category            string                       `json:"category" validate:"required"`
	Brand               string                       `json:"brand" validate:"required"`
	Condition           string                       `json:"condition" validate:"required"`
	UnitPrice           float64                      `json:"unit_price" validate:"gte=0"`
	RefundableDeposit   float64                      `json:"refundable_deposit"`
	BillingModel        string                       `json:"billing_model" validate:"required"`
	Quantity            int                          `json:"quantity" validate:"gte=0"`
	LocationID          *string                      `json:"location_id"`
	AvailabilityWindows []AvailabilityWindowRequest  `json:"availability_windows"`
	BlackoutWindows     []BlackoutWindowRequest      `json:"blackout_windows"`
}

// UpdateItemRequest contains optional fields for updating a catalog item.
type UpdateItemRequest struct {
	SKU                 *string                      `json:"sku"`
	Name                *string                      `json:"name"`
	Description         *string                      `json:"description"`
	Category            *string                      `json:"category"`
	Brand               *string                      `json:"brand"`
	Condition           *string                      `json:"condition"`
	UnitPrice           *float64                     `json:"unit_price"`
	RefundableDeposit   *float64                     `json:"refundable_deposit"`
	BillingModel        *string                      `json:"billing_model"`
	Quantity            *int                         `json:"quantity"`
	LocationID          *string                      `json:"location_id"`
	AvailabilityWindows *[]AvailabilityWindowRequest `json:"availability_windows"`
	BlackoutWindows     *[]BlackoutWindowRequest     `json:"blackout_windows"`
	Version             int                          `json:"version" validate:"required"`
}

// AvailabilityWindowRequest defines an availability time window for an item.
type AvailabilityWindowRequest struct {
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
}

// BlackoutWindowRequest defines a blackout time window for an item.
type BlackoutWindowRequest struct {
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
}

// BatchEditRequest contains a set of row-level edits to apply to items.
type BatchEditRequest struct {
	Edits []BatchEditRow `json:"edits" validate:"required,min=1"`
}

// BatchEditRow describes a single field change on a single item.
type BatchEditRow struct {
	ItemID              string                      `json:"item_id" validate:"required,uuid"`
	Field               string                      `json:"field" validate:"required,oneof=name description category brand condition billing_model status unit_price refundable_deposit quantity availability_windows"`
	NewValue            *string                     `json:"new_value"`
	AvailabilityWindows []AvailabilityWindowRequest `json:"availability_windows"`
}

// CreateCampaignRequest contains fields for creating a group-buy campaign.
type CreateCampaignRequest struct {
	ItemID      string    `json:"item_id" validate:"required"`
	MinQuantity int       `json:"min_quantity" validate:"required,gt=0"`
	CutoffTime  time.Time `json:"cutoff_time" validate:"required"`
}

// JoinCampaignRequest contains the quantity for joining a group-buy campaign.
type JoinCampaignRequest struct {
	Quantity int `json:"quantity" validate:"required,gt=0"`
}

// CreateOrderRequest contains fields for creating a new order.
type CreateOrderRequest struct {
	ItemID     string  `json:"item_id" validate:"required"`
	CampaignID *string `json:"campaign_id"`
	Quantity   int     `json:"quantity" validate:"required,gt=0"`
}

// PayOrderRequest contains the settlement marker for order payment.
type PayOrderRequest struct {
	SettlementMarker string `json:"settlement_marker" validate:"required"`
}

// AddOrderNoteRequest contains a note to add to an order's timeline.
type AddOrderNoteRequest struct {
	Note string `json:"note" validate:"required"`
}

// CreateSupplierRequest contains fields for creating a new supplier.
type CreateSupplierRequest struct {
	Name         string `json:"name" validate:"required"`
	ContactName  string `json:"contact_name"`
	ContactEmail string `json:"contact_email" validate:"omitempty,email"`
	ContactPhone string `json:"contact_phone"`
	Address      string `json:"address"`
}

// CreatePurchaseOrderRequest contains fields for creating a purchase order.
type CreatePurchaseOrderRequest struct {
	SupplierID string            `json:"supplier_id" validate:"required"`
	Lines      []POLineRequest   `json:"lines" validate:"required,min=1"`
}

// POLineRequest describes a single line item in a purchase order.
type POLineRequest struct {
	ItemID          string  `json:"item_id" validate:"required"`
	OrderedQuantity int     `json:"ordered_quantity" validate:"required,gt=0"`
	OrderedUnitPrice float64 `json:"ordered_unit_price" validate:"required,gt=0"`
}

// ReceivePurchaseOrderRequest contains received line details for a purchase order.
type ReceivePurchaseOrderRequest struct {
	Lines []POReceiveLineRequest `json:"lines" validate:"required,min=1"`
}

// POReceiveLineRequest describes received quantities and prices for a PO line.
type POReceiveLineRequest struct {
	POLineID          string  `json:"po_line_id" validate:"required"`
	ReceivedQuantity  int     `json:"received_quantity" validate:"required,gte=0"`
	ReceivedUnitPrice float64 `json:"received_unit_price" validate:"required,gte=0"`
}

// ResolveVarianceRequest contains the resolution notes for closing a variance.
type ResolveVarianceRequest struct {
	Action          string `json:"action" validate:"required,oneof=adjustment return"`
	ResolutionNotes string `json:"resolution_notes" validate:"required"`
	QuantityChange  *int   `json:"quantity_change"`
}

// CreateExportRequest initiates the generation of a report export.
type CreateExportRequest struct {
	ReportID   string            `json:"report_id" validate:"required"`
	Format     string            `json:"format" validate:"required,oneof=csv pdf"`
	Parameters map[string]string `json:"parameters"`
}

// TriggerBackupRequest initiates a database backup.
type TriggerBackupRequest struct{}

// CreateUserRequest contains fields for creating a new user account.
type CreateUserRequest struct {
	Email       string  `json:"email" validate:"required,email"`
	Password    string  `json:"password" validate:"required,min=8"`
	Role        string  `json:"role" validate:"required"`
	DisplayName string  `json:"display_name" validate:"required"`
	LocationID  *string `json:"location_id"`
}

// CreateLocationRequest contains fields for creating a new location.
type CreateLocationRequest struct {
	Name     string `json:"name" validate:"required"`
	Address  string `json:"address" validate:"required"`
	Timezone string `json:"timezone" validate:"required"`
}

// InventoryAdjustmentRequest contains fields for manually adjusting inventory.
type InventoryAdjustmentRequest struct {
	ItemID         string `json:"item_id" validate:"required"`
	QuantityChange int    `json:"quantity_change" validate:"required"`
	Reason         string `json:"reason" validate:"required"`
}

// CaptchaVerifyRequest contains the challenge ID and submitted answer for
// verifying a CAPTCHA challenge after a login lockout.
type CaptchaVerifyRequest struct {
	ChallengeID string `json:"challenge_id" validate:"required"`
	Answer      string `json:"answer" validate:"required"`
}

// SplitOrderRequest contains the sub-quantities for splitting an order.
// The quantities must sum to the original order's quantity.
// The optional fulfillment fields (supplier_id, warehouse_bin_id, pickup_point)
// enable supplier/bin/pickup-aware traceability grouping.
type SplitOrderRequest struct {
	Quantities     []int   `json:"quantities" validate:"required,min=2"`
	SupplierID     *string `json:"supplier_id"`
	WarehouseBinID *string `json:"warehouse_bin_id"`
	PickupPoint    string  `json:"pickup_point"`
}

// MergeOrdersRequest contains the IDs of orders to merge into one.
// All orders must belong to the same user and item.
// The optional fulfillment fields enable traceability grouping on the merged order.
type MergeOrdersRequest struct {
	OrderIDs       []string `json:"order_ids" validate:"required,min=2"`
	SupplierID     *string  `json:"supplier_id"`
	WarehouseBinID *string  `json:"warehouse_bin_id"`
	PickupPoint    string   `json:"pickup_point"`
}

// CreateWarehouseBinRequest contains fields for creating a warehouse bin.
type CreateWarehouseBinRequest struct {
	LocationID  string `json:"location_id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

// UpdateUserRequest contains fields for updating a user account.
type UpdateUserRequest struct {
	DisplayName string  `json:"display_name"`
	Role        string  `json:"role"`
	LocationID  *string `json:"location_id"`
}

// ReturnPurchaseOrderRequest contains an optional reason for returning a PO.
type ReturnPurchaseOrderRequest struct {
	Reason string `json:"reason"`
}

// UpdateRetentionPolicyRequest contains the updated retention duration for a policy.
type UpdateRetentionPolicyRequest struct {
	RetentionDays int `json:"retention_days" validate:"required,gt=0"`
}

// RegisterBiometricRequest contains fields for enrolling a user's biometric template.
type RegisterBiometricRequest struct {
	UserID      string `json:"user_id" validate:"required"`
	TemplateRef string `json:"template_ref" validate:"required"`
}
