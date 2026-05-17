package models

import (
	adm "dorm-man/internal/models/administration"
	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type AttendanceIntent string

const (
	AttendanceIntentGoing    AttendanceIntent = "going"
	AttendanceIntentNotGoing AttendanceIntent = "not_going"
)

// ForumAttendanceIntent records a tenant's going / not-going intent (see poll.go for poll votes).
type ForumAttendanceIntent struct {
	platform.BaseModel
	PostID   uuid.UUID        `gorm:"type:uuid;not null;index;uniqueIndex:idx_attendance_post_tenant"`
	TenantID uuid.UUID        `gorm:"type:uuid;not null;index;uniqueIndex:idx_attendance_post_tenant"`
	Intent   AttendanceIntent `gorm:"type:varchar(20);not null"`

	Post   ForumPost  `gorm:"foreignKey:PostID;references:ID"`
	Tenant adm.Tenant `gorm:"foreignKey:TenantID;references:ID"`
}
