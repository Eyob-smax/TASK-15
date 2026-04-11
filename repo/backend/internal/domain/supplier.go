package domain

import (
	"time"

	"github.com/google/uuid"
)

// Supplier represents an external supplier of equipment and products.
type Supplier struct {
	ID           uuid.UUID
	Name         string
	ContactName  string
	ContactEmail string
	ContactPhone string
	Address      string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
