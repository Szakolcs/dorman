package models

import (
	adm "dorm-man/internal/models/administration"
	crosscutting "dorm-man/internal/models/cross-cutting"

	"github.com/google/uuid"
)

type GuestEntry struct {
	crosscutting.BaseModel
	HostTenantID uuid.UUID `gorm:"type:uuid;not null;index"`
	GuestName    string    `gorm:"not null;index"`
	IDNotes      string    `gorm:"type:text"`

	HostTenant *adm.Tenant `gorm:"foreignKey:HostTenantID;references:ID"`
}
