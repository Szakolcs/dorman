package models

import (
	"time"

	adm "dorm-man/internal/models/administration"
	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type ForumPostKind string

const (
	ForumPostKindOfficialNews      ForumPostKind = "official_news"
	ForumPostKindCommunityActivity ForumPostKind = "community_activity"
	ForumPostKindEvent             ForumPostKind = "event"
	ForumPostKindPoll              ForumPostKind = "poll"
	ForumPostKindAnnouncement      ForumPostKind = "announcement"
	ForumPostKindSystemNotice      ForumPostKind = "system_notice"
)

type ForumPostState string

const (
	ForumPostStateDraft     ForumPostState = "draft"
	ForumPostStatePublished ForumPostState = "published"
	ForumPostStateArchived  ForumPostState = "archived"
	ForumPostStateCanceled  ForumPostState = "canceled"
	ForumPostStatePostponed ForumPostState = "postponed"
	ForumPostStateHidden    ForumPostState = "hidden"
)

type ForumPostSource string

const (
	ForumPostSourceForum          ForumPostSource = "forum"
	ForumPostSourceAdministration ForumPostSource = "administration"
)

type ForumPost struct {
	platform.BaseModel
	AuthorUserID   uuid.UUID       `gorm:"type:uuid;not null;index"`
	AuthorTenantID *uuid.UUID      `gorm:"type:uuid;index"`
	ActivityID     *uuid.UUID      `gorm:"type:uuid;index"`
	EventID        *uuid.UUID      `gorm:"type:uuid;index"`
	Kind           ForumPostKind   `gorm:"type:varchar(30);not null;index"`
	Title          string          `gorm:"not null"`
	Body           string          `gorm:"type:text;not null"`
	State          ForumPostState  `gorm:"type:varchar(20);not null;default:'draft';index"`
	Source         ForumPostSource `gorm:"type:varchar(20);not null;default:'forum';index"`
	PublishedAt    *time.Time      `gorm:"index"`
	PinnedAt       *time.Time      `gorm:"index"`
	PinPriority    int             `gorm:"not null;default:0"`
	Tags           string          `gorm:"type:text"`

	AuthorUser   adm.User      `gorm:"foreignKey:AuthorUserID;references:ID"`
	AuthorTenant *adm.Tenant   `gorm:"foreignKey:AuthorTenantID;references:ID"`
	Activity     *adm.Activity `gorm:"foreignKey:ActivityID;references:ID"`
	Event        *adm.Event    `gorm:"foreignKey:EventID;references:ID"`
	Schedule     *ForumPostSchedule
	Poll         *ForumPoll
	Comments     []ForumComment          `gorm:"foreignKey:PostID"`
	Updates      []ForumPostUpdate       `gorm:"foreignKey:ParentPostID"`
	Attendance   []ForumAttendanceIntent `gorm:"foreignKey:PostID"`
	Views        []ForumPostView         `gorm:"foreignKey:PostID"`
}

type ForumPostSchedule struct {
	platform.BaseModel
	ForumPostID          uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex"`
	StartsAt             *time.Time `gorm:"index"`
	EndsAt               *time.Time
	Location             string
	Capacity             *int
	RegistrationDeadline *time.Time

	ForumPost ForumPost `gorm:"foreignKey:ForumPostID;references:ID"`
}

type ForumPostUpdate struct {
	platform.BaseModel
	ParentPostID uuid.UUID `gorm:"type:uuid;not null;index"`
	AuthorUserID uuid.UUID `gorm:"type:uuid;not null;index"`
	Body         string    `gorm:"type:text;not null"`

	ParentPost ForumPost `gorm:"foreignKey:ParentPostID;references:ID"`
	AuthorUser adm.User  `gorm:"foreignKey:AuthorUserID;references:ID"`
}
