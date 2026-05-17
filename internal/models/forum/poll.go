package models

import (
	"time"

	adm "dorm-man/internal/models/administration"
	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type ForumPollChoiceMode string

const (
	ForumPollChoiceModeSingle   ForumPollChoiceMode = "single"
	ForumPollChoiceModeMultiple ForumPollChoiceMode = "multiple"
)

type ForumPollResultsVisibility string

const (
	ForumPollResultsVisibilityLive      ForumPollResultsVisibility = "live"
	ForumPollResultsVisibilityPostClose ForumPollResultsVisibility = "post_close"
)

type ForumPoll struct {
	platform.BaseModel
	ForumPostID       uuid.UUID                  `gorm:"type:uuid;not null;uniqueIndex"`
	ChoiceMode        ForumPollChoiceMode        `gorm:"type:varchar(20);not null"`
	ResultsVisibility ForumPollResultsVisibility `gorm:"type:varchar(20);not null"`
	ClosesAt          time.Time                  `gorm:"not null;index"`
	Eligibility       string                     `gorm:"type:text"`
	CreatedByUserID   uuid.UUID                  `gorm:"type:uuid;not null;index"`

	ForumPost ForumPost         `gorm:"foreignKey:ForumPostID;references:ID"`
	CreatedBy adm.User          `gorm:"foreignKey:CreatedByUserID;references:ID"`
	Options   []ForumPollOption `gorm:"foreignKey:PollID"`
	Votes     []ForumPollVote   `gorm:"foreignKey:PollID"`
}

type ForumPollOption struct {
	platform.BaseModel
	PollID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Label     string    `gorm:"not null"`
	SortOrder int       `gorm:"not null;default:0"`

	Poll  ForumPoll       `gorm:"foreignKey:PollID;references:ID"`
	Votes []ForumPollVote `gorm:"foreignKey:OptionID"`
}

type ForumPollVote struct {
	platform.BaseModel
	PollID   uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_forum_poll_vote"`
	OptionID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_forum_poll_vote"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_forum_poll_vote"`
	VotedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index"`

	Poll   ForumPoll       `gorm:"foreignKey:PollID;references:ID"`
	Option ForumPollOption `gorm:"foreignKey:OptionID;references:ID"`
	Tenant adm.Tenant      `gorm:"foreignKey:TenantID;references:ID"`
}
