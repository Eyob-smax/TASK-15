package dto

// SuccessResponse wraps a successful payload in a standard envelope.
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// PaginatedResponse wraps a paginated list payload with pagination metadata.
type PaginatedResponse struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginationMeta contains pagination metadata for list endpoints.
type PaginationMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

// UserResponse represents a user in API responses, omitting sensitive fields.
type UserResponse struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	Status      string  `json:"status"`
	DisplayName string  `json:"display_name"`
	LocationID  *string `json:"location_id,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// SessionResponse represents session information returned after login.
type SessionResponse struct {
	Token             string `json:"token"`
	IdleExpiresAt     string `json:"idle_expires_at"`
	AbsoluteExpiresAt string `json:"absolute_expires_at"`
}

// ItemResponse represents an item in API responses.
type ItemResponse struct {
	ID                  string                       `json:"id"`
	SKU                 string                       `json:"sku"`
	Name                string                       `json:"name"`
	Description         string                       `json:"description"`
	Category            string                       `json:"category"`
	Brand               string                       `json:"brand"`
	Condition           string                       `json:"condition"`
	UnitPrice           float64                      `json:"unit_price"`
	RefundableDeposit   float64                      `json:"refundable_deposit"`
	BillingModel        string                       `json:"billing_model"`
	Status              string                       `json:"status"`
	Quantity            int                          `json:"quantity"`
	LocationID          *string                      `json:"location_id,omitempty"`
	CreatedBy           string                       `json:"created_by"`
	CreatedAt           string                       `json:"created_at"`
	UpdatedAt           string                       `json:"updated_at"`
	Version             int                          `json:"version"`
	AvailabilityWindows []AvailabilityWindowResponse `json:"availability_windows,omitempty"`
	BlackoutWindows     []BlackoutWindowResponse     `json:"blackout_windows,omitempty"`
}

// AvailabilityWindowResponse represents an availability window in API responses.
type AvailabilityWindowResponse struct {
	ID        string `json:"id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// BlackoutWindowResponse represents a blackout window in API responses.
type BlackoutWindowResponse struct {
	ID        string `json:"id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// CampaignResponse represents a group-buy campaign in API responses.
type CampaignResponse struct {
	ID                  string  `json:"id"`
	ItemID              string  `json:"item_id"`
	MinQuantity         int     `json:"min_quantity"`
	MaxQuantity         *int    `json:"max_quantity,omitempty"`
	CurrentCommittedQty int     `json:"current_committed_qty"`
	CutoffTime          string  `json:"cutoff_time"`
	Status              string  `json:"status"`
	CreatedBy           string  `json:"created_by"`
	CreatedAt           string  `json:"created_at"`
	EvaluatedAt         *string `json:"evaluated_at,omitempty"`
}

// ParticipantResponse represents a campaign participant in API responses.
type ParticipantResponse struct {
	ID         string `json:"id"`
	CampaignID string `json:"campaign_id"`
	UserID     string `json:"user_id"`
	Quantity   int    `json:"quantity"`
	OrderID    string `json:"order_id"`
	JoinedAt   string `json:"joined_at"`
}

// OrderResponse represents an order in API responses.
type OrderResponse struct {
	ID               string  `json:"id"`
	UserID           string  `json:"user_id"`
	ItemID           string  `json:"item_id"`
	CampaignID       *string `json:"campaign_id,omitempty"`
	Quantity         int     `json:"quantity"`
	UnitPrice        float64 `json:"unit_price"`
	TotalAmount      float64 `json:"total_amount"`
	Status           string  `json:"status"`
	SettlementMarker string  `json:"settlement_marker,omitempty"`
	Notes            string  `json:"notes,omitempty"`
	AutoCloseAt      string  `json:"auto_close_at"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	PaidAt           *string `json:"paid_at,omitempty"`
	CancelledAt      *string `json:"cancelled_at,omitempty"`
	RefundedAt       *string `json:"refunded_at,omitempty"`
}

// TimelineEntryResponse represents an order timeline entry in API responses.
type TimelineEntryResponse struct {
	ID          string `json:"id"`
	OrderID     string `json:"order_id"`
	Action      string `json:"action"`
	Description string `json:"description"`
	PerformedBy string `json:"performed_by"`
	CreatedAt   string `json:"created_at"`
}

// SupplierResponse represents a supplier in API responses.
type SupplierResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ContactName  string `json:"contact_name"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
	Address      string `json:"address"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// PurchaseOrderResponse represents a purchase order in API responses.
