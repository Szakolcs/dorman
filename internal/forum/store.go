package forum

import (
	"errors"

	"dorm-man/internal/administration"
	"dorm-man/internal/pagination"

	adm "dorm-man/internal/models/administration"
	fm "dorm-man/internal/models/forum"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store interface {
	LoadPrincipal(userID uuid.UUID) (administration.Principal, error)
	GetTenantByUserID(userID uuid.UUID) (adm.Tenant, error)
	GetTenant(tenantID uuid.UUID) (adm.Tenant, error)
	CreateAudit(event adm.AuditEvent) error
	Transaction(fn func(tx *gorm.DB) error) error

	CreatePost(post fm.ForumPost) (fm.ForumPost, error)
	UpdatePost(post fm.ForumPost) (fm.ForumPost, error)
	GetPost(id uuid.UUID) (fm.ForumPost, error)
	ListPosts(filter FeedListFilter) ([]fm.ForumPost, int64, error)

	UpsertSchedule(schedule fm.ForumPostSchedule) (fm.ForumPostSchedule, error)
	GetScheduleByPostID(postID uuid.UUID) (fm.ForumPostSchedule, error)

	CreatePostUpdate(update fm.ForumPostUpdate) (fm.ForumPostUpdate, error)
	ListPostUpdates(postID uuid.UUID) ([]fm.ForumPostUpdate, error)

	CreatePoll(poll fm.ForumPoll) (fm.ForumPoll, error)
	GetPollByID(id uuid.UUID) (fm.ForumPoll, error)
	GetPollByPostID(postID uuid.UUID) (fm.ForumPoll, error)
	CreatePollOption(opt fm.ForumPollOption) (fm.ForumPollOption, error)
	ListPollOptions(pollID uuid.UUID) ([]fm.ForumPollOption, error)
	CreatePollVote(vote fm.ForumPollVote) (fm.ForumPollVote, error)
	CountPollVotesByOption(pollID uuid.UUID) (map[uuid.UUID]int64, error)
	CountPollVotesByTenant(pollID, tenantID uuid.UUID) (int64, error)
	HasPollVoteForOption(pollID, tenantID, optionID uuid.UUID) (bool, error)

	CreateComment(c fm.ForumComment) (fm.ForumComment, error)
	UpdateComment(c fm.ForumComment) (fm.ForumComment, error)
	GetComment(id uuid.UUID) (fm.ForumComment, error)
	ListComments(filter CommentListFilter) ([]fm.ForumComment, int64, error)
	CountVisibleComments(postID uuid.UUID) (int64, error)

	GetReaction(userID uuid.UUID, targetType fm.ReactionTargetType, targetID uuid.UUID) (fm.ForumReaction, error)
	CreateReaction(r fm.ForumReaction) (fm.ForumReaction, error)
	DeleteReaction(id uuid.UUID) error
	CountReactions(targetType fm.ReactionTargetType, targetID uuid.UUID) (int64, error)

	GetAttendanceIntent(postID, tenantID uuid.UUID) (fm.ForumAttendanceIntent, error)
	UpsertAttendanceIntent(intent fm.ForumAttendanceIntent) (fm.ForumAttendanceIntent, error)
	CountAttendance(postID uuid.UUID, intent fm.AttendanceIntent) (int64, error)

	CreatePostView(view fm.ForumPostView) error
	CreateModerationAction(action fm.ForumModerationAction) (fm.ForumModerationAction, error)
}

type GormStore struct {
	db      *gorm.DB
	adminDB administration.Store
}

func NewStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db, adminDB: administration.NewStore(db)}
}

func (s *GormStore) LoadPrincipal(userID uuid.UUID) (administration.Principal, error) {
	return s.adminDB.LoadPrincipal(userID)
}

func (s *GormStore) GetTenantByUserID(userID uuid.UUID) (adm.Tenant, error) {
	var tenant adm.Tenant
	err := s.db.Where("user_id = ? AND is_active = ?", userID, true).First(&tenant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return adm.Tenant{}, ErrNotFound
	}
	return tenant, err
}

