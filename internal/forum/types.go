package forum

import (
	"errors"
	"time"

	fm "dorm-man/internal/models/forum"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized        = errors.New("authorization denied")
	ErrNotFound            = errors.New("not found")
	ErrValidation          = errors.New("validation error")
	ErrPollClosed          = errors.New("poll closed")
	ErrAlreadyVoted        = errors.New("already voted")
	ErrIneligibleVoter     = errors.New("ineligible voter")
	ErrResultsNotVisible   = errors.New("results not visible")
	ErrEditNotAllowed      = errors.New("edit not allowed")
	ErrConcurrencyConflict = errors.New("concurrency conflict")
)

type FeedSort string

const (
	FeedSortNewest      FeedSort = "newest"
	FeedSortPinnedFirst FeedSort = "pinned_first"
)

type FeedListFilter struct {
	Kind          fm.ForumPostKind
	OfficialOnly  bool
	CommunityOnly bool
	Sort          FeedSort
	Limit         int
	Offset        int
	IncludeHidden bool // staff moderation reads
}

type CommentListFilter struct {
	PostID uuid.UUID
	Limit  int
	Offset int
}

type CommunityPostInput struct {
	Title    string     `json:"title"`
	Body     string     `json:"body"`
	Tags     string     `json:"tags"`
	StartsAt *time.Time `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
	Location string     `json:"location"`
	Publish  bool       `json:"publish"`
}

type CommunityPostUpdateInput struct {
	Title    string     `json:"title"`
	Body     string     `json:"body"`
	Tags     string     `json:"tags"`
	StartsAt *time.Time `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
	Location string     `json:"location"`
}

type CommentInput struct {
	Body            string     `json:"body"`
	ParentCommentID *uuid.UUID `json:"parent_comment_id"`
}

type CommentUpdateInput struct {
	Body string `json:"body"`
}

type PollCreateInput struct {
	Title             string                       `json:"title"`
	Body              string                       `json:"body"`
	ChoiceMode        fm.ForumPollChoiceMode       `json:"choice_mode"`
	ResultsVisibility fm.ForumPollResultsVisibility `json:"results_visibility"`
	ClosesAt          time.Time                    `json:"closes_at"`
	Options           []PollOptionInput            `json:"options"`
}

type PollOptionInput struct {
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
}

type PollVoteInput struct {
	OptionIDs []uuid.UUID `json:"option_ids"`
}

type AttendanceInput struct {
	Intent fm.AttendanceIntent `json:"intent"`
}

type ReactionTargetInput struct {
	TargetType fm.ReactionTargetType `json:"target_type"`
	TargetID   uuid.UUID             `json:"target_id"`
}

type ModerationInput struct {
	TargetType fm.ForumModerationTargetType `json:"target_type"`
	TargetID   uuid.UUID                    `json:"target_id"`
	Action     fm.ForumModerationActionKind `json:"action"`
	Reason     string                       `json:"reason"`
}

type PostOrganizerUpdateInput struct {
	Body string `json:"body"`
}

type PostAggregates struct {
	ReactionCount          int64                 `json:"reaction_count"`
	CommentCount           int64                 `json:"comment_count"`
	GoingCount             int64                 `json:"going_count"`
	NotGoingCount          int64                 `json:"not_going_count"`
	ViewerReaction         *fm.ReactionType      `json:"viewer_reaction,omitempty"`
	ViewerAttendanceIntent *fm.AttendanceIntent  `json:"viewer_attendance_intent,omitempty"`
}

type PostDetail struct {
	Post       fm.ForumPost   `json:"post"`
	Aggregates PostAggregates `json:"aggregates"`
}

type PollResults struct {
	Poll    fm.ForumPoll      `json:"poll"`
	Options []PollOptionCount `json:"options"`
	Visible bool              `json:"visible"`
}

type PollOptionCount struct {
	OptionID uuid.UUID `json:"option_id"`
	Label    string    `json:"label"`
	Votes    int64     `json:"votes"`
}