type PurchaseOrderResponse struct {
	ID          string               `json:"id"`
	SupplierID  string               `json:"supplier_id"`
	Status      string               `json:"status"`
	TotalAmount float64              `json:"total_amount"`
	CreatedBy   string               `json:"created_by"`
	ApprovedBy  *string              `json:"approved_by,omitempty"`
	CreatedAt   string               `json:"created_at"`
	ApprovedAt  *string              `json:"approved_at,omitempty"`
	ReceivedAt  *string              `json:"received_at,omitempty"`
	Version     int                  `json:"version"`
	Lines       []POLineResponse     `json:"lines,omitempty"`
}

// POLineResponse represents a purchase order line in API responses.
type POLineResponse struct {
	ID                string   `json:"id"`
	PurchaseOrderID   string   `json:"purchase_order_id"`
	ItemID            string   `json:"item_id"`
	OrderedQuantity   int      `json:"ordered_quantity"`
	OrderedUnitPrice  float64  `json:"ordered_unit_price"`
	ReceivedQuantity  *int     `json:"received_quantity,omitempty"`
	ReceivedUnitPrice *float64 `json:"received_unit_price,omitempty"`
}

// VarianceResponse represents a variance record in API responses.
type VarianceResponse struct {
	ID                string  `json:"id"`
	POLineID          string  `json:"po_line_id"`
	Type              string  `json:"type"`
	ExpectedValue     float64 `json:"expected_value"`
	ActualValue       float64 `json:"actual_value"`
	DifferenceAmount  float64 `json:"difference_amount"`
	Status            string  `json:"status"`
	ResolutionDueDate string  `json:"resolution_due_date"`
	ResolvedAt        *string `json:"resolved_at,omitempty"`
	ResolutionAction  string  `json:"resolution_action,omitempty"`
	ResolutionNotes   string  `json:"resolution_notes,omitempty"`
	QuantityChange    *int    `json:"quantity_change,omitempty"`
	RequiresEscalation bool   `json:"requires_escalation"`
	IsOverdue          bool   `json:"is_overdue"`
	CreatedAt         string  `json:"created_at"`
}

// LandedCostResponse represents a landed cost entry in API responses.
type LandedCostResponse struct {
	ID               string  `json:"id"`
	ItemID           string  `json:"item_id"`
	PurchaseOrderID  string  `json:"purchase_order_id"`
	POLineID         string  `json:"po_line_id"`
	Period           string  `json:"period"`
	CostComponent    string  `json:"cost_component"`
	RawAmount        float64 `json:"raw_amount"`
	AllocatedAmount  float64 `json:"allocated_amount"`
	AllocationMethod string  `json:"allocation_method"`
	CreatedAt        string  `json:"created_at"`
}

// KPIDashboardResponse contains the key performance indicators for the dashboard.
type KPIDashboardResponse struct {
	MemberGrowth      KPIMetric `json:"member_growth"`
	Churn             KPIMetric `json:"churn"`
	RenewalRate       KPIMetric `json:"renewal_rate"`
	Engagement        KPIMetric `json:"engagement"`
	ClassFillRate     KPIMetric `json:"class_fill_rate"`
	CoachProductivity KPIMetric `json:"coach_productivity"`
}

// KPIMetric represents a single KPI value with trend information.
type KPIMetric struct {
	Value         float64 `json:"value"`
	PreviousValue float64 `json:"previous_value"`
	ChangePercent float64 `json:"change_percent"`
	Period        string  `json:"period"`
}

// BatchEditResponse contains the results of a batch edit operation.
type BatchEditResponse struct {
	JobID        string                    `json:"job_id"`
	TotalRows    int                       `json:"total_rows"`
	SuccessCount int                       `json:"success_count"`
	FailureCount int                       `json:"failure_count"`
	Results      []BatchEditResultResponse `json:"results"`
}

// BatchEditResultResponse represents a single row result in a batch edit.
type BatchEditResultResponse struct {
	ItemID        string `json:"item_id"`
	Field         string `json:"field"`
	OldValue      string `json:"old_value"`
	NewValue      string `json:"new_value"`
	Success       bool   `json:"success"`
	FailureReason string `json:"failure_reason,omitempty"`
}

// LocationResponse represents a location in API responses.
type LocationResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Timezone  string `json:"timezone"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// MemberResponse represents a member in API responses.
type MemberResponse struct {
	ID               string  `json:"id"`
	UserID           string  `json:"user_id"`
	LocationID       string  `json:"location_id"`
	MembershipStatus string  `json:"membership_status"`
	JoinedAt         string  `json:"joined_at"`
	RenewalDate      *string `json:"renewal_date,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// CoachResponse represents a coach in API responses.