func (s *GormStore) GetTenant(tenantID uuid.UUID) (adm.Tenant, error) {
	return s.adminDB.GetTenant(tenantID)
}

func (s *GormStore) CreateAudit(event adm.AuditEvent) error {
	return s.adminDB.CreateAudit(event)
}

func (s *GormStore) Transaction(fn func(tx *gorm.DB) error) error {
	return s.db.Transaction(fn)
}

func (s *GormStore) CreatePost(post fm.ForumPost) (fm.ForumPost, error) {
	if err := s.db.Create(&post).Error; err != nil {
		return fm.ForumPost{}, err
	}
	return post, nil
}

func (s *GormStore) UpdatePost(post fm.ForumPost) (fm.ForumPost, error) {
	if err := s.db.Save(&post).Error; err != nil {
		return fm.ForumPost{}, err
	}
	return post, nil
}

func (s *GormStore) GetPost(id uuid.UUID) (fm.ForumPost, error) {
	var post fm.ForumPost
	err := s.db.
		Preload("AuthorUser").
		Preload("AuthorTenant").
		Preload("Schedule").
		Preload("Poll.Options").
		Preload("Activity").
		Preload("Event").
		First(&post, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fm.ForumPost{}, ErrNotFound
	}
	return post, err
}

func (s *GormStore) ListPosts(filter FeedListFilter) ([]fm.ForumPost, int64, error) {
	q := s.db.Model(&fm.ForumPost{})

	if !filter.IncludeHidden {
		q = q.Where("state = ?", fm.ForumPostStatePublished)
	} else {
		q = q.Where("state NOT IN ?", []fm.ForumPostState{fm.ForumPostStateDraft, fm.ForumPostStateArchived})
	}

	if filter.Kind != "" {
		q = q.Where("kind = ?", filter.Kind)
	}
	if filter.OfficialOnly {
		q = q.Where("source = ?", fm.ForumPostSourceAdministration)
	}
	if filter.CommunityOnly {
		q = q.Where("source = ? AND kind = ?", fm.ForumPostSourceForum, fm.ForumPostKindCommunityActivity)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	switch filter.Sort {
	case FeedSortPinnedFirst:
		q = q.Order("CASE WHEN pinned_at IS NOT NULL THEN 0 ELSE 1 END").
			Order("pin_priority DESC").
			Order("pinned_at DESC NULLS LAST").
			Order("published_at DESC NULLS LAST").
			Order("id DESC")
	default:
		q = q.Order("published_at DESC NULLS LAST").Order("id DESC")
	}

	var posts []fm.ForumPost
	err := q.Scopes(pagination.Scope(filter.Params)).
		Preload("AuthorUser").
		Preload("AuthorTenant").
		Preload("Schedule").
		Find(&posts).Error
	return posts, total, err
}

func (s *GormStore) UpsertSchedule(schedule fm.ForumPostSchedule) (fm.ForumPostSchedule, error) {
	var existing fm.ForumPostSchedule
	err := s.db.Where("forum_post_id = ?", schedule.ForumPostID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := s.db.Create(&schedule).Error; err != nil {
			return fm.ForumPostSchedule{}, err
		}
		return schedule, nil
	}
	if err != nil {
		return fm.ForumPostSchedule{}, err
	}
	schedule.ID = existing.ID
	if err := s.db.Save(&schedule).Error; err != nil {
		return fm.ForumPostSchedule{}, err
	}
	return schedule, nil
}

func (s *GormStore) GetScheduleByPostID(postID uuid.UUID) (fm.ForumPostSchedule, error) {
	var schedule fm.ForumPostSchedule
	err := s.db.Where("forum_post_id = ?", postID).First(&schedule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fm.ForumPostSchedule{}, ErrNotFound
	}
	return schedule, err
}

func (s *GormStore) CreatePostUpdate(update fm.ForumPostUpdate) (fm.ForumPostUpdate, error) {
	if err := s.db.Create(&update).Error; err != nil {
		return fm.ForumPostUpdate{}, err
	}
	return update, nil
}

func (s *GormStore) ListPostUpdates(postID uuid.UUID) ([]fm.ForumPostUpdate, error) {
	var updates []fm.ForumPostUpdate
	err := s.db.Where("parent_post_id = ?", postID).
		Preload("AuthorUser").
		Order("created_at ASC").
		Find(&updates).Error
	return updates, err
}

func (s *GormStore) CreatePoll(poll fm.ForumPoll) (fm.ForumPoll, error) {
	if err := s.db.Create(&poll).Error; err != nil {
		return fm.ForumPoll{}, err
	}
	return poll, nil
}

func (s *GormStore) GetPollByID(id uuid.UUID) (fm.ForumPoll, error) {
	var poll fm.ForumPoll
	err := s.db.Preload("Options").Preload("ForumPost").First(&poll, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fm.ForumPoll{}, ErrNotFound
	}
	return poll, err
}

func (s *GormStore) GetPollByPostID(postID uuid.UUID) (fm.ForumPoll, error) {
	var poll fm.ForumPoll
	err := s.db.Preload("Options").Where("forum_post_id = ?", postID).First(&poll).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fm.ForumPoll{}, ErrNotFound
	}
	return poll, err
}

