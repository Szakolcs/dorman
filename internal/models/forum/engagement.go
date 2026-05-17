package models

import (
	"time"

	adm "dorm-man/internal/models/administration"
	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type CommentModerationState string

const (
	CommentModerationStateVisible CommentModerationState = "visible"
	CommentModerationStateHidden  CommentModerationState = "hidden"
	CommentModerationStateRemoved CommentModerationState = "removed"
)

type ReactionTargetType string

const (
	ReactionTargetPost    ReactionTargetType = "post"
	ReactionTargetComment ReactionTargetType = "comment"
)

type ReactionType string

const (
	ReactionTypeLike ReactionType = "like"
)

// ForumComment and ForumReaction use AuthorUserID / UserID (authenticated user), not tenant ID.
type ForumComment struct {
	platform.BaseModel
	PostID          uuid.UUID  `gorm:"type:uuid;not null;index"`
	AuthorUserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	ParentCommentID *uuid.UUID `gorm:"type:uuid;index"`
	Body            string     `gorm:"type:text;not null"`
	EditedAt        *time.Time
	ModerationState CommentModerationState `gorm:"type:varchar(20);not null;default:'visible';index"`

	Post          ForumPost      `gorm:"foreignKey:PostID;references:ID"`
	AuthorUser    adm.User       `gorm:"foreignKey:AuthorUserID;references:ID"`
	ParentComment *ForumComment  `gorm:"foreignKey:ParentCommentID;references:ID"`
	Replies       []ForumComment `gorm:"foreignKey:ParentCommentID"`
}

type ForumReaction struct {
	platform.BaseModel
	UserID       uuid.UUID          `gorm:"type:uuid;not null;index;uniqueIndex:idx_forum_reaction_unique"`
	TargetType   ReactionTargetType `gorm:"type:varchar(20);not null;uniqueIndex:idx_forum_reaction_unique"`
	TargetID     uuid.UUID          `gorm:"type:uuid;not null;index;uniqueIndex:idx_forum_reaction_unique"`
	ReactionType ReactionType       `gorm:"type:varchar(20);not null;default:'like';uniqueIndex:idx_forum_reaction_unique"`

	User adm.User `gorm:"foreignKey:UserID;references:ID"`
}
