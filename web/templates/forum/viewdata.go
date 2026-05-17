package forum

import (
	"dorm-man/internal/pagination"
	fm "dorm-man/internal/models/forum"

	"github.com/google/uuid"
)

type FeedPageData struct {
	Posts         []fm.ForumPost
	Sort          string
	Kind          string
	OfficialOnly  bool
	CommunityOnly bool
	ActorUserID   string
	Pagination    pagination.Meta
	Preserve      map[string]string
}

type PostAggregatesView struct {
	ReactionCount          int64
	CommentCount           int64
	GoingCount             int64
	NotGoingCount          int64
	ViewerLiked            bool
	ViewerAttendanceIntent *fm.AttendanceIntent
}

type PollOptionRowView struct {
	Label string
	Votes int64
}

type PollResultsView struct {
	Options []PollOptionRowView
}

type PostPageData struct {
	Post            fm.ForumPost
	Aggregates      PostAggregatesView
	Schedule        *fm.ForumPostSchedule
	Comments        []fm.ForumComment
	CommentsPagination pagination.Meta
	CommentsPreserve   map[string]string
	Poll            *fm.ForumPoll
	PollResults     *PollResultsView
	PollResultsOK   bool
	Updates         []fm.ForumPostUpdate
	ActorUserID     string
}

type ReactionViewData struct {
	PostID      uuid.UUID
	Count       int64
	ViewerLiked bool
}

type AttendanceViewData struct {
	PostID     uuid.UUID
	Aggregates PostAggregatesView
	Intent     *fm.AttendanceIntent
}
