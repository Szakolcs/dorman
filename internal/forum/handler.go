package forum

import (
	"errors"
	"net/http"
	"strconv"

	"dorm-man/internal/administration"

	fm "dorm-man/internal/models/forum"

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
	value := c.Request().Header.Get("X-Actor-User-ID")
	if value == "" {
		return administration.Principal{}, ErrUnauthorized
	}
	id, err := uuid.Parse(value)
	if err != nil {
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
	filter := FeedListFilter{
		Kind:          fm.ForumPostKind(c.QueryParam("kind")),
		OfficialOnly:  parseBoolQuery(c, "official_only"),
		CommunityOnly: parseBoolQuery(c, "community_only"),
		Sort:          FeedSort(c.QueryParam("sort")),
		Limit:         queryPositiveInt(c, "limit", 50),
		Offset:        queryPositiveInt(c, "offset", 0),
		IncludeHidden: parseBoolQuery(c, "include_hidden"),
	}
	posts, err := h.service.ListFeed(principal, filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": posts})
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
	list, err := h.service.ListComments(principal, CommentListFilter{
		PostID: postID,
		Limit:  queryPositiveInt(c, "limit", 100),
		Offset: queryPositiveInt(c, "offset", 0),
	})
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": list})
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
