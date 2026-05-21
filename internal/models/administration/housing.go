package models

import (
	platform "dorm-man/internal/models/cross-cutting"

	"github.com/google/uuid"
)

type Building struct {
	platform.BaseModel
	Name string `gorm:"not null;uniqueIndex"`
	Code string `gorm:"not null;uniqueIndex"`

	Flats       []Flat       `gorm:"foreignKey:BuildingID"`
	SharedAreas []SharedArea `gorm:"foreignKey:BuildingID"`
}

type Flat struct {
	platform.BaseModel
	BuildingID *uuid.UUID `gorm:"type:uuid;index"`
	Name       string     `gorm:"not null;index"`
	Floor      int        `gorm:"not null;default:0"`

	Building           *Building           `gorm:"foreignKey:BuildingID;references:ID"`
	Rooms              []Room              `gorm:"foreignKey:FlatID"`
	Inventory          []InventoryItem     `gorm:"foreignKey:FlatID"`
	MaintenanceTickets []MaintenanceTicket `gorm:"foreignKey:FlatID"`
}

type Room struct {
	platform.BaseModel
	FlatID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Number   string    `gorm:"not null;index"`
	Capacity int       `gorm:"not null;default:1"`

	Flat        Flat             `gorm:"foreignKey:FlatID;references:ID"`
	Assignments []RoomAssignment `gorm:"foreignKey:RoomID"`
}

type SharedArea struct {
	platform.BaseModel
	BuildingID *uuid.UUID `gorm:"type:uuid;"`
	Name       string     `gorm:"not null;uniqueIndex"`
	Code       string     `gorm:"not null;uniqueIndex"`

	Building   *Building       `gorm:"foreignKey:BuildingID;references:ID"`
	Activities []Activity      `gorm:"foreignKey:SharedAreaID"`
	Events     []Event         `gorm:"foreignKey:SharedAreaID"`
	Inventory  []InventoryItem `gorm:"foreignKey:SharedAreaID"`
}
