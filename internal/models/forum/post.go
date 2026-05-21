package models

import (
	"dorm-man/internal/models/cross-cutting"
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

// ForumPostState matches administration PublicationState (platform.PublicationState).
type ForumPostState = platform.PublicationState

const (
	ForumPostStateDraft     = platform.PublicationStateDraft
	ForumPostStatePublished = platform.PublicationStatePublished
	ForumPostStateArchived  = platform.PublicationStateArchived
	ForumPostStateCanceled  = platform.PublicationStateCanceled
	ForumPostStatePostponed = platform.PublicationStatePostponed
	ForumPostStateHidden    = platform.PublicationStateHidden
)

type ForumPostSource string

const (
	ForumPostSourceForum          ForumPostSource = "forum"
	ForumPostSourceAdministration ForumPostSource = "administration"
)

// ForumPost author fields: AuthorUserID is always the authenticated user (staff or linked tenant user).
// AuthorTenantID is set for tenant-authored community content (organizer identity in the dorm).
type ForumPost struct {
	cross_cutting.BaseModel
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
	cross_cutting.BaseModel
	ForumPostID          uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex"`
	StartsAt             *time.Time `gorm:"index"`
	EndsAt               *time.Time `gorm:"index"`
	Location             string
	Capacity             *int
	RegistrationDeadline *time.Time

	ForumPost ForumPost `gorm:"foreignKey:ForumPostID;references:ID"`
}

type ForumPostUpdate struct {
	cross_cutting.BaseModel
	ParentPostID uuid.UUID `gorm:"type:uuid;not null;index"`
	AuthorUserID uuid.UUID `gorm:"type:uuid;not null;index"`
	Body         string    `gorm:"type:text;not null"`

	ParentPost ForumPost `gorm:"foreignKey:ParentPostID;references:ID"`
	AuthorUser adm.User  `gorm:"foreignKey:AuthorUserID;references:ID"`
}
