package models

import (
	"time"

	"github.com/google/uuid"
)

type InventoryItem struct {
	BaseModel
	Name         string             `gorm:"not null;index"`
	Description  string             `gorm:"type:text"`
	RoomID       *uuid.UUID         `gorm:"type:uuid;index"`
	FlatID       *uuid.UUID         `gorm:"type:uuid;index"`
	BuildingID   *uuid.UUID         `gorm:"type:uuid;index"`
	SharedAreaID *uuid.UUID         `gorm:"type:uuid;index"`
	Condition    InventoryCondition `gorm:"type:varchar(20);not null;default:'good';index"`
	Status       InventoryStatus    `gorm:"type:varchar(20);not null;default:'in_use';index"`
	PurchaseDate time.Time          `gorm:"not null;index"`
	InUseDate    *time.Time         `gorm:"index"`
	WithdrawDate *time.Time         `gorm:"index"`

	Room       *Room       `gorm:"foreignKey:RoomID;references:ID"`
	Flat       *Flat       `gorm:"foreignKey:FlatID;references:ID"`
	Building   *Building   `gorm:"foreignKey:BuildingID;references:ID"`
	SharedArea *SharedArea `gorm:"foreignKey:SharedAreaID;references:ID"`
}
