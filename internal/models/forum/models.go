package models

// All returns forum-owned model pointers in migration order (FK dependencies first).
func All() []any {
	return []any{
		&ForumPost{},
		&ForumPostSchedule{},
		&ForumPostUpdate{},
		&ForumPoll{},
		&ForumPollOption{},
		&ForumPollVote{},
		&ForumComment{},
		&ForumReaction{},
		&ForumAttendanceIntent{},
		&ForumPostView{},
		&ForumModerationAction{},
	}
}
