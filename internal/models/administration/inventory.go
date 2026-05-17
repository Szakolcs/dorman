package models

import (
	"time"

	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type InventoryLocationType string

const (
	InventoryLocationRoom       InventoryLocationType = "room"
	InventoryLocationSharedArea InventoryLocationType = "shared_area"
)

type InventoryCondition string

const (
	InventoryConditionNew     InventoryCondition = "new"
	InventoryConditionGood    InventoryCondition = "good"
	InventoryConditionUsed    InventoryCondition = "used"
	InventoryConditionDamaged InventoryCondition = "damaged"
	InventoryConditionBroken  InventoryCondition = "broken"
)

type InventoryStatus string

const (
	InventoryStatusInStock   InventoryStatus = "in_stock"
	InventoryStatusInUse     InventoryStatus = "in_use"
	InventoryStatusWithdrawn InventoryStatus = "withdrawn"
	InventoryStatusDestroyed InventoryStatus = "destroyed"
)

type InventoryItem struct {
	platform.BaseModel
	Name         string `gorm:"not null;index"`
	Description  string `gorm:"type:text"`
	LocationType InventoryLocationType `gorm:"type:varchar(20);not null;index"`
	RoomID       *uuid.UUID            `gorm:"type:uuid;index"`
	FlatID       *uuid.UUID            `gorm:"type:uuid;index"`
	BuildingID   *uuid.UUID            `gorm:"type:uuid;index"`
	Condition    InventoryCondition    `gorm:"type:varchar(20);not null;default:'good';index"`
	Status       InventoryStatus       `gorm:"type:varchar(20);not null;default:'in_use';index"`
	PurchaseDate time.Time             `gorm:"not null;index"`
	InUseDate    *time.Time            `gorm:"index"`
	WithdrawDate *time.Time            `gorm:"index"`

	Room     *Room     `gorm:"foreignKey:RoomID;references:ID"`
	Flat     *Flat     `gorm:"foreignKey:FlatID;references:ID"`
	Building *Building `gorm:"foreignKey:BuildingID;references:ID"`
}
