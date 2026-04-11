package store

import (
	"context"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
)

// UserRepository manages persistence for user entities.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	List(ctx context.Context, page, pageSize int) ([]domain.User, int, error)
	ListByRole(ctx context.Context, role domain.UserRole, page, pageSize int) ([]domain.User, int, error)
}

// SessionRepository manages persistence for user sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByToken(ctx context.Context, token string) (*domain.Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context, now time.Time) (int, error)
	UpdateIdleExpiry(ctx context.Context, id uuid.UUID, newExpiry time.Time) error
}

// CaptchaRepository manages persistence for CAPTCHA challenges.
type CaptchaRepository interface {
	Create(ctx context.Context, challenge *domain.CaptchaChallenge) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CaptchaChallenge, error)
	MarkVerified(ctx context.Context, id uuid.UUID) error
}

// LocationRepository manages persistence for location entities.
type LocationRepository interface {
	Create(ctx context.Context, location *domain.Location) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Location, error)
	List(ctx context.Context, page, pageSize int) ([]domain.Location, int, error)
}

// MemberRepository manages persistence for member entities.
type MemberRepository interface {
	Create(ctx context.Context, member *domain.Member) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Member, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Member, error)
	List(ctx context.Context, locationID *uuid.UUID, page, pageSize int) ([]domain.Member, int, error)
	CountByPeriod(ctx context.Context, locationID *uuid.UUID, start, end time.Time) (int, error)
}

// CoachRepository manages persistence for coach entities.
type CoachRepository interface {
	Create(ctx context.Context, coach *domain.Coach) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Coach, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Coach, error)
	List(ctx context.Context, locationID *uuid.UUID, page, pageSize int) ([]domain.Coach, int, error)
}

// FulfillmentRepository manages persistence for fulfillment groups and their
// order associations, enabling supplier/bin/pickup traceability on split/merge.
type FulfillmentRepository interface {
	CreateGroup(ctx context.Context, group *domain.FulfillmentGroup) error
	AddGroupOrder(ctx context.Context, entry *domain.FulfillmentGroupOrder) error
}

// ItemRepository manages persistence for catalog items.
type ItemRepository interface {
	Create(ctx context.Context, item *domain.Item) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Item, error)
	List(ctx context.Context, filters map[string]string, page, pageSize int) ([]domain.Item, int, error)
	Update(ctx context.Context, item *domain.Item) error
	BatchUpdate(ctx context.Context, items []*domain.Item) error
}

// AvailabilityWindowRepository manages persistence for item availability windows.
type AvailabilityWindowRepository interface {
	Create(ctx context.Context, window *domain.AvailabilityWindow) error
	ListByItemID(ctx context.Context, itemID uuid.UUID) ([]domain.AvailabilityWindow, error)
	DeleteByItemID(ctx context.Context, itemID uuid.UUID) error
}

// BlackoutWindowRepository manages persistence for item blackout windows.
type BlackoutWindowRepository interface {
	Create(ctx context.Context, window *domain.BlackoutWindow) error
	ListByItemID(ctx context.Context, itemID uuid.UUID) ([]domain.BlackoutWindow, error)
	DeleteByItemID(ctx context.Context, itemID uuid.UUID) error
}

// InventoryRepository manages persistence for inventory snapshots and adjustments.
type InventoryRepository interface {
	CreateSnapshot(ctx context.Context, snapshot *domain.InventorySnapshot) error
	ListSnapshots(ctx context.Context, itemID *uuid.UUID, locationID *uuid.UUID) ([]domain.InventorySnapshot, error)
	CreateAdjustment(ctx context.Context, adjustment *domain.InventoryAdjustment) error
	ListAdjustments(ctx context.Context, itemID *uuid.UUID, page, pageSize int) ([]domain.InventoryAdjustment, int, error)
}

// WarehouseBinRepository manages persistence for warehouse bins.
type WarehouseBinRepository interface {
	Create(ctx context.Context, bin *domain.WarehouseBin) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.WarehouseBin, error)
	List(ctx context.Context, locationID *uuid.UUID, page, pageSize int) ([]domain.WarehouseBin, int, error)
}

// BatchEditRepository manages persistence for batch edit jobs and results.
type BatchEditRepository interface {
	CreateJob(ctx context.Context, job *domain.BatchEditJob) error
	CreateResult(ctx context.Context, result *domain.BatchEditResult) error
	GetJob(ctx context.Context, id uuid.UUID) (*domain.BatchEditJob, error)
	ListResults(ctx context.Context, batchID uuid.UUID) ([]domain.BatchEditResult, error)
}

// CampaignRepository manages persistence for group-buy campaigns.
type CampaignRepository interface {
	Create(ctx context.Context, campaign *domain.GroupBuyCampaign) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.GroupBuyCampaign, error)
	List(ctx context.Context, page, pageSize int) ([]domain.GroupBuyCampaign, int, error)
	Update(ctx context.Context, campaign *domain.GroupBuyCampaign) error
	ListActive(ctx context.Context) ([]domain.GroupBuyCampaign, error)
	// ListDueCampaigns returns active campaigns whose cutoff_at is <= now.
	// Used by CutoffEvalJob to find campaigns requiring evaluation.
	ListDueCampaigns(ctx context.Context, now time.Time) ([]domain.GroupBuyCampaign, error)
}