func (s *GormStore) CreatePollOption(opt fm.ForumPollOption) (fm.ForumPollOption, error) {
	if err := s.db.Create(&opt).Error; err != nil {
		return fm.ForumPollOption{}, err
	}
	return opt, nil
}

func (s *GormStore) ListPollOptions(pollID uuid.UUID) ([]fm.ForumPollOption, error) {
	var opts []fm.ForumPollOption
	err := s.db.Where("poll_id = ?", pollID).Order("sort_order ASC").Find(&opts).Error
	return opts, err
}

func (s *GormStore) CreatePollVote(vote fm.ForumPollVote) (fm.ForumPollVote, error) {
	if err := s.db.Create(&vote).Error; err != nil {
		return fm.ForumPollVote{}, err
	}
	return vote, nil
}

func (s *GormStore) CountPollVotesByOption(pollID uuid.UUID) (map[uuid.UUID]int64, error) {
	type row struct {
		OptionID uuid.UUID
		Count    int64
	}
	var rows []row
	err := s.db.Model(&fm.ForumPollVote{}).
		Select("option_id, COUNT(*) as count").
		Where("poll_id = ?", pollID).
		Group("option_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]int64, len(rows))
	for _, r := range rows {
		out[r.OptionID] = r.Count
	}
	return out, nil
}

func (s *GormStore) CountPollVotesByTenant(pollID, tenantID uuid.UUID) (int64, error) {
	var n int64
	err := s.db.Model(&fm.ForumPollVote{}).
		Where("poll_id = ? AND tenant_id = ?", pollID, tenantID).
		Count(&n).Error
	return n, err
}

func (s *GormStore) HasPollVoteForOption(pollID, tenantID, optionID uuid.UUID) (bool, error) {
	var n int64
	err := s.db.Model(&fm.ForumPollVote{}).
		Where("poll_id = ? AND tenant_id = ? AND option_id = ?", pollID, tenantID, optionID).
		Count(&n).Error
	return n > 0, err
}

func (s *GormStore) CreateComment(c fm.ForumComment) (fm.ForumComment, error) {
	if err := s.db.Create(&c).Error; err != nil {
		return fm.ForumComment{}, err
	}
	return c, nil
}

func (s *GormStore) UpdateComment(c fm.ForumComment) (fm.ForumComment, error) {
	if err := s.db.Save(&c).Error; err != nil {
		return fm.ForumComment{}, err
	}
	return c, nil
}

func (s *GormStore) GetComment(id uuid.UUID) (fm.ForumComment, error) {
	var c fm.ForumComment
	err := s.db.Preload("AuthorUser").First(&c, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fm.ForumComment{}, ErrNotFound
	}
	return c, err
}

func (s *GormStore) ListComments(filter CommentListFilter) ([]fm.ForumComment, int64, error) {
	q := s.db.Model(&fm.ForumComment{}).
		Where("post_id = ? AND moderation_state = ?", filter.PostID, fm.CommentModerationStateVisible).
		Where("deleted_at IS NULL")

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []fm.ForumComment
	err := q.Order("created_at ASC").
		Preload("AuthorUser").
		Scopes(pagination.Scope(filter.Params)).
		Find(&list).Error
	return list, total, err
}

func (s *GormStore) CountVisibleComments(postID uuid.UUID) (int64, error) {
	var n int64
	err := s.db.Model(&fm.ForumComment{}).
		Where("post_id = ? AND moderation_state = ?", postID, fm.CommentModerationStateVisible).
		Where("deleted_at IS NULL").
		Count(&n).Error
	return n, err
}

func (s *GormStore) GetReaction(userID uuid.UUID, targetType fm.ReactionTargetType, targetID uuid.UUID) (fm.ForumReaction, error) {
	var r fm.ForumReaction
	err := s.db.Where(
		"user_id = ? AND target_type = ? AND target_id = ?",
		userID, targetType, targetID,
	).First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fm.ForumReaction{}, ErrNotFound
	}
	return r, err
}

