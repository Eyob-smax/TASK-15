package domain

// --- UserRole ---

type UserRole string

const (
	UserRoleAdministrator         UserRole = "administrator"
	UserRoleOperationsManager     UserRole = "operations_manager"
	UserRoleProcurementSpecialist UserRole = "procurement_specialist"
	UserRoleCoach                 UserRole = "coach"
	UserRoleMember                UserRole = "member"
)

func AllUserRoles() []UserRole {
	return []UserRole{
		UserRoleAdministrator,
		UserRoleOperationsManager,
		UserRoleProcurementSpecialist,
		UserRoleCoach,
		UserRoleMember,
	}
}

func (r UserRole) IsValid() bool {
	for _, v := range AllUserRoles() {
		if r == v {
			return true
		}
	}
	return false
}

func (r UserRole) IsStaff() bool {
	return r == UserRoleAdministrator || r == UserRoleOperationsManager || r == UserRoleProcurementSpecialist
}

func (r UserRole) CanAccessDashboard() bool {
	return r == UserRoleAdministrator || r == UserRoleOperationsManager || r == UserRoleProcurementSpecialist || r == UserRoleCoach
}

// --- UserStatus ---

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusLocked   UserStatus = "locked"
)

func AllUserStatuses() []UserStatus {
	return []UserStatus{
		UserStatusActive,
		UserStatusInactive,
		UserStatusLocked,
	}
}

func (s UserStatus) IsValid() bool {
	for _, v := range AllUserStatuses() {
		if s == v {
			return true
		}
	}
	return false
}

// --- ItemCondition ---

type ItemCondition string

const (
	ItemConditionNew     ItemCondition = "new"
	ItemConditionOpenBox ItemCondition = "open_box"
	ItemConditionUsed    ItemCondition = "used"
)

func AllItemConditions() []ItemCondition {
	return []ItemCondition{
		ItemConditionNew,
		ItemConditionOpenBox,
		ItemConditionUsed,
	}
}

func (c ItemCondition) IsValid() bool {
	for _, v := range AllItemConditions() {
		if c == v {
			return true
		}
	}
	return false
}

// --- BillingModel ---

type BillingModel string

const (
	BillingModelOneTime       BillingModel = "one_time"
	BillingModelMonthlyRental BillingModel = "monthly_rental"
)

func AllBillingModels() []BillingModel {
	return []BillingModel{
		BillingModelOneTime,
		BillingModelMonthlyRental,
	}
}

func (b BillingModel) IsValid() bool {
	for _, v := range AllBillingModels() {
		if b == v {
			return true
		}
	}
	return false
}

// --- ItemStatus ---

type ItemStatus string

const (
	ItemStatusDraft       ItemStatus = "draft"
	ItemStatusPublished   ItemStatus = "published"
	ItemStatusUnpublished ItemStatus = "unpublished"
)

func AllItemStatuses() []ItemStatus {
	return []ItemStatus{
		ItemStatusDraft,
		ItemStatusPublished,
		ItemStatusUnpublished,
	}
}

func (s ItemStatus) IsValid() bool {
	for _, v := range AllItemStatuses() {
		if s == v {
			return true
		}
	}
	return false
}

// --- OrderStatus ---

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "created"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusRefunded  OrderStatus = "refunded"
	OrderStatusAutoClosed OrderStatus = "auto_closed"
)

func AllOrderStatuses() []OrderStatus {
	return []OrderStatus{
		OrderStatusCreated,
		OrderStatusPaid,
		OrderStatusCancelled,
		OrderStatusRefunded,
		OrderStatusAutoClosed,
	}
}

func (s OrderStatus) IsValid() bool {
	for _, v := range AllOrderStatuses() {
		if s == v {
			return true
		}
	}
	return false
}

func (s OrderStatus) IsTerminal() bool {
	return s == OrderStatusCancelled || s == OrderStatusRefunded || s == OrderStatusAutoClosed
}

// --- CampaignStatus ---

type CampaignStatus string

