package models

import (
	"time"

	adm "dorm-man/internal/models/administration"
	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type ForumModerationTargetType string

const (
	ForumModerationTargetPost    ForumModerationTargetType = "post"
	ForumModerationTargetComment ForumModerationTargetType = "comment"
)

type ForumModerationActionKind string

const (
	ForumModerationActionHide    ForumModerationActionKind = "hide"
	ForumModerationActionRestore ForumModerationActionKind = "restore"
	ForumModerationActionRemove  ForumModerationActionKind = "remove"
)

type ForumModerationAction struct {
	platform.BaseModel
	ModeratorUserID uuid.UUID                 `gorm:"type:uuid;not null;index"`
	TargetType      ForumModerationTargetType `gorm:"type:varchar(20);not null;index:idx_forum_moderation_target,priority:1"`
	TargetID        uuid.UUID                 `gorm:"type:uuid;not null;index:idx_forum_moderation_target,priority:2"`
	Action          ForumModerationActionKind `gorm:"type:varchar(20);not null;index"`
	Reason          string                    `gorm:"type:text"`
	OccurredAt      time.Time                 `gorm:"not null;default:CURRENT_TIMESTAMP;index"`

	ModeratorUser adm.User `gorm:"foreignKey:ModeratorUserID;references:ID"`
}