func (s *GormStore) CreateReaction(r fm.ForumReaction) (fm.ForumReaction, error) {
	if err := s.db.Create(&r).Error; err != nil {
		return fm.ForumReaction{}, err
	}
	return r, nil
}

func (s *GormStore) DeleteReaction(id uuid.UUID) error {
	return s.db.Delete(&fm.ForumReaction{}, "id = ?", id).Error
}

func (s *GormStore) CountReactions(targetType fm.ReactionTargetType, targetID uuid.UUID) (int64, error) {
	var n int64
	err := s.db.Model(&fm.ForumReaction{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Count(&n).Error
	return n, err
}

func (s *GormStore) GetAttendanceIntent(postID, tenantID uuid.UUID) (fm.ForumAttendanceIntent, error) {
	var intent fm.ForumAttendanceIntent
	err := s.db.Where("post_id = ? AND tenant_id = ?", postID, tenantID).First(&intent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fm.ForumAttendanceIntent{}, ErrNotFound
	}
	return intent, err
}

func (s *GormStore) UpsertAttendanceIntent(intent fm.ForumAttendanceIntent) (fm.ForumAttendanceIntent, error) {
	var existing fm.ForumAttendanceIntent
	err := s.db.Where("post_id = ? AND tenant_id = ?", intent.PostID, intent.TenantID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := s.db.Create(&intent).Error; err != nil {
			return fm.ForumAttendanceIntent{}, err
		}
		return intent, nil
	}
	if err != nil {
		return fm.ForumAttendanceIntent{}, err
	}
	existing.Intent = intent.Intent
	if err := s.db.Save(&existing).Error; err != nil {
		return fm.ForumAttendanceIntent{}, err
	}
	return existing, nil
}

func (s *GormStore) CountAttendance(postID uuid.UUID, intent fm.AttendanceIntent) (int64, error) {
	var n int64
	err := s.db.Model(&fm.ForumAttendanceIntent{}).
		Where("post_id = ? AND intent = ?", postID, intent).
		Count(&n).Error
	return n, err
}

func (s *GormStore) CreatePostView(view fm.ForumPostView) error {
	return s.db.Create(&view).Error
}

func (s *GormStore) CreateModerationAction(action fm.ForumModerationAction) (fm.ForumModerationAction, error) {
	if err := s.db.Create(&action).Error; err != nil {
		return fm.ForumModerationAction{}, err
	}
	return action, nil
}
