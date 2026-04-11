package domain

import (
	"time"

	"github.com/google/uuid"
)

// Member represents a gym/fitness club member.
type Member struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	LocationID       uuid.UUID
	MembershipStatus MembershipStatus
	JoinedAt         time.Time
	RenewalDate      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Coach represents a fitness coach assigned to a location.
type Coach struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	LocationID     uuid.UUID
	Specialization string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
