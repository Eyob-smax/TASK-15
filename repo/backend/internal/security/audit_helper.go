package security

import (
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
)

// Auth and session audit event type constants.
const (
	EventLoginSuccess    = "auth.login.success"
	EventLoginFailure    = "auth.login.failure"
	EventLoginLockout    = "auth.login.lockout"
	EventLogout          = "auth.logout"
	EventSessionExpired  = "auth.session.expired"
	EventCaptchaVerified = "auth.captcha.verified"
	EventCaptchaFailed   = "auth.captcha.failed"
)

// Catalog and inventory audit event type constants (Prompt 4).
const (
	EventItemCreated     = "item.created"
	EventItemUpdated     = "item.updated"
	EventItemPublished   = "item.published"
	EventItemUnpublished = "item.unpublished"
	EventItemBatchEdit   = "item.batch_edit"

	EventInventoryAdjusted = "inventory.adjusted"

	EventCampaignCreated   = "campaign.created"
	EventCampaignJoined    = "campaign.joined"
	EventCampaignEvaluated = "campaign.evaluated"
	EventCampaignCancelled = "campaign.cancelled"

	EventOrderCreated    = "order.created"
	EventOrderPaid       = "order.paid"
	EventOrderCancelled  = "order.cancelled"
	EventOrderRefunded   = "order.refunded"
	EventOrderAutoClosed = "order.auto_closed"
	EventOrderNoteAdded  = "order.note_added"
	EventOrderSplit      = "order.split"
	EventOrderMerged     = "order.merged"
)

// Procurement audit event type constants (Prompt 7).
const (
	EventSupplierCreated = "supplier.created"
	EventSupplierUpdated = "supplier.updated"

	EventPOCreated  = "po.created"
	EventPOApproved = "po.approved"
	EventPOReceived = "po.received"
	EventPOReturned = "po.returned"
	EventPOVoided   = "po.voided"

	EventVarianceResolved = "variance.resolved"
)

// Admin, backup, retention, biometric, user audit event type constants (Prompt 7).
const (
	EventUserCreated      = "user.created"
	EventUserUpdated      = "user.updated"
	EventUserDeactivated  = "user.deactivated"

	EventBackupCompleted = "backup.completed"

	EventRetentionPolicyUpdated = "retention.policy.updated"

	EventBiometricRegistered = "biometric.registered"
	EventBiometricRevoked    = "biometric.revoked"

	EventEncryptionKeyRotated = "encryption_key.rotated"

	EventExportGenerated = "export.generated"
)

// sensitiveKeys is the set of map keys that must never appear in audit event
// details. Any key in this set is stripped by RedactSensitiveFields.
var sensitiveKeys = map[string]bool{
	"password":       true,
	"password_hash":  true,
	"salt":           true,
	"token":          true,
	"session_token":  true,
	"answer":         true,
	"note":           true,
	"notes":          true,
	"key":            true,
	"secret":         true,
	"encrypted_data": true,
}

// RedactSensitiveFields returns a copy of details with all sensitive keys removed.
// Call this before passing details to BuildAuditEvent or any log helper.
func RedactSensitiveFields(details map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(details))
	for k, v := range details {
		if !sensitiveKeys[k] {
			out[k] = v
		}
	}
	return out
}

// SafeDetails is an alias for RedactSensitiveFields provided for clarity at
// call sites: SafeDetails(details) makes the intent explicit.
func SafeDetails(details map[string]interface{}) map[string]interface{} {
	return RedactSensitiveFields(details)
}

// BuildAuditEvent constructs a domain.AuditEvent and computes its integrity
// hash using the supplied previousHash from the last event in the chain.
// Pass an empty string as previousHash when there are no prior events.
// Always call SafeDetails on the details map before passing it here.
func BuildAuditEvent(
	eventType, entityType string,
	entityID, actorID uuid.UUID,
	details map[string]interface{},
	previousHash string,
) domain.AuditEvent {
	event := domain.AuditEvent{
		ID:           uuid.New(),
		EventType:    eventType,
		EntityType:   entityType,
		EntityID:     entityID,
		ActorID:      actorID,
		Details:      details,
		PreviousHash: previousHash,
		CreatedAt:    time.Now().UTC(),
	}
	event.IntegrityHash = event.ComputeHash(previousHash)
	return event
}
