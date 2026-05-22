package models

import "github.com/google/uuid"

// Post is the shared base for every forum post type: Publication, Event, Activity.
// It is meant to be embedded; each subtype has its own table.
type Post struct {
	BaseModel
	Title       string           `gorm:"not null;index"`
	Description string           `gorm:"type:text"`
	State       PublicationState `gorm:"type:varchar(20);not null;default:'draft';index"`
	AuthorID    uuid.UUID        `gorm:"type:uuid;not null;index"`

	Author *User `gorm:"foreignKey:AuthorID;references:ID"`
}
