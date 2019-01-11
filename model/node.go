package model

import (
	"time"
)

// Node states
const (
	Creating string = "creating"
	Deployed string = "deployed"
	Deleting string = "deleting"
	Error    string = "error"
)

// Node network types
const (
	Public  string = "public"
	Private string = "private"
)

// Node sync modes
const (
	Full  string = "full"
	Fast  string = "fast"
	Light string = "light"
)

// Node is a VM hosting a geth instance
type Node struct {
	ID            uint64 `gorm:"primary_key"`
	Name          string `gorm:"not null;unique"`
	VMID          string
	CloudProvider string
	DomainName    string
	HasSSL        bool
	NetworkType   string `gorm:"not null"`
	NetworkID     uint64
	SyncMode      string
	APIKey        string
	Status        string `gorm:"not null"`
	Events        []Event
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
