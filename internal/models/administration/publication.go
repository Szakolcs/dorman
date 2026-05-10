package models

import (
	"time"

	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type PublicationState string

const (
	PublicationStateDraft     PublicationState = "draft"
	PublicationStatePublished PublicationState = "published"
	PublicationStateArchived  PublicationState = "archived"
	PublicationStateCanceled  PublicationState = "canceled"
	PublicationStatePostponed PublicationState = "postponed"
)

type ForumPostKind string

const (
	ForumPostKindOfficialNews ForumPostKind = "official_news"
)

type ForumPost struct {
	platform.BaseModel
	AuthorUserID uuid.UUID        `gorm:"type:uuid;not null;index"`
	Kind         ForumPostKind    `gorm:"type:varchar(30);not null;index"`
	State        PublicationState `gorm:"type:varchar(20);not null;default:'draft';index"`
	Title        string           `gorm:"not null"`
	Body         string           `gorm:"type:text;not null"`
	Tags         string
	PublishedAt  *time.Time `gorm:"index"`

	AuthorUser User `gorm:"foreignKey:AuthorUserID;references:ID"`
}

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
