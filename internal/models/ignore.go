package models

import (
	"github.com/google/uuid"
)

// Ignore is a one-directional block: the tenant identified by TenantID has
// chosen to hide messages from the tenant identified by IgnoredTenantID.
// Mutual ignores are represented by two rows. The (TenantID, IgnoredTenantID)
// pair is unique; a self-ignore is rejected at the service layer.
type Ignore struct {
	BaseModel
	TenantID        uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_ignore_pair"`
	IgnoredTenantID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_ignore_pair"`

	Tenant        *Tenant `gorm:"foreignKey:TenantID;references:ID"`
	IgnoredTenant *Tenant `gorm:"foreignKey:IgnoredTenantID;references:ID"`
}
