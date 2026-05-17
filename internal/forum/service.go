package forum

import (
	"errors"
	"strings"
	"time"

	"dorm-man/internal/administration"

	adm "dorm-man/internal/models/administration"
	fm "dorm-man/internal/models/forum"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ResolvePrincipal(actorID uuid.UUID) (administration.Principal, error) {
	return s.store.LoadPrincipal(actorID)
}

func (s *Service) ListFeed(principal administration.Principal, filter FeedListFilter) ([]fm.ForumPost, error) {
	if err := s.requireAuthenticated(principal); err != nil {
		return nil, err
	}
	if isStaffModerator(principal) && filter.IncludeHidden {
		// staff may request hidden items via filter flag
	} else {
		filter.IncludeHidden = false
	}
	if filter.Sort == "" {
		filter.Sort = FeedSortNewest
	}
	return s.store.ListPosts(filter)
}

func (s *Service) GetPost(principal administration.Principal, postID uuid.UUID, recordView bool) (PostDetail, error) {
	if err := s.requireAuthenticated(principal); err != nil {
		return PostDetail{}, err
	}
	post, err := s.store.GetPost(postID)
	if err != nil {
		return PostDetail{}, err
	}
	if err := s.ensurePostReadable(principal, post); err != nil {
		return PostDetail{}, err
	}
	agg, err := s.buildPostAggregates(principal, post)
	if err != nil {
		return PostDetail{}, err
	}
	if recordView {
		s.recordPostViewBestEffort(principal, postID)
	}
	return PostDetail{Post: post, Aggregates: agg}, nil
}

func (s *Service) CreateCommunityActivity(principal administration.Principal, in CommunityPostInput) (fm.ForumPost, error) {
	tenant, err := s.requireActiveTenant(principal)
	if err != nil {
		return fm.ForumPost{}, err
	}
	title := strings.TrimSpace(in.Title)
	body := strings.TrimSpace(in.Body)
	if title == "" || body == "" {
		return fm.ForumPost{}, ErrValidation
	}

	state := fm.ForumPostStateDraft
	var publishedAt *time.Time
	if in.Publish {
		state = fm.ForumPostStatePublished
		now := time.Now().UTC()
		publishedAt = &now
	}

	post := fm.ForumPost{
		AuthorUserID:   principal.UserID,
		AuthorTenantID: &tenant.ID,
		Kind:           fm.ForumPostKindCommunityActivity,
		Title:          title,
		Body:           body,
		State:          state,
		Source:         fm.ForumPostSourceForum,
		PublishedAt:    publishedAt,
		Tags:           strings.TrimSpace(in.Tags),
	}

	created, err := s.store.CreatePost(post)
	if err != nil {
		return fm.ForumPost{}, err
	}

	if in.StartsAt != nil || in.EndsAt != nil || strings.TrimSpace(in.Location) != "" {
		_, err = s.store.UpsertSchedule(fm.ForumPostSchedule{
			ForumPostID: created.ID,
			StartsAt:    in.StartsAt,
			EndsAt:      in.EndsAt,
			Location:    strings.TrimSpace(in.Location),
		})
		if err != nil {
			return fm.ForumPost{}, err
		}
	}

	return s.store.GetPost(created.ID)
}

func (s *Service) UpdateCommunityActivity(principal administration.Principal, postID uuid.UUID, in CommunityPostUpdateInput) (fm.ForumPost, error) {
	tenant, err := s.requireActiveTenant(principal)
	if err != nil {
		return fm.ForumPost{}, err
	}
	post, err := s.store.GetPost(postID)
	if err != nil {
		return fm.ForumPost{}, err
	}
	if err := s.ensureCommunityAuthor(post, tenant.ID, principal.UserID); err != nil {
		return fm.ForumPost{}, err
	}
	if post.State != fm.ForumPostStatePublished && post.State != fm.ForumPostStateDraft {
		return fm.ForumPost{}, ErrEditNotAllowed
	}

	title := strings.TrimSpace(in.Title)
	body := strings.TrimSpace(in.Body)
	if title != "" {
		post.Title = title
	}
	if body != "" {
		post.Body = body
	}
	if in.Tags != "" {
		post.Tags = strings.TrimSpace(in.Tags)
	}

	updated, err := s.store.UpdatePost(post)
	if err != nil {
		return fm.ForumPost{}, err
	}

	if in.StartsAt != nil || in.EndsAt != nil || in.Location != "" {
		schedule := fm.ForumPostSchedule{
			ForumPostID: postID,
			StartsAt:    in.StartsAt,
			EndsAt:      in.EndsAt,
			Location:    strings.TrimSpace(in.Location),
		}
		if _, err := s.store.UpsertSchedule(schedule); err != nil {
			return fm.ForumPost{}, err
		}
	}

	return s.store.GetPost(updated.ID)
}

func (s *Service) ArchiveCommunityActivity(principal administration.Principal, postID uuid.UUID) (fm.ForumPost, error) {
	tenant, err := s.requireActiveTenant(principal)
	if err != nil {
		return fm.ForumPost{}, err
	}
	post, err := s.store.GetPost(postID)
	if err != nil {
		return fm.ForumPost{}, err
	}
	if err := s.ensureCommunityAuthor(post, tenant.ID, principal.UserID); err != nil {
		return fm.ForumPost{}, err
	}
	post.State = fm.ForumPostStateArchived
	updated, err := s.store.UpdatePost(post)
	if err != nil {
		return fm.ForumPost{}, err
	}
	return updated, nil
}

func (s *Service) CreateComment(principal administration.Principal, postID uuid.UUID, in CommentInput) (fm.ForumComment, error) {
	if err := s.requireAuthenticated(principal); err != nil {
		return fm.ForumComment{}, err
	}
	post, err := s.store.GetPost(postID)
	if err != nil {
		return fm.ForumComment{}, err
	}
	if err := s.ensurePostReadable(principal, post); err != nil {
		return fm.ForumComment{}, err
	}
	body := strings.TrimSpace(in.Body)
	if body == "" {
		return fm.ForumComment{}, ErrValidation
	}
	if in.ParentCommentID != nil {
		parent, err := s.store.GetComment(*in.ParentCommentID)
		if err != nil {
			return fm.ForumComment{}, err
		}
		if parent.PostID != postID {
			return fm.ForumComment{}, ErrValidation
		}
	}

	comment := fm.ForumComment{
		PostID:          postID,
		AuthorUserID:    principal.UserID,
		ParentCommentID: in.ParentCommentID,
		Body:            body,
		ModerationState: fm.CommentModerationStateVisible,
	}
	return s.store.CreateComment(comment)
}

func (s *Service) UpdateComment(principal administration.Principal, commentID uuid.UUID, in CommentUpdateInput) (fm.ForumComment, error) {
	if err := s.requireAuthenticated(principal); err != nil {
		return fm.ForumComment{}, err
	}
	comment, err := s.store.GetComment(commentID)
	if err != nil {
		return fm.ForumComment{}, err
	}
	if comment.AuthorUserID != principal.UserID {
		return fm.ForumComment{}, ErrEditNotAllowed
	}
	body := strings.TrimSpace(in.Body)
	if body == "" {
		return fm.ForumComment{}, ErrValidation
	}
	now := time.Now().UTC()
	comment.Body = body
	comment.EditedAt = &now
	return s.store.UpdateComment(comment)
}

func (s *Service) DeleteComment(principal administration.Principal, commentID uuid.UUID) error {
	if err := s.requireAuthenticated(principal); err != nil {
		return err
	}
	comment, err := s.store.GetComment(commentID)
	if err != nil {
		return err
	}
	if comment.AuthorUserID != principal.UserID && !isStaffModerator(principal) {
		return ErrEditNotAllowed
	}
	comment.ModerationState = fm.CommentModerationStateRemoved
	_, err = s.store.UpdateComment(comment)
	return err
}

func (s *Service) ListComments(principal administration.Principal, filter CommentListFilter) ([]fm.ForumComment, error) {
	if err := s.requireAuthenticated(principal); err != nil {
		return nil, err
	}
	post, err := s.store.GetPost(filter.PostID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePostReadable(principal, post); err != nil {
		return nil, err
	}
	return s.store.ListComments(filter)
}

func (s *Service) ToggleReaction(principal administration.Principal, targetType fm.ReactionTargetType, targetID uuid.UUID) (bool, int64, error) {
	if err := s.requireAuthenticated(principal); err != nil {
		return false, 0, err
	}
	if err := s.ensureReactionTargetReadable(principal, targetType, targetID); err != nil {
		return false, 0, err
	}

	existing, err := s.store.GetReaction(principal.UserID, targetType, targetID)
	if err == nil {
		if err := s.store.DeleteReaction(existing.ID); err != nil {
			return false, 0, err
		}
		count, err := s.store.CountReactions(targetType, targetID)
		return false, count, err
	}
	if !errors.Is(err, ErrNotFound) {
		return false, 0, err
	}

	_, err = s.store.CreateReaction(fm.ForumReaction{
		UserID:       principal.UserID,
		TargetType:   targetType,
		TargetID:     targetID,
		ReactionType: fm.ReactionTypeLike,
	})
	if err != nil {
		return false, 0, err
	}
	count, err := s.store.CountReactions(targetType, targetID)
	return true, count, err
}

func (s *Service) SetAttendanceIntent(principal administration.Principal, postID uuid.UUID, in AttendanceInput) (fm.ForumAttendanceIntent, PostAggregates, error) {
	tenant, err := s.requireActiveTenant(principal)
	if err != nil {
		return fm.ForumAttendanceIntent{}, PostAggregates{}, err
	}
	post, err := s.store.GetPost(postID)
	if err != nil {
		return fm.ForumAttendanceIntent{}, PostAggregates{}, err
	}
	if err := s.ensurePostReadable(principal, post); err != nil {
		return fm.ForumAttendanceIntent{}, PostAggregates{}, err
	}
	if !supportsAttendance(post.Kind) {
		return fm.ForumAttendanceIntent{}, PostAggregates{}, ErrValidation
	}
	switch in.Intent {
	case fm.AttendanceIntentGoing, fm.AttendanceIntentNotGoing:
	default:
		return fm.ForumAttendanceIntent{}, PostAggregates{}, ErrValidation
	}

	intent, err := s.store.UpsertAttendanceIntent(fm.ForumAttendanceIntent{
		PostID:   postID,
		TenantID: tenant.ID,
		Intent:   in.Intent,
	})
	if err != nil {
		return fm.ForumAttendanceIntent{}, PostAggregates{}, err
	}
	agg, err := s.buildPostAggregates(principal, post)
	return intent, agg, err
}

func (s *Service) CreatePoll(principal administration.Principal, in PollCreateInput) (fm.ForumPost, error) {
	if !canManageOfficialPoll(principal) {
		return fm.ForumPost{}, ErrUnauthorized
	}
	title := strings.TrimSpace(in.Title)
	body := strings.TrimSpace(in.Body)
	if title == "" || body == "" {
		return fm.ForumPost{}, ErrValidation
	}
	if len(in.Options) < 2 {
		return fm.ForumPost{}, ErrValidation
	}
	if in.ClosesAt.IsZero() || !in.ClosesAt.After(time.Now().UTC()) {
		return fm.ForumPost{}, ErrValidation
	}
	choiceMode := in.ChoiceMode
	if choiceMode == "" {
		choiceMode = fm.ForumPollChoiceModeSingle
	}
	visibility := in.ResultsVisibility
	if visibility == "" {
		visibility = fm.ForumPollResultsVisibilityLive
	}

	now := time.Now().UTC()
	post := fm.ForumPost{
		AuthorUserID: principal.UserID,
		Kind:         fm.ForumPostKindPoll,
		Title:        title,
		Body:         body,
		State:        fm.ForumPostStatePublished,
		Source:       fm.ForumPostSourceForum,
		PublishedAt:  &now,
	}

	var createdPost fm.ForumPost
	err := s.store.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&post).Error; err != nil {
			return err
		}
		createdPost = post

		poll := fm.ForumPoll{
			ForumPostID:       post.ID,
			ChoiceMode:        choiceMode,
			ResultsVisibility: visibility,
			ClosesAt:          in.ClosesAt.UTC(),
			CreatedByUserID:   principal.UserID,
		}
		if err := tx.Create(&poll).Error; err != nil {
			return err
		}
		for i, optIn := range in.Options {
			label := strings.TrimSpace(optIn.Label)
			if label == "" {
				return ErrValidation
			}
			order := optIn.SortOrder
			if order == 0 {
				order = i
			}
			opt := fm.ForumPollOption{
				PollID:    poll.ID,
				Label:     label,
				SortOrder: order,
			}
			if err := tx.Create(&opt).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fm.ForumPost{}, err
	}
	_ = s.store.CreateAudit(newAudit(principal.UserID, "forum.poll.create", "forum_post", createdPost.ID.String(), adm.AuditOutcomeSuccess))
	return s.store.GetPost(createdPost.ID)
}

func (s *Service) CastPollVote(principal administration.Principal, pollID uuid.UUID, in PollVoteInput) error {
	tenant, err := s.requireActiveTenant(principal)
	if err != nil {
		return err
	}
	poll, err := s.store.GetPollByID(pollID)
	if err != nil {
		return err
	}
	if poll.ForumPost.State != fm.ForumPostStatePublished {
		return ErrNotFound
	}
	if time.Now().UTC().After(poll.ClosesAt) {
		return ErrPollClosed
	}
	if len(in.OptionIDs) == 0 {
		return ErrValidation
	}

	switch poll.ChoiceMode {
	case fm.ForumPollChoiceModeSingle:
		if len(in.OptionIDs) != 1 {
			return ErrValidation
		}
		n, err := s.store.CountPollVotesByTenant(poll.ID, tenant.ID)
		if err != nil {
			return err
		}
		if n > 0 {
			return ErrAlreadyVoted
		}
	default:
		for _, optionID := range in.OptionIDs {
			ok, err := s.store.HasPollVoteForOption(poll.ID, tenant.ID, optionID)
			if err != nil {
				return err
			}
			if ok {
				return ErrAlreadyVoted
			}
		}
	}

	opts, err := s.store.ListPollOptions(poll.ID)
	if err != nil {
		return err
	}
	valid := make(map[uuid.UUID]struct{}, len(opts))
	for _, o := range opts {
		valid[o.ID] = struct{}{}
	}
	now := time.Now().UTC()
	for _, optionID := range in.OptionIDs {
		if _, ok := valid[optionID]; !ok {
			return ErrValidation
		}
		_, err := s.store.CreatePollVote(fm.ForumPollVote{
			PollID:   poll.ID,
			OptionID: optionID,
			TenantID: tenant.ID,
			VotedAt:  now,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetPollResults(principal administration.Principal, pollID uuid.UUID) (PollResults, error) {
	if err := s.requireAuthenticated(principal); err != nil {
		return PollResults{}, err
	}
	poll, err := s.store.GetPollByID(pollID)
	if err != nil {
		return PollResults{}, err
	}
	if err := s.ensurePostReadable(principal, poll.ForumPost); err != nil {
		return PollResults{}, err
	}

	visible := poll.ResultsVisibility == fm.ForumPollResultsVisibilityLive ||
		time.Now().UTC().After(poll.ClosesAt) ||
		canManageOfficialPoll(principal)

	if !visible {
		return PollResults{}, ErrResultsNotVisible
	}

	counts, err := s.store.CountPollVotesByOption(poll.ID)
	if err != nil {
		return PollResults{}, err
	}
	opts, err := s.store.ListPollOptions(poll.ID)
	if err != nil {
		return PollResults{}, err
	}
	result := PollResults{Poll: poll, Visible: true, Options: make([]PollOptionCount, 0, len(opts))}
	for _, o := range opts {
		result.Options = append(result.Options, PollOptionCount{
			OptionID: o.ID,
			Label:    o.Label,
			Votes:    counts[o.ID],
		})
	}
	return result, nil
}

func (s *Service) AddOrganizerUpdate(principal administration.Principal, postID uuid.UUID, in PostOrganizerUpdateInput) (fm.ForumPostUpdate, error) {
	if !canManageOfficialPoll(principal) {
		return fm.ForumPostUpdate{}, ErrUnauthorized
	}
	body := strings.TrimSpace(in.Body)
	if body == "" {
		return fm.ForumPostUpdate{}, ErrValidation
	}
	post, err := s.store.GetPost(postID)
	if err != nil {
		return fm.ForumPostUpdate{}, err
	}
	if post.Kind != fm.ForumPostKindEvent && post.Kind != fm.ForumPostKindOfficialNews {
		return fm.ForumPostUpdate{}, ErrValidation
	}
	return s.store.CreatePostUpdate(fm.ForumPostUpdate{
		ParentPostID: postID,
		AuthorUserID: principal.UserID,
		Body:         body,
	})
}

func (s *Service) ListOrganizerUpdates(principal administration.Principal, postID uuid.UUID) ([]fm.ForumPostUpdate, error) {
	if err := s.requireAuthenticated(principal); err != nil {
		return nil, err
	}
	post, err := s.store.GetPost(postID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePostReadable(principal, post); err != nil {
		return nil, err
	}
	return s.store.ListPostUpdates(postID)
}

func (s *Service) Moderate(principal administration.Principal, in ModerationInput) error {
	if !isStaffModerator(principal) {
		return ErrUnauthorized
	}
	now := time.Now().UTC()
	action := fm.ForumModerationAction{
		ModeratorUserID: principal.UserID,
		TargetType:      in.TargetType,
		TargetID:        in.TargetID,
		Action:          in.Action,
		Reason:          strings.TrimSpace(in.Reason),
		OccurredAt:      now,
	}

	var auditAction string
	switch in.TargetType {
	case fm.ForumModerationTargetPost:
		post, err := s.store.GetPost(in.TargetID)
		if err != nil {
			return err
		}
		switch in.Action {
		case fm.ForumModerationActionHide:
			post.State = fm.ForumPostStateHidden
			auditAction = "forum.moderation.hide"
		case fm.ForumModerationActionRestore:
			if post.PublishedAt != nil {
				post.State = fm.ForumPostStatePublished
			} else {
				post.State = fm.ForumPostStateDraft
			}
			auditAction = "forum.moderation.restore"
		case fm.ForumModerationActionRemove:
			post.State = fm.ForumPostStateArchived
			auditAction = "forum.moderation.remove"
		default:
			return ErrValidation
		}
		if _, err := s.store.UpdatePost(post); err != nil {
			return err
		}
	case fm.ForumModerationTargetComment:
		comment, err := s.store.GetComment(in.TargetID)
		if err != nil {
			return err
		}
		switch in.Action {
		case fm.ForumModerationActionHide:
			comment.ModerationState = fm.CommentModerationStateHidden
			auditAction = "forum.moderation.hide"
		case fm.ForumModerationActionRestore:
			comment.ModerationState = fm.CommentModerationStateVisible
			auditAction = "forum.moderation.restore"
		case fm.ForumModerationActionRemove:
			comment.ModerationState = fm.CommentModerationStateRemoved
			auditAction = "forum.moderation.remove"
		default:
			return ErrValidation
		}
		if _, err := s.store.UpdateComment(comment); err != nil {
			return err
		}
	default:
		return ErrValidation
	}

	if _, err := s.store.CreateModerationAction(action); err != nil {
		return err
	}
	_ = s.store.CreateAudit(newAudit(principal.UserID, auditAction, string(in.TargetType), in.TargetID.String(), adm.AuditOutcomeSuccess))
	return nil
}

func (s *Service) buildPostAggregates(principal administration.Principal, post fm.ForumPost) (PostAggregates, error) {
	reactions, err := s.store.CountReactions(fm.ReactionTargetPost, post.ID)
	if err != nil {
		return PostAggregates{}, err
	}
	comments, err := s.store.CountVisibleComments(post.ID)
	if err != nil {
		return PostAggregates{}, err
	}
	going, err := s.store.CountAttendance(post.ID, fm.AttendanceIntentGoing)
	if err != nil {
		return PostAggregates{}, err
	}
	notGoing, err := s.store.CountAttendance(post.ID, fm.AttendanceIntentNotGoing)
	if err != nil {
		return PostAggregates{}, err
	}

	agg := PostAggregates{
		ReactionCount: reactions,
		CommentCount:  comments,
		GoingCount:    going,
		NotGoingCount: notGoing,
	}

	if r, err := s.store.GetReaction(principal.UserID, fm.ReactionTargetPost, post.ID); err == nil {
		t := r.ReactionType
		agg.ViewerReaction = &t
	} else if !errors.Is(err, ErrNotFound) {
		return PostAggregates{}, err
	}

	if hasTenantRole(principal) {
		tenant, err := s.store.GetTenantByUserID(principal.UserID)
		if err == nil {
			if intent, err := s.store.GetAttendanceIntent(post.ID, tenant.ID); err == nil {
				i := intent.Intent
				agg.ViewerAttendanceIntent = &i
			} else if !errors.Is(err, ErrNotFound) {
				return PostAggregates{}, err
			}
		}
	}

	return agg, nil
}

func (s *Service) recordPostViewBestEffort(principal administration.Principal, postID uuid.UUID) {
	view := fm.ForumPostView{
		PostID:   postID,
		ViewedAt: time.Now().UTC(),
	}
	view.ViewerUserID = &principal.UserID
	if hasTenantRole(principal) {
		if tenant, err := s.store.GetTenantByUserID(principal.UserID); err == nil {
			view.ViewerTenantID = &tenant.ID
		}
	}
	_ = s.store.CreatePostView(view)
}

func (s *Service) ensurePostReadable(principal administration.Principal, post fm.ForumPost) error {
	if isStaffModerator(principal) {
		return nil
	}
	if post.State == fm.ForumPostStatePublished {
		return nil
	}
	return ErrNotFound
}

func (s *Service) ensureReactionTargetReadable(principal administration.Principal, targetType fm.ReactionTargetType, targetID uuid.UUID) error {
	switch targetType {
	case fm.ReactionTargetPost:
		post, err := s.store.GetPost(targetID)
		if err != nil {
			return err
		}
		return s.ensurePostReadable(principal, post)
	case fm.ReactionTargetComment:
		comment, err := s.store.GetComment(targetID)
		if err != nil {
			return err
		}
		if comment.ModerationState != fm.CommentModerationStateVisible {
			return ErrNotFound
		}
		post, err := s.store.GetPost(comment.PostID)
		if err != nil {
			return err
		}
		return s.ensurePostReadable(principal, post)
	default:
		return ErrValidation
	}
}

func (s *Service) ensureCommunityAuthor(post fm.ForumPost, tenantID uuid.UUID, userID uuid.UUID) error {
	if post.Source == fm.ForumPostSourceAdministration {
		return ErrEditNotAllowed
	}
	if post.Kind != fm.ForumPostKindCommunityActivity {
		return ErrEditNotAllowed
	}
	if post.AuthorUserID != userID {
		return ErrEditNotAllowed
	}
	if post.AuthorTenantID == nil || *post.AuthorTenantID != tenantID {
		return ErrEditNotAllowed
	}
	return nil
}

func (s *Service) requireAuthenticated(principal administration.Principal) error {
	if principal.UserID == uuid.Nil {
		return ErrUnauthorized
	}
	return nil
}

func (s *Service) requireActiveTenant(principal administration.Principal) (adm.Tenant, error) {
	if !hasTenantRole(principal) {
		return adm.Tenant{}, ErrUnauthorized
	}
	tenant, err := s.store.GetTenantByUserID(principal.UserID)
	if err != nil {
		return adm.Tenant{}, ErrIneligibleVoter
	}
	return tenant, nil
}

func supportsAttendance(kind fm.ForumPostKind) bool {
	switch kind {
	case fm.ForumPostKindCommunityActivity, fm.ForumPostKindEvent:
		return true
	default:
		return false
	}
}

func hasTenantRole(p administration.Principal) bool {
	for _, role := range p.Roles {
		if role == adm.RoleTenant {
			return true
		}
	}
	return false
}

func isStaffModerator(p administration.Principal) bool {
	for _, role := range p.Roles {
		switch role {
		case adm.RoleAdministrator, adm.RoleOfficeWorker, adm.RoleDirector:
			return true
		}
	}
	return false
}

func canManageOfficialPoll(p administration.Principal) bool {
	return isStaffModerator(p)
}

func newAudit(actorID uuid.UUID, action, targetType, targetID string, outcome adm.AuditOutcome) adm.AuditEvent {
	return adm.AuditEvent{
		ActorUserID: &actorID,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		Outcome:     outcome,
		OccurredAt:  time.Now().UTC(),
	}
}
