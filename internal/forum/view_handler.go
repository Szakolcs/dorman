package forum

import (
	"errors"
	"strings"
	"time"

	"dorm-man/internal/administration"

	forumviews "dorm-man/web/templates/forum"

	fm "dorm-man/internal/models/forum"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func isHtmx(c echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}

func actorUserIDFromRequest(c echo.Context) string {
	if v := c.Request().Header.Get("X-Actor-User-ID"); v != "" {
		return v
	}
	if v := c.FormValue("actor_user_id"); v != "" {
		return v
	}
	return c.QueryParam("actor_user_id")
}

func (h *Handler) renderViewError(c echo.Context, err error) error {
	msg := err.Error()
	if isHtmx(c) {
		return renderComponent(c, forumviews.ErrorAlert(msg))
	}
	return renderComponent(c, forumviews.ForumDocument("Forum error", actorUserIDFromRequest(c), forumviews.ErrorAlert(msg)))
}

func (h *Handler) feedPage(c echo.Context) error {
	principal, err := h.actorFromRequest(c)
	if err != nil {
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
	return renderComponent(c, forumviews.FeedList(data.Posts))
}

func (h *Handler) buildFeedPageData(c echo.Context, principal administration.Principal) (forumviews.FeedPageData, error) {
	filter := feedFilterFromRequest(c)
	posts, err := h.service.ListFeed(principal, filter)
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
		Limit:         queryPositiveInt(c, "limit", 50),
		Offset:        queryPositiveInt(c, "offset", 0),
		IncludeHidden: parseBoolQuery(c, "include_hidden"),
	}
}

func (h *Handler) postPage(c echo.Context) error {
	principal, err := h.actorFromRequest(c)
	if err != nil {
		return h.renderViewError(c, err)
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.renderViewError(c, ErrValidation)
	}
	data, err := h.buildPostPageData(principal, postID, true, actorUserIDFromRequest(c))
	if err != nil {
		return h.renderViewError(c, err)
	}
	return renderComponent(c, forumviews.PostPage(data))
}

func (h *Handler) buildPostPageData(principal administration.Principal, postID uuid.UUID, recordView bool, actorUserID string) (forumviews.PostPageData, error) {
	detail, err := h.service.GetPost(principal, postID, recordView)
	if err != nil {
		return forumviews.PostPageData{}, err
	}
	comments, err := h.service.ListComments(principal, CommentListFilter{PostID: postID, Limit: 200})
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
		Post:          detail.Post,
		Aggregates:    agg,
		Schedule:      detail.Post.Schedule,
		Comments:      comments,
		Updates:       updates,
		ActorUserID:   actorUserID,
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
	return renderComponent(c, forumviews.FeedList(data.Posts))
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
	comments, err := h.service.ListComments(principal, CommentListFilter{PostID: postID, Limit: 200})
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
	data, err := h.buildPostPageData(principal, poll.ForumPostID, false, actorUserIDFromRequest(c))
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