// ParticipantRepository manages persistence for campaign participants.
type ParticipantRepository interface {
	Create(ctx context.Context, participant *domain.GroupBuyParticipant) error
	ListByCampaign(ctx context.Context, campaignID uuid.UUID) ([]domain.GroupBuyParticipant, error)
	CountCommittedQuantity(ctx context.Context, campaignID uuid.UUID) (int, error)
}

// OrderRepository manages persistence for customer orders.
type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	List(ctx context.Context, userID *uuid.UUID, page, pageSize int) ([]domain.Order, int, error)
	Update(ctx context.Context, order *domain.Order) error
	ListExpiredUnpaid(ctx context.Context, now time.Time) ([]domain.Order, error)
}

// TimelineRepository manages persistence for order timeline entries.
type TimelineRepository interface {
	Create(ctx context.Context, entry *domain.OrderTimelineEntry) error
	ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]domain.OrderTimelineEntry, error)
}

// SupplierRepository manages persistence for supplier entities.
type SupplierRepository interface {
	Create(ctx context.Context, supplier *domain.Supplier) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error)
	List(ctx context.Context, page, pageSize int) ([]domain.Supplier, int, error)
	Update(ctx context.Context, supplier *domain.Supplier) error
}

// PurchaseOrderRepository manages persistence for purchase orders.
type PurchaseOrderRepository interface {
	Create(ctx context.Context, po *domain.PurchaseOrder) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseOrder, error)
	List(ctx context.Context, page, pageSize int) ([]domain.PurchaseOrder, int, error)
	Update(ctx context.Context, po *domain.PurchaseOrder) error
}

// POLineRepository manages persistence for purchase order line items.
type POLineRepository interface {
	Create(ctx context.Context, line *domain.PurchaseOrderLine) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseOrderLine, error)
	ListByPOID(ctx context.Context, poID uuid.UUID) ([]domain.PurchaseOrderLine, error)
	Update(ctx context.Context, line *domain.PurchaseOrderLine) error
}

// VarianceRepository manages persistence for variance records.
type VarianceRepository interface {
	Create(ctx context.Context, record *domain.VarianceRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.VarianceRecord, error)
	List(ctx context.Context, status *domain.VarianceStatus, page, pageSize int) ([]domain.VarianceRecord, int, error)
	Update(ctx context.Context, record *domain.VarianceRecord) error
	ListUnresolved(ctx context.Context) ([]domain.VarianceRecord, error)
}

// LandedCostRepository manages persistence for landed cost entries.
type LandedCostRepository interface {
	Create(ctx context.Context, entry *domain.LandedCostEntry) error
	ListByItemAndPeriod(ctx context.Context, itemID uuid.UUID, period string) ([]domain.LandedCostEntry, error)
	ListByPOID(ctx context.Context, poID uuid.UUID) ([]domain.LandedCostEntry, error)
}

// AuditRepository manages persistence for the append-only audit log.
type AuditRepository interface {
	Create(ctx context.Context, event *domain.AuditEvent) error
	List(ctx context.Context, entityType string, entityID *uuid.UUID, page, pageSize int) ([]domain.AuditEvent, int, error)
	ListByEventTypes(ctx context.Context, eventTypes []string, page, pageSize int) ([]domain.AuditEvent, int, error)
	GetLatestHash(ctx context.Context) (string, error)
}

// ReportRepository manages persistence for report definitions.
type ReportRepository interface {
	Create(ctx context.Context, report *domain.ReportDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ReportDefinition, error)
	List(ctx context.Context) ([]domain.ReportDefinition, error)
}

// ExportRepository manages persistence for export jobs.
type ExportRepository interface {
	Create(ctx context.Context, job *domain.ExportJob) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ExportJob, error)
	Update(ctx context.Context, job *domain.ExportJob) error
}

// BackupRepository manages persistence for backup run records.
type BackupRepository interface {
	Create(ctx context.Context, run *domain.BackupRun) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BackupRun, error)
	List(ctx context.Context, page, pageSize int) ([]domain.BackupRun, int, error)
	Update(ctx context.Context, run *domain.BackupRun) error
}

// RetentionRepository manages persistence for retention policies.
type RetentionRepository interface {
	Create(ctx context.Context, policy *domain.RetentionPolicy) error
	GetByEntityType(ctx context.Context, entityType string) (*domain.RetentionPolicy, error)
	List(ctx context.Context) ([]domain.RetentionPolicy, error)
	Update(ctx context.Context, policy *domain.RetentionPolicy) error
	// DeleteByIDs deletes records from the given table by ID slice.
	// Uses executorFromContext so it participates in any ambient transaction.
	DeleteByIDs(ctx context.Context, table string, ids []uuid.UUID) (int64, error)
}

// BiometricRepository manages persistence for biometric enrollments.
type BiometricRepository interface {
	Create(ctx context.Context, enrollment *domain.BiometricEnrollment) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.BiometricEnrollment, error)
	List(ctx context.Context) ([]domain.BiometricEnrollment, error)
	Update(ctx context.Context, enrollment *domain.BiometricEnrollment) error
}

// EncryptionKeyRepository manages persistence for encryption keys.
type EncryptionKeyRepository interface {
	Create(ctx context.Context, key *domain.EncryptionKey) error
	GetActive(ctx context.Context, purpose string) (*domain.EncryptionKey, error)
	List(ctx context.Context, purpose string) ([]domain.EncryptionKey, error)
	Update(ctx context.Context, key *domain.EncryptionKey) error
}
