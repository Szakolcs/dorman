package models

import (
	"time"

	"github.com/google/uuid"
)

// Event is a forum post tied to a specific time window and (optionally) a
// shared area. It supports per-user attendance intents, free-form comments
// and is backed by its own dedicated chatroom.
type Event struct {
	Post
	SharedAreaID *uuid.UUID `gorm:"type:uuid;index"`
	StartsAt     time.Time  `gorm:"not null;index"`
	EndsAt       time.Time  `gorm:"not null;index"`
	ChatRoomID   *uuid.UUID `gorm:"type:uuid;uniqueIndex"`

	ChatRoom    *ChatRoom         `gorm:"foreignKey:ChatRoomID;references:ID"`
	Attendances []EventAttendance `gorm:"foreignKey:EventID"`
	Comments    []EventComment    `gorm:"foreignKey:EventID"`
}

// EventAttendance records a single user's intent (interested, not interested,
// or busy) for an event. Each (event, user) pair is unique so users can
// change their mind by updating the existing row.
type EventAttendance struct {
	BaseModel
	EventID uuid.UUID             `gorm:"type:uuid;not null;index;uniqueIndex:idx_event_user_attendance"`
	UserID  uuid.UUID             `gorm:"type:uuid;not null;index;uniqueIndex:idx_event_user_attendance"`
	Intent  EventAttendanceIntent `gorm:"type:varchar(20);not null;index"`

	Event *Event `gorm:"foreignKey:EventID;references:ID"`
	User  *User  `gorm:"foreignKey:UserID;references:ID"`
}

// EventComment is a free-form comment posted by a user against an Event.
type EventComment struct {
	BaseModel
	EventID  uuid.UUID `gorm:"type:uuid;not null;index"`
	AuthorID uuid.UUID `gorm:"type:uuid;not null;index"`
	Body     string    `gorm:"type:text;not null"`

	Event  *Event `gorm:"foreignKey:EventID;references:ID"`
	Author *User  `gorm:"foreignKey:AuthorID;references:ID"`
}
