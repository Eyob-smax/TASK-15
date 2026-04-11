package domain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AuditEvent represents an immutable audit log entry with chain integrity.
type AuditEvent struct {
	ID            uuid.UUID
	EventType     string
	EntityType    string
	EntityID      uuid.UUID
	ActorID       uuid.UUID
	Details       map[string]interface{}
	IntegrityHash string
	PreviousHash  string
	CreatedAt     time.Time
}

// ComputeHash computes a SHA-256 hash of the audit event's key fields
// concatenated with the previous hash in the chain. The hash covers
// event type, entity type, entity ID, actor ID, timestamp, details JSON,
// and the previous hash to form an integrity chain.
func (e *AuditEvent) ComputeHash(previousHash string) string {
	detailsJSON, err := json.Marshal(e.Details)
	if err != nil {
		detailsJSON = []byte("{}")
	}

	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		e.EventType,
		e.EntityType,
		e.EntityID.String(),
		e.ActorID.String(),
		e.CreatedAt.UTC().Format(time.RFC3339Nano),
		string(detailsJSON),
		previousHash,
	)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
