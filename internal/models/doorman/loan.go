package models

import (
	"time"

	adm "dorm-man/internal/models/administration"
	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type ItemLoan struct {
	platform.BaseModel
	TenantID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	InventoryItemID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	CheckedOutByUserID uuid.UUID  `gorm:"type:uuid;not null;index"`
	ReturnedByUserID   *uuid.UUID `gorm:"type:uuid;index"`
	CheckedOutAt       time.Time  `gorm:"not null;index"`
	ExpectedReturnAt   *time.Time `gorm:"index"`
	ReturnedAt         *time.Time `gorm:"index"`
	Notes              string     `gorm:"type:text"`

	Tenant           *adm.Tenant        `gorm:"foreignKey:TenantID;references:ID"`
	InventoryItem    *adm.InventoryItem `gorm:"foreignKey:InventoryItemID;references:ID"`
	CheckedOutByUser *adm.User          `gorm:"foreignKey:CheckedOutByUserID;references:ID"`
	ReturnedByUser   *adm.User          `gorm:"foreignKey:ReturnedByUserID;references:ID"`
}
