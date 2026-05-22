package models

import (
	"github.com/google/uuid"
)

type GuestEntry struct {
	BaseModel
	HostTenantID uuid.UUID `gorm:"type:uuid;not null;index"`
	GuestName    string    `gorm:"not null;index"`
	IDNotes      string    `gorm:"type:text"`

	HostTenant *Tenant `gorm:"foreignKey:HostTenantID;references:ID"`
}
