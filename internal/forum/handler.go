package forum

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dorm-man/internal/administration"
	"dorm-man/internal/pagination"
	"dorm-man/internal/platform"
	fm "dorm-man/internal/models/forum"
	forumviews "dorm-man/web/templates/forum"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) actor(c echo.Context) (administration.Principal, error) {
	return h.actorFromRequest(c)
}

func (h *Handler) actorFromRequest(c echo.Context) (administration.Principal, error) {
	id, ok := platform.ActorUserID(c)
	if !ok {
		return administration.Principal{}, ErrUnauthorized
	}
	return h.service.ResolvePrincipal(id)
}

func (h *Handler) writeError(c echo.Context, err error) error {
	category := classifyForumError(err)
	status := http.StatusInternalServerError
	switch category {
	case "validation_error":
		status = http.StatusBadRequest
	case "authorization_denied":
		status = http.StatusForbidden
	case "not_found", "results_not_visible":
		status = http.StatusNotFound
	case "poll_closed", "already_voted", "edit_not_allowed", "ineligible_voter", "concurrency_conflict":
		status = http.StatusConflict
	}
	return c.JSON(status, map[string]any{
		"error": map[string]string{
			"category": category,
			"message":  err.Error(),
		},
	})
}

func classifyForumError(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrValidation):
		return "validation_error"
	case errors.Is(err, ErrUnauthorized):
		return "authorization_denied"
	case errors.Is(err, ErrNotFound):
		return "not_found"
	case errors.Is(err, ErrPollClosed):
		return "poll_closed"
	case errors.Is(err, ErrAlreadyVoted):
		return "already_voted"
	case errors.Is(err, ErrIneligibleVoter):
		return "ineligible_voter"
	case errors.Is(err, ErrResultsNotVisible):
		return "results_not_visible"
	case errors.Is(err, ErrEditNotAllowed):
		return "edit_not_allowed"
	case errors.Is(err, ErrConcurrencyConflict):
		return "concurrency_conflict"
	default:
		return "internal_error"
	}
}

func queryPositiveInt(c echo.Context, name string, def int) int {
	s := c.QueryParam(name)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return def
	}
	return n
}

func parseBoolQuery(c echo.Context, name string) bool {
	return c.QueryParam(name) == "true" || c.QueryParam(name) == "1"
}

