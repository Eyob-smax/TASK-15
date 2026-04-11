package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
)

// AuthService handles authentication, session management, and CAPTCHA verification.
type AuthService interface {
	Login(ctx context.Context, email, password string) (*domain.Session, *domain.User, error)
	Logout(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (*domain.Session, *domain.User, error)
	VerifyCaptcha(ctx context.Context, challengeID uuid.UUID, answer string) error
}

// UserService manages user accounts.
type UserService interface {
	Create(ctx context.Context, email, password string, role domain.UserRole, displayName string, locationID *uuid.UUID) (*domain.User, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.User, error)
	List(ctx context.Context, page, pageSize int) ([]domain.User, int, error)
	Update(ctx context.Context, user *domain.User) error
	Deactivate(ctx context.Context, id uuid.UUID) error
}

// ItemService manages the equipment/product catalog.
type ItemService interface {
	Create(ctx context.Context, item *domain.Item, availability []domain.AvailabilityWindow, blackouts []domain.BlackoutWindow) (*domain.Item, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.Item, error)
	GetDetail(ctx context.Context, id uuid.UUID) (*ItemDetail, error)
	List(ctx context.Context, page, pageSize int, filters map[string]string) ([]domain.Item, int, error)
	Update(ctx context.Context, item *domain.Item, availability []domain.AvailabilityWindow, blackouts []domain.BlackoutWindow) (*domain.Item, error)
	Publish(ctx context.Context, id uuid.UUID) error
	Unpublish(ctx context.Context, id uuid.UUID) error
	BatchEdit(ctx context.Context, createdBy uuid.UUID, edits []BatchEditInput) (*domain.BatchEditJob, []domain.BatchEditResult, error)
}

// ItemDetail contains the core item plus its configured availability and blackout windows.
type ItemDetail struct {
	Item                *domain.Item
	AvailabilityWindows []domain.AvailabilityWindow
	BlackoutWindows     []domain.BlackoutWindow
}

// BatchEditInput represents a single edit instruction for batch operations.
type BatchEditInput struct {
	ItemID              uuid.UUID
	Field               string
	NewValue            *string
	AvailabilityWindows []domain.AvailabilityWindow
}

// InventoryService manages inventory snapshots, adjustments, and warehouse bins.
type InventoryService interface {
	GetSnapshots(ctx context.Context, itemID *uuid.UUID, locationID *uuid.UUID) ([]domain.InventorySnapshot, error)
	CreateAdjustment(ctx context.Context, adjustment *domain.InventoryAdjustment) (*domain.InventoryAdjustment, error)
	ListAdjustments(ctx context.Context, itemID *uuid.UUID, page, pageSize int) ([]domain.InventoryAdjustment, int, error)
	CreateWarehouseBin(ctx context.Context, bin *domain.WarehouseBin) (*domain.WarehouseBin, error)
	GetWarehouseBin(ctx context.Context, id uuid.UUID) (*domain.WarehouseBin, error)
	ListWarehouseBins(ctx context.Context, locationID *uuid.UUID, page, pageSize int) ([]domain.WarehouseBin, int, error)
}

// CampaignService manages group-buy campaigns.
type CampaignService interface {
	Create(ctx context.Context, campaign *domain.GroupBuyCampaign) (*domain.GroupBuyCampaign, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.GroupBuyCampaign, error)
	List(ctx context.Context, page, pageSize int) ([]domain.GroupBuyCampaign, int, error)
	Join(ctx context.Context, campaignID, userID uuid.UUID, quantity int) (*domain.GroupBuyParticipant, error)
	Cancel(ctx context.Context, id uuid.UUID, performedBy uuid.UUID) error
	EvaluateAtCutoff(ctx context.Context, id uuid.UUID, now time.Time, performedBy uuid.UUID) error
	// ListPastCutoff returns active campaigns whose cutoff time has passed.
	// Used by CutoffEvalJob to dispatch evaluation without direct repo access.
	ListPastCutoff(ctx context.Context, now time.Time) ([]domain.GroupBuyCampaign, error)
}

// FulfillmentInput carries optional fulfillment metadata for traceability on
// split and merge operations. When provided, the service creates a
// fulfillment_group record linking the resulting orders to a supplier, bin, and
// pickup point.
type FulfillmentInput struct {
	SupplierID     *uuid.UUID
	WarehouseBinID *uuid.UUID
	PickupPoint    string
}

// OrderService manages customer orders throughout their lifecycle.
type OrderService interface {
	Create(ctx context.Context, order *domain.Order) (*domain.Order, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetForActor(ctx context.Context, actor *domain.User, id uuid.UUID) (*domain.Order, error)
	List(ctx context.Context, userID *uuid.UUID, page, pageSize int) ([]domain.Order, int, error)
	ListForActor(ctx context.Context, actor *domain.User, page, pageSize int) ([]domain.Order, int, error)
	Pay(ctx context.Context, id uuid.UUID, settlementMarker string, performedBy uuid.UUID) error
	Cancel(ctx context.Context, id uuid.UUID, performedBy uuid.UUID) error
	CancelForActor(ctx context.Context, actor *domain.User, id uuid.UUID) error
	Refund(ctx context.Context, id uuid.UUID, performedBy uuid.UUID) error
	AddNote(ctx context.Context, id uuid.UUID, note string, performedBy uuid.UUID) error
	Split(ctx context.Context, orderID uuid.UUID, quantities []int, fulfillment *FulfillmentInput) ([]domain.Order, error)
	SplitForActor(ctx context.Context, actor *domain.User, orderID uuid.UUID, quantities []int, fulfillment *FulfillmentInput) ([]domain.Order, error)
	Merge(ctx context.Context, orderIDs []uuid.UUID, fulfillment *FulfillmentInput) (*domain.Order, error)
	MergeForActor(ctx context.Context, actor *domain.User, orderIDs []uuid.UUID, fulfillment *FulfillmentInput) (*domain.Order, error)
	AutoCloseExpired(ctx context.Context, now time.Time) (int, error)
	// GetTimeline returns the chronological timeline of events for an order.
	GetTimeline(ctx context.Context, orderID uuid.UUID) ([]domain.OrderTimelineEntry, error)
	GetTimelineForActor(ctx context.Context, actor *domain.User, orderID uuid.UUID) ([]domain.OrderTimelineEntry, error)
}

// SupplierService manages supplier records.
type SupplierService interface {
	Create(ctx context.Context, supplier *domain.Supplier) (*domain.Supplier, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.Supplier, error)
	List(ctx context.Context, page, pageSize int) ([]domain.Supplier, int, error)
	Update(ctx context.Context, supplier *domain.Supplier) error
}

// PurchaseOrderService manages the procurement workflow.
type PurchaseOrderService interface {
	Create(ctx context.Context, po *domain.PurchaseOrder, lines []domain.PurchaseOrderLine) (*domain.PurchaseOrder, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.PurchaseOrder, []domain.PurchaseOrderLine, error)
	List(ctx context.Context, page, pageSize int) ([]domain.PurchaseOrder, int, error)
	Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) error
	Receive(ctx context.Context, id uuid.UUID, receivedLines []ReceivedLineInput, actorID uuid.UUID) error
	Return(ctx context.Context, id uuid.UUID, performedBy uuid.UUID) error
	Void(ctx context.Context, id uuid.UUID, performedBy uuid.UUID) error
}

// ReceivedLineInput represents the received data for a purchase order line.
type ReceivedLineInput struct {
	POLineID          uuid.UUID
	ReceivedQuantity  int
	ReceivedUnitPrice float64
}

// VarianceService manages variance records from procurement receipt.
type VarianceService interface {
	Get(ctx context.Context, id uuid.UUID) (*domain.VarianceRecord, error)
	List(ctx context.Context, status *domain.VarianceStatus, page, pageSize int) ([]domain.VarianceRecord, int, error)
	Resolve(ctx context.Context, id uuid.UUID, action, resolutionNotes string, quantityChange *int, performedBy uuid.UUID) error
	EscalateOverdue(ctx context.Context) (int, error)
}

// LocationService manages physical location records.
type LocationService interface {
	Create(ctx context.Context, location *domain.Location) (*domain.Location, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Location, error)
	List(ctx context.Context, page, pageSize int) ([]domain.Location, int, error)
}

// MemberService manages gym member records.
type MemberService interface {
	Create(ctx context.Context, member *domain.Member) (*domain.Member, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Member, error)
	GetByIDForActor(ctx context.Context, actor *domain.User, id uuid.UUID) (*domain.Member, error)
	List(ctx context.Context, locationID *uuid.UUID, page, pageSize int) ([]domain.Member, int, error)
	ListForActor(ctx context.Context, actor *domain.User, requestedLocationID *uuid.UUID, page, pageSize int) ([]domain.Member, int, error)
}

// CoachService manages coach records.
type CoachService interface {
	Create(ctx context.Context, coach *domain.Coach) (*domain.Coach, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Coach, error)
	GetByIDForActor(ctx context.Context, actor *domain.User, id uuid.UUID) (*domain.Coach, error)
	List(ctx context.Context, locationID *uuid.UUID, page, pageSize int) ([]domain.Coach, int, error)
	ListForActor(ctx context.Context, actor *domain.User, requestedLocationID *uuid.UUID, page, pageSize int) ([]domain.Coach, int, error)
}

// LandedCostService provides landed cost summaries for items.
type LandedCostService interface {
	GetSummary(ctx context.Context, itemID uuid.UUID, period string) ([]domain.LandedCostEntry, error)
	GetByPOID(ctx context.Context, poID uuid.UUID) ([]domain.LandedCostEntry, error)
}

// ReportService manages report definitions and export generation.
type ReportService interface {
	GetReport(ctx context.Context, id uuid.UUID) (*domain.ReportDefinition, error)
	List(ctx context.Context, userRole domain.UserRole) ([]domain.ReportDefinition, error)
	GetData(ctx context.Context, reportID uuid.UUID, filters map[string]string, callerRole domain.UserRole) (interface{}, error)
	GenerateExport(ctx context.Context, reportID uuid.UUID, format domain.ExportFormat, parameters map[string]string, createdBy uuid.UUID, callerRole domain.UserRole) (*domain.ExportJob, error)
	GetExport(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRole domain.UserRole) (*domain.ExportJob, error)
	DownloadExport(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRole domain.UserRole) (*domain.ExportJob, error)
}

// AuditService manages the append-only audit log.
type AuditService interface {
	Log(ctx context.Context, eventType, entityType string, entityID, actorID uuid.UUID, details map[string]interface{}) error
	List(ctx context.Context, entityType string, entityID *uuid.UUID, page, pageSize int) ([]domain.AuditEvent, int, error)
	// ListByEventTypes filters audit events by a set of event type strings.
	// Used for security-event inspection (login failures, lockouts, session events).
	ListByEventTypes(ctx context.Context, eventTypes []string, page, pageSize int) ([]domain.AuditEvent, int, error)
}

// BackupService manages database backup operations.
// performedBy is nil for scheduled/system-triggered runs; non-nil for manually triggered runs.
// BackupRun does not store the actor; it is recorded only in the audit event.
type BackupService interface {
	Trigger(ctx context.Context, performedBy *uuid.UUID) (*domain.BackupRun, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BackupRun, error)
	List(ctx context.Context, page, pageSize int) ([]domain.BackupRun, int, error)
	VerifyIntegrity(ctx context.Context, id uuid.UUID) (*domain.BackupRun, error)
}

// RetentionService manages data retention policies and enforces the cleanup job
// by hard-deleting records older than the configured retention window.
type RetentionService interface {
	GetByEntityType(ctx context.Context, entityType string) (*domain.RetentionPolicy, error)
	List(ctx context.Context) ([]domain.RetentionPolicy, error)
	Update(ctx context.Context, policy *domain.RetentionPolicy) error
	RunCleanup(ctx context.Context) error
}

// BiometricService manages biometric enrollment records and encryption key rotation.
// All template references are stored as opaque refs; raw biometric data never enters the application layer.
type BiometricService interface {
	Register(ctx context.Context, userID uuid.UUID, templateRef string) (*domain.BiometricEnrollment, error)
	GetByUser(ctx context.Context, userID uuid.UUID) (*domain.BiometricEnrollment, error)
	Revoke(ctx context.Context, userID uuid.UUID) error
	RotateKey(ctx context.Context, performedBy uuid.UUID) (*domain.EncryptionKey, error)
	GetActiveKey(ctx context.Context) (*domain.EncryptionKey, error)
	ListKeys(ctx context.Context) ([]domain.EncryptionKey, error)
}

// DashboardService provides aggregated KPI data for the operations dashboard.
type DashboardService interface {
	GetKPIs(ctx context.Context, locationID *uuid.UUID, period string, coachID *uuid.UUID, category string, from, to string) (*DashboardKPIs, error)
}

// DashboardKPIs holds the computed KPI values for the dashboard.
type DashboardKPIs struct {
	MemberGrowth      KPIValue
	Churn             KPIValue
	RenewalRate       KPIValue
	Engagement        KPIValue
	ClassFillRate     KPIValue
	CoachProductivity KPIValue
}

// KPIValue holds a metric value with trend data.
type KPIValue struct {
	Value         float64
	PreviousValue float64
	ChangePercent float64
	Period        string
}
