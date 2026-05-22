package models

import (
	"github.com/google/uuid"
)

type Building struct {
	BaseModel
	Name string `gorm:"not null;uniqueIndex"`
	Code string `gorm:"not null;uniqueIndex"`

	Flats       []Flat       `gorm:"foreignKey:BuildingID"`
	SharedAreas []SharedArea `gorm:"foreignKey:BuildingID"`
}

type Flat struct {
	BaseModel
	BuildingID *uuid.UUID `gorm:"type:uuid;index"`
	Name       string     `gorm:"not null;index"`
	Floor      int        `gorm:"not null;default:0"`

	Building           *Building       `gorm:"foreignKey:BuildingID;references:ID"`
	Rooms              []Room          `gorm:"foreignKey:FlatID"`
	Inventory          []InventoryItem `gorm:"foreignKey:FlatID"`
	MaintenanceTickets []Ticket        `gorm:"foreignKey:FlatID"`
}

type Room struct {
	BaseModel
	FlatID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Number   string    `gorm:"not null;index"`
	Capacity int       `gorm:"not null;default:1"`

	Flat        Flat             `gorm:"foreignKey:FlatID;references:ID"`
	Assignments []RoomAssignment `gorm:"foreignKey:RoomID"`
}

// SharedArea is intentionally agnostic about the forum/chat domains: the
// Activity/Event tables in `forum` already carry SharedAreaID foreign keys,
// so reverse relations live on the forum side and can be queried with
// db.Where("shared_area_id = ?", id).
type SharedArea struct {
	BaseModel
	BuildingID *uuid.UUID `gorm:"type:uuid"`
	Name       string     `gorm:"not null;uniqueIndex"`
	Code       string     `gorm:"not null;uniqueIndex"`

	Building  *Building       `gorm:"foreignKey:BuildingID;references:ID"`
	Inventory []InventoryItem `gorm:"foreignKey:SharedAreaID"`
}
