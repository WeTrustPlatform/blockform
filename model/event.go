package model

import (
	"time"
)

// Event types
const (
	Issue string = "issue"
	Fine  string = "fine"
)

// Event is something that happened to a node
type Event struct {
	ID          uint64 `gorm:"primary_key"`
	NodeID      uint64
	Type        string
	Title       string `gorm:"not null"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