type CoachResponse struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	LocationID     string `json:"location_id"`
	Specialization string `json:"specialization"`
	IsActive       bool   `json:"is_active"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// AuditEventResponse represents an audit log entry in API responses.
type AuditEventResponse struct {
	ID            string                 `json:"id"`
	EventType     string                 `json:"event_type"`
	EntityType    string                 `json:"entity_type"`
	EntityID      string                 `json:"entity_id"`
	ActorID       string                 `json:"actor_id"`
	Details       map[string]interface{} `json:"details"`
	IntegrityHash string                 `json:"integrity_hash"`
	CreatedAt     string                 `json:"created_at"`
}

// BackupResponse represents a backup run in API responses.
type BackupResponse struct {
	ID               string  `json:"id"`
	ArchivePath      string  `json:"archive_path"`
	Checksum         string  `json:"checksum"`
	ChecksumAlgorithm string `json:"checksum_algorithm"`
	Status           string  `json:"status"`
	FileSize         int64   `json:"file_size"`
	StartedAt        string  `json:"started_at"`
	CompletedAt      *string `json:"completed_at,omitempty"`
}

// ExportResponse represents an export job in API responses.
type ExportResponse struct {
	ID          string  `json:"id"`
	ReportID    string  `json:"report_id"`
	Format      string  `json:"format"`
	Filename    string  `json:"filename"`
	Status      string  `json:"status"`
	CreatedBy   string  `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
	CompletedAt *string `json:"completed_at,omitempty"`
}

// InventorySnapshotResponse represents an inventory snapshot in API responses.
type InventorySnapshotResponse struct {
	ID             string  `json:"id"`
	ItemID         string  `json:"item_id"`
	QuantityOnHand int     `json:"quantity_on_hand"`
	LocationID     *string `json:"location_id,omitempty"`
	SnapshotDate   string  `json:"snapshot_date"`
}

// InventoryAdjustmentResponse represents an inventory adjustment in API responses.
type InventoryAdjustmentResponse struct {
	ID             string `json:"id"`
	ItemID         string `json:"item_id"`
	QuantityChange int    `json:"quantity_change"`
	Reason         string `json:"reason"`
	CreatedBy      string `json:"created_by"`
	CreatedAt      string `json:"created_at"`
}

// WarehouseBinResponse represents a warehouse bin in API responses.
type WarehouseBinResponse struct {
	ID          string `json:"id"`
	LocationID  string `json:"location_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ReportResponse represents a report definition in API responses.
type ReportResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	ReportType  string   `json:"report_type"`
	Description string   `json:"description"`
	AllowedRoles []string `json:"allowed_roles"`
	CreatedAt   string   `json:"created_at"`
}

// CaptchaResponse is embedded in error responses when CAPTCHA verification is
// required. It provides the challenge ID and the human-readable challenge text
// so the client can display the puzzle to the user.
type CaptchaResponse struct {
	ChallengeID   string `json:"challenge_id"`
	ChallengeData string `json:"challenge_data"`
}

// LoginResponse is the success payload returned by POST /auth/login. The session
// token is delivered via an HttpOnly cookie and is NOT included in this body.
type LoginResponse struct {
	User    UserResponse        `json:"user"`
	Session SessionMetaResponse `json:"session"`
}

// SessionMetaResponse contains the session timing metadata returned at login and
// by GET /auth/session. The token itself is never included in response bodies.
type SessionMetaResponse struct {
	IdleExpiresAt     string `json:"idle_expires_at"`
	AbsoluteExpiresAt string `json:"absolute_expires_at"`
}

// RetentionPolicyResponse represents a retention policy in API responses.
type RetentionPolicyResponse struct {
	ID            string `json:"id"`
	EntityType    string `json:"entity_type"`
	RetentionDays int    `json:"retention_days"`
	UpdatedAt     string `json:"updated_at"`
}

// BiometricEnrollmentResponse represents a biometric enrollment in API responses.
// The TemplateRef field is always redacted; only enrollment metadata is exposed.
type BiometricEnrollmentResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	TemplateRef string `json:"template_ref"` // always "[BIOMETRIC REDACTED]"
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// EncryptionKeyResponse represents an encryption key record in API responses.
// Key material is never returned; only metadata is exposed.
type EncryptionKeyResponse struct {
	ID        string `json:"id"`
	Purpose   string `json:"purpose"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

