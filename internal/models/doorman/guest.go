package models

import (
	"time"

	adm "dorm-man/internal/models/administration"
	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type GuestVisitStatus string

const (
	GuestVisitStatusScheduled  GuestVisitStatus = "scheduled"
	GuestVisitStatusCheckedIn  GuestVisitStatus = "checked_in"
	GuestVisitStatusCheckedOut GuestVisitStatus = "checked_out"
	GuestVisitStatusDenied     GuestVisitStatus = "denied"
)

type GuestAccessEventType string

const (
	GuestAccessEventCheckIn  GuestAccessEventType = "check_in"
	GuestAccessEventCheckOut GuestAccessEventType = "check_out"
	GuestAccessEventDenied   GuestAccessEventType = "denied"
)

// Guest denial / audit reasons (domain); align with specs e.g. outside_visit_window.
const (
	GuestDenialReasonOutsideVisitWindow = "outside_visit_window"
)

type GuestVisit struct {
	platform.BaseModel
	HostTenantID       uuid.UUID        `gorm:"type:uuid;not null;index"`
	GuestName          string           `gorm:"not null;index"`
	IDNotes            string           `gorm:"type:text"`
	ValidFrom          time.Time        `gorm:"not null;index"`
	ValidTo            time.Time        `gorm:"not null;index"`
	Status             GuestVisitStatus `gorm:"type:varchar(20);not null;default:'scheduled';index"`
	RegisteredByUserID *uuid.UUID       `gorm:"type:uuid;index"`

	HostTenant       *adm.Tenant        `gorm:"foreignKey:HostTenantID;references:ID"`
	RegisteredByUser *adm.User          `gorm:"foreignKey:RegisteredByUserID;references:ID"`
	AccessEvents     []GuestAccessEvent `gorm:"foreignKey:GuestVisitID"`
}

type GuestAccessEvent struct {
	platform.BaseModel
	GuestVisitID uuid.UUID            `gorm:"type:uuid;not null;index"`
	ActorUserID  uuid.UUID            `gorm:"type:uuid;not null;index"`
	EventType    GuestAccessEventType `gorm:"type:varchar(20);not null;index"`
	OccurredAt   time.Time            `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	Reason       *string              `gorm:"type:varchar(64);index"`

	GuestVisit *GuestVisit `gorm:"foreignKey:GuestVisitID;references:ID"`
	ActorUser  *adm.User   `gorm:"foreignKey:ActorUserID;references:ID"`
}