func (h *Handler) listFeed(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	filter := feedFilterFromRequest(c)
	posts, total, err := h.service.ListFeed(principal, filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return writeListJSON(c, posts, pagination.NewMeta(filter.Params, total))
}

func (h *Handler) getPost(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	detail, err := h.service.GetPost(principal, id, parseBoolQuery(c, "record_view"))
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": detail})
}

func (h *Handler) createCommunityPost(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body CommunityPostInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	post, err := h.service.CreateCommunityActivity(principal, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": post})
}

func (h *Handler) updateCommunityPost(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body CommunityPostUpdateInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	post, err := h.service.UpdateCommunityActivity(principal, id, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": post})
}

func (h *Handler) archiveCommunityPost(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	post, err := h.service.ArchiveCommunityActivity(principal, id)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": post})
}

func (h *Handler) listComments(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	params := pageParams(c)
	filter := CommentListFilter{PostID: postID, Params: params}
	list, total, err := h.service.ListComments(principal, filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return writeListJSON(c, list, pagination.NewMeta(params, total))
}

func (h *Handler) createComment(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body CommentInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	comment, err := h.service.CreateComment(principal, postID, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": comment})
}

func (h *Handler) updateComment(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body CommentUpdateInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	comment, err := h.service.UpdateComment(principal, commentID, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": comment})
}

func (h *Handler) deleteComment(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	if err := h.service.DeleteComment(principal, commentID); err != nil {
		return h.writeError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) togglePostReaction(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	added, count, err := h.service.ToggleReaction(principal, fm.ReactionTargetPost, postID)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{"added": added, "reaction_count": count},
	})
}

func (h *Handler) toggleReaction(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body ReactionTargetInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	added, count, err := h.service.ToggleReaction(principal, body.TargetType, body.TargetID)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{"added": added, "reaction_count": count},
	})
}

func (h *Handler) setAttendance(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body AttendanceInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	intent, agg, err := h.service.SetAttendanceIntent(principal, postID, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"intent":     intent,
			"aggregates": agg,
		},
	})
}

func (h *Handler) createPoll(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body PollCreateInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	post, err := h.service.CreatePoll(principal, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": post})
}

func (h *Handler) castPollVote(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	pollID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body PollVoteInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	if err := h.service.CastPollVote(principal, pollID, body); err != nil {
		return h.writeError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) getPollResults(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	pollID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	results, err := h.service.GetPollResults(principal, pollID)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": results})
}

func (h *Handler) addOrganizerUpdate(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body PostOrganizerUpdateInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	update, err := h.service.AddOrganizerUpdate(principal, postID, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": update})
}

func (h *Handler) listOrganizerUpdates(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	updates, err := h.service.ListOrganizerUpdates(principal, postID)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": updates})
}

func (h *Handler) moderate(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body ModerationInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	if err := h.service.Moderate(principal, body); err != nil {
		return h.writeError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func isHtmx(c echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}

func actorUserIDFromRequest(c echo.Context) string {
	return platform.ActorUserIDString(c)
}

func (h *Handler) renderViewError(c echo.Context, err error) error {
	msg := err.Error()
	if isHtmx(c) {
		return renderComponent(c, forumviews.ErrorAlert(msg))
	}
	return renderComponent(c, forumviews.ForumDocument("Forum error", actorUserIDFromRequest(c), forumviews.ErrorAlert(msg)))
}

func (h *Handler) feedPage(c echo.Context) error {
	actorID := actorUserIDFromRequest(c)
	principal, err := h.actorFromRequest(c)
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			return renderComponent(c, forumviews.FeedPage(forumviews.FeedPageData{ActorUserID: actorID}))
		}
		return h.renderViewError(c, err)
	}
	data, err := h.buildFeedPageData(c, principal)
	if err != nil {
		return h.renderViewError(c, err)
	}
	return renderComponent(c, forumviews.FeedPage(data))
}

func (h *Handler) feedListFragment(c echo.Context) error {
	principal, err := h.actorFromRequest(c)
	if err != nil {
		return h.renderViewError(c, err)
	}
	data, err := h.buildFeedPageData(c, principal)
	if err != nil {
		return h.renderViewError(c, err)
	}
	return renderComponent(c, forumviews.FeedListFragment(data))
}

func (h *Handler) buildFeedPageData(c echo.Context, principal administration.Principal) (forumviews.FeedPageData, error) {
	filter := feedFilterFromRequest(c)
	posts, total, err := h.service.ListFeed(principal, filter)
	if err != nil {
		return forumviews.FeedPageData{}, err
	}
	return forumviews.FeedPageData{
		Posts:         posts,
		Sort:          string(filter.Sort),
		Kind:          string(filter.Kind),
		OfficialOnly:  filter.OfficialOnly,
		CommunityOnly: filter.CommunityOnly,
		ActorUserID:   actorUserIDFromRequest(c),
		Pagination:    pagination.NewMeta(filter.Params, total),
		Preserve:      queryPreserve(c),
	}, nil
}

func feedFilterFromRequest(c echo.Context) FeedListFilter {
	sort := FeedSort(c.QueryParam("sort"))
	if sort == "" {
		sort = FeedSortNewest
	}
	return FeedListFilter{
		Kind:          fm.ForumPostKind(c.QueryParam("kind")),
		OfficialOnly:  parseBoolQuery(c, "official_only"),
		CommunityOnly: parseBoolQuery(c, "community_only"),
		Sort:          sort,
		IncludeHidden: parseBoolQuery(c, "include_hidden"),
		Params:        pageParams(c),
	}
}

func (h *Handler) postPage(c echo.Context) error {
	actorID := actorUserIDFromRequest(c)
	principal, err := h.actorFromRequest(c)
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			return renderComponent(c, forumviews.ForumDocument("Forum", actorID, forumviews.ForumEmptyBody()))
		}
		return h.renderViewError(c, err)
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.renderViewError(c, ErrValidation)
	}
	data, err := h.buildPostPageData(c, principal, postID, true, actorUserIDFromRequest(c))
	if err != nil {
		return h.renderViewError(c, err)
	}
	return renderComponent(c, forumviews.PostPage(data))
}

func (h *Handler) buildPostPageData(c echo.Context, principal administration.Principal, postID uuid.UUID, recordView bool, actorUserID string) (forumviews.PostPageData, error) {
	detail, err := h.service.GetPost(principal, postID, recordView)
	if err != nil {
		return forumviews.PostPageData{}, err
	}
	commentParams := pageParamsNamed(c, "comments_page", "comments_page_size")
	comments, commentTotal, err := h.service.ListComments(principal, CommentListFilter{PostID: postID, Params: commentParams})
	if err != nil {
		return forumviews.PostPageData{}, err
	}
	updates, err := h.service.ListOrganizerUpdates(principal, postID)
	if err != nil {
		return forumviews.PostPageData{}, err
	}

	agg := forumviews.PostAggregatesView{
		ReactionCount: detail.Aggregates.ReactionCount,
		CommentCount:  detail.Aggregates.CommentCount,
		GoingCount:    detail.Aggregates.GoingCount,
		NotGoingCount: detail.Aggregates.NotGoingCount,
		ViewerLiked:   detail.Aggregates.ViewerReaction != nil,
	}
	if detail.Aggregates.ViewerAttendanceIntent != nil {
		agg.ViewerAttendanceIntent = detail.Aggregates.ViewerAttendanceIntent
	}

	data := forumviews.PostPageData{
		Post:               detail.Post,
		Aggregates:         agg,
		Schedule:           detail.Post.Schedule,
		Comments:           comments,
		CommentsPagination: pagination.NewMeta(commentParams, commentTotal),
		CommentsPreserve:   queryPreserve(c, "comments_page", "comments_page_size"),
		Updates:            updates,
		ActorUserID:        actorUserID,
	}

	if detail.Post.Poll != nil {
		data.Poll = detail.Post.Poll
		results, err := h.service.GetPollResults(principal, detail.Post.Poll.ID)
		if err == nil {
			rows := make([]forumviews.PollOptionRowView, 0, len(results.Options))
			for _, o := range results.Options {
				rows = append(rows, forumviews.PollOptionRowView{Label: o.Label, Votes: o.Votes})
			}
			data.PollResults = &forumviews.PollResultsView{Options: rows}
			data.PollResultsOK = true
		} else if errors.Is(err, ErrResultsNotVisible) {
			data.PollResultsOK = false
		} else {
			return forumviews.PostPageData{}, err
		}
	}

	return data, nil
}

func (h *Handler) createCommunityPostView(c echo.Context) error {
	principal, err := h.actorFromRequest(c)
	if err != nil {
		return h.renderViewError(c, err)
	}
	input, err := communityPostInputFromForm(c)
	if err != nil {
		return h.renderViewError(c, err)
	}
	if _, err := h.service.CreateCommunityActivity(principal, input); err != nil {
		return h.renderViewError(c, err)
	}
	data, err := h.buildFeedPageData(c, principal)
	if err != nil {
		return h.renderViewError(c, err)
	}
	return renderComponent(c, forumviews.FeedListFragment(data))
}

func communityPostInputFromForm(c echo.Context) (CommunityPostInput, error) {
	input := CommunityPostInput{
		Title:    c.FormValue("title"),
		Body:     c.FormValue("body"),
		Tags:     c.FormValue("tags"),
		Location: c.FormValue("location"),
		Publish:  c.FormValue("publish") == "true" || c.FormValue("publish") == "on",
	}
	if v := strings.TrimSpace(c.FormValue("starts_at")); v != "" {
		t, err := time.Parse("2006-01-02T15:04", v)
		if err != nil {
			return CommunityPostInput{}, ErrValidation
		}
		input.StartsAt = &t
	}
	if v := strings.TrimSpace(c.FormValue("ends_at")); v != "" {
		t, err := time.Parse("2006-01-02T15:04", v)
		if err != nil {
			return CommunityPostInput{}, ErrValidation
		}
		input.EndsAt = &t
	}
	return input, nil
}

func (h *Handler) createCommentView(c echo.Context) error {
	principal, err := h.actorFromRequest(c)
	if err != nil {
		return h.renderViewError(c, err)
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.renderViewError(c, ErrValidation)
	}
	body := strings.TrimSpace(c.FormValue("body"))
	if body == "" {
		return h.renderViewError(c, ErrValidation)
	}
	if _, err := h.service.CreateComment(principal, postID, CommentInput{Body: body}); err != nil {
		return h.renderViewError(c, err)
	}
	commentParams := pageParamsNamed(c, "comments_page", "comments_page_size")
	comments, _, err := h.service.ListComments(principal, CommentListFilter{PostID: postID, Params: commentParams})
	if err != nil {
		return h.renderViewError(c, err)
	}
	return renderComponent(c, forumviews.CommentsList(comments))
}

func (h *Handler) togglePostReactionView(c echo.Context) error {
	principal, err := h.actorFromRequest(c)
	if err != nil {
		return h.renderViewError(c, err)
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.renderViewError(c, ErrValidation)
	}
	_, count, err := h.service.ToggleReaction(principal, fm.ReactionTargetPost, postID)
	if err != nil {
		return h.renderViewError(c, err)
	}
	detail, err := h.service.GetPost(principal, postID, false)
	if err != nil {
		return h.renderViewError(c, err)
	}
	liked := detail.Aggregates.ViewerReaction != nil
	return renderComponent(c, forumviews.PostReaction(forumviews.ReactionViewData{
		PostID:      postID,
		Count:       count,
		ViewerLiked: liked,
	}))
}

func (h *Handler) setAttendanceView(c echo.Context) error {
	principal, err := h.actorFromRequest(c)
	if err != nil {
		return h.renderViewError(c, err)
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.renderViewError(c, ErrValidation)
	}
	intent := fm.AttendanceIntent(c.FormValue("intent"))
	_, agg, err := h.service.SetAttendanceIntent(principal, postID, AttendanceInput{Intent: intent})
	if err != nil {
		return h.renderViewError(c, err)
	}
	var viewerIntent *fm.AttendanceIntent
	if agg.ViewerAttendanceIntent != nil {
		viewerIntent = agg.ViewerAttendanceIntent
	}
	aggView := forumviews.PostAggregatesView{
		ReactionCount: agg.ReactionCount,
		CommentCount:  agg.CommentCount,
		GoingCount:    agg.GoingCount,
		NotGoingCount: agg.NotGoingCount,
		ViewerLiked:   agg.ViewerReaction != nil,
	}
	if agg.ViewerAttendanceIntent != nil {
		aggView.ViewerAttendanceIntent = agg.ViewerAttendanceIntent
	}
	return renderComponent(c, forumviews.PostAttendance(forumviews.AttendanceViewData{
		PostID:     postID,
		Aggregates: aggView,
		Intent:     viewerIntent,
	}))
}

func (h *Handler) castPollVoteView(c echo.Context) error {
	principal, err := h.actorFromRequest(c)
	if err != nil {
		return h.renderViewError(c, err)
	}
	pollID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.renderViewError(c, ErrValidation)
	}
	optionIDs, err := pollOptionIDsFromForm(c)
	if err != nil {
		return h.renderViewError(c, err)
	}
	if err := h.service.CastPollVote(principal, pollID, PollVoteInput{OptionIDs: optionIDs}); err != nil {
		return h.renderViewError(c, err)
	}

	poll, err := h.service.GetPollByID(principal, pollID)
	if err != nil {
		return h.renderViewError(c, err)
	}
	data, err := h.buildPostPageData(c, principal, poll.ForumPostID, false, actorUserIDFromRequest(c))
	if err != nil {
		return h.renderViewError(c, err)
	}
	return renderComponent(c, forumviews.PollSectionFragment(data))
}

func pollOptionIDsFromForm(c echo.Context) ([]uuid.UUID, error) {
	if v := c.FormValue("option_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return nil, ErrValidation
		}
		return []uuid.UUID{id}, nil
	}
	values := c.Request().Form["option_ids"]
	if len(values) == 0 {
		return nil, ErrValidation
	}
	ids := make([]uuid.UUID, 0, len(values))
	for _, v := range values {
		id, err := uuid.Parse(v)
		if err != nil {
			return nil, ErrValidation
		}
		ids = append(ids, id)
	}
	return ids, nil
}
