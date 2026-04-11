package domain

import (
	"time"

	"github.com/google/uuid"
)

// Location represents a physical club or facility location.
type Location struct {
	ID        uuid.UUID
	Name      string
	Address   string
	Timezone  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