const (
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusSucceeded CampaignStatus = "succeeded"
	CampaignStatusFailed    CampaignStatus = "failed"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

func AllCampaignStatuses() []CampaignStatus {
	return []CampaignStatus{
		CampaignStatusActive,
		CampaignStatusSucceeded,
		CampaignStatusFailed,
		CampaignStatusCancelled,
	}
}

func (s CampaignStatus) IsValid() bool {
	for _, v := range AllCampaignStatuses() {
		if s == v {
			return true
		}
	}
	return false
}

func (s CampaignStatus) IsTerminal() bool {
	return s == CampaignStatusSucceeded || s == CampaignStatusFailed || s == CampaignStatusCancelled
}

// --- POStatus (PurchaseOrderStatus) ---

type POStatus string

const (
	POStatusCreated  POStatus = "created"
	POStatusApproved POStatus = "approved"
	POStatusReceived POStatus = "received"
	POStatusReturned POStatus = "returned"
	POStatusVoided   POStatus = "voided"
)

func AllPOStatuses() []POStatus {
	return []POStatus{
		POStatusCreated,
		POStatusApproved,
		POStatusReceived,
		POStatusReturned,
		POStatusVoided,
	}
}

func (s POStatus) IsValid() bool {
	for _, v := range AllPOStatuses() {
		if s == v {
			return true
		}
	}
	return false
}

func (s POStatus) IsTerminal() bool {
	return s == POStatusReturned || s == POStatusVoided
}

// --- VarianceType ---

type VarianceType string

const (
	VarianceTypeShortage        VarianceType = "shortage"
	VarianceTypeOverage         VarianceType = "overage"
	VarianceTypePriceDifference VarianceType = "price_difference"
)

func AllVarianceTypes() []VarianceType {
	return []VarianceType{
		VarianceTypeShortage,
		VarianceTypeOverage,
		VarianceTypePriceDifference,
	}
}

func (t VarianceType) IsValid() bool {
	for _, v := range AllVarianceTypes() {
		if t == v {
			return true
		}
	}
	return false
}

// --- VarianceStatus ---

type VarianceStatus string

const (
	VarianceStatusOpen      VarianceStatus = "open"
	VarianceStatusResolved  VarianceStatus = "resolved"
	VarianceStatusEscalated VarianceStatus = "escalated"
)

func AllVarianceStatuses() []VarianceStatus {
	return []VarianceStatus{
		VarianceStatusOpen,
		VarianceStatusResolved,
		VarianceStatusEscalated,
	}
}

func (s VarianceStatus) IsValid() bool {
	for _, v := range AllVarianceStatuses() {
		if s == v {
			return true
		}
	}
	return false
}

// --- MembershipStatus ---

type MembershipStatus string

const (
	MembershipStatusActive    MembershipStatus = "active"
	MembershipStatusExpired   MembershipStatus = "expired"
	MembershipStatusCancelled MembershipStatus = "cancelled"
	MembershipStatusSuspended MembershipStatus = "suspended"
)

func AllMembershipStatuses() []MembershipStatus {
	return []MembershipStatus{
		MembershipStatusActive,
		MembershipStatusExpired,
		MembershipStatusCancelled,
		MembershipStatusSuspended,
	}
}

func (s MembershipStatus) IsValid() bool {
	for _, v := range AllMembershipStatuses() {
		if s == v {
			return true
		}
	}
	return false
}

// --- ExportFormat ---

type ExportFormat string

const (
	ExportFormatCSV ExportFormat = "csv"
	ExportFormatPDF ExportFormat = "pdf"
)

func AllExportFormats() []ExportFormat {
	return []ExportFormat{
		ExportFormatCSV,
		ExportFormatPDF,
	}
}

func (f ExportFormat) IsValid() bool {
	for _, v := range AllExportFormats() {
		if f == v {
			return true
		}
	}
	return false
}

// --- ExportStatus ---

type ExportStatus string

const (
	ExportStatusPending    ExportStatus = "pending"
	ExportStatusProcessing ExportStatus = "processing"
	ExportStatusCompleted  ExportStatus = "completed"
	ExportStatusFailed     ExportStatus = "failed"
)

func AllExportStatuses() []ExportStatus {
	return []ExportStatus{
		ExportStatusPending,
		ExportStatusProcessing,
		ExportStatusCompleted,
		ExportStatusFailed,
	}
}

func (s ExportStatus) IsValid() bool {
	for _, v := range AllExportStatuses() {
		if s == v {
			return true
		}
	}
	return false
}

// --- BackupStatus ---

type BackupStatus string

const (
	BackupStatusRunning   BackupStatus = "running"
	BackupStatusCompleted BackupStatus = "completed"
	BackupStatusFailed    BackupStatus = "failed"
)

func AllBackupStatuses() []BackupStatus {
	return []BackupStatus{
		BackupStatusRunning,
		BackupStatusCompleted,
		BackupStatusFailed,
	}
}

func (s BackupStatus) IsValid() bool {
	for _, v := range AllBackupStatuses() {
		if s == v {
			return true
		}
	}
	return false
}

// --- EncryptionKeyStatus ---

type EncryptionKeyStatus string

const (
	EncryptionKeyStatusActive  EncryptionKeyStatus = "active"
	EncryptionKeyStatusRotated EncryptionKeyStatus = "rotated"
	EncryptionKeyStatusRevoked EncryptionKeyStatus = "revoked"
)

func AllEncryptionKeyStatuses() []EncryptionKeyStatus {
	return []EncryptionKeyStatus{
		EncryptionKeyStatusActive,
		EncryptionKeyStatusRotated,
		EncryptionKeyStatusRevoked,
	}
}

func (s EncryptionKeyStatus) IsValid() bool {
	for _, v := range AllEncryptionKeyStatuses() {
		if s == v {
			return true
		}
	}
	return false
}
