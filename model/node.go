package model

import (
	"time"
)

// Node states
const (
	Creating string = "creating"
	Deployed string = "deployed"
	Deleting string = "deleting"
)

// Node network types
const (
	Public  string = "public"
	Private string = "private"
)

// Node is a VM hosting a geth instance
type Node struct {
	ID            uint64 `gorm:"primary_key"`
	Name          string `gorm:"not null;unique"`
	VMID          string
	CloudProvider string
	NetworkType   string `gorm:"not null"`
	NetworkID     uint64
	APIKey        string
	Status        string `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
