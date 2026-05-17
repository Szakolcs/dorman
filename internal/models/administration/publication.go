package models

import (
	"time"

	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type PublicationState = platform.PublicationState

const (
	PublicationStateDraft     = platform.PublicationStateDraft
	PublicationStatePublished = platform.PublicationStatePublished
	PublicationStateArchived  = platform.PublicationStateArchived
	PublicationStateCanceled  = platform.PublicationStateCanceled
	PublicationStatePostponed = platform.PublicationStatePostponed
	PublicationStateHidden    = platform.PublicationStateHidden
)

type Activity struct {
	platform.BaseModel
	Title       string           `gorm:"not null;index"`
	Description string           `gorm:"type:text"`
	BuildingID  *uuid.UUID       `gorm:"type:uuid;index"`
	Location    string           `gorm:"not null"`
	Capacity    int              `gorm:"not null;default:1"`
	State       PublicationState `gorm:"type:varchar(20);not null;default:'draft';index"`

	Building *Building `gorm:"foreignKey:BuildingID;references:ID"`
}

type Event struct {
	platform.BaseModel
	Title           string           `gorm:"not null;index"`
	Description     string           `gorm:"type:text"`
	BuildingID      *uuid.UUID       `gorm:"type:uuid;index"`
	OrganizerUserID uuid.UUID        `gorm:"type:uuid;not null;index"`
	Location        string           `gorm:"not null"`
	StartsAt        time.Time        `gorm:"not null;index"`
	EndsAt          time.Time        `gorm:"not null;index"`
	Capacity        int              `gorm:"not null;default:1"`
	State           PublicationState `gorm:"type:varchar(20);not null;default:'draft';index"`

	Building  *Building `gorm:"foreignKey:BuildingID;references:ID"`
	Organizer User      `gorm:"foreignKey:OrganizerUserID;references:ID"`
}
