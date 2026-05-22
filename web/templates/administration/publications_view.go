package adminviews

import (
	"fmt"

	"dorm-man/internal/models"

	"github.com/google/uuid"
)

// PublicationListItem is a publication row with its forum kind for list UIs.
type PublicationListItem struct {
	models.Publication
	Kind     string
	Activity *ActivityStats
	Event    *EventInterestStats
}

type ActivityStats struct {
	Capacity    int
	BookedCount int
}

type EventInterestStats struct {
	Interested    int
	NotInterested int
	Busy          int
}

func publicationKindFilterURL(kind string, activeKind string) string {
	if activeKind == kind {
		return "/administration/publications"
	}
	return "/administration/publications?kind=" + kind
}

func publicationKindLabel(kind string) string {
	switch kind {
	case "news":
		return "News"
	case "activity":
		return "Activity"
	case "event":
		return "Event"
	default:
		return kind
	}
}

func activityStatsLabel(stats *ActivityStats) string {
	if stats == nil {
		return ""
	}
	spots := stats.Capacity - stats.BookedCount
	if spots < 0 {
		spots = 0
	}
	return fmt.Sprintf("%d / %d booked · %d spots left", stats.BookedCount, stats.Capacity, spots)
}

func eventInterestStatsLabel(stats *EventInterestStats) string {
	if stats == nil {
		return ""
	}
	return fmt.Sprintf(
		"%d interested · %d not interested · %d busy",
		stats.Interested,
		stats.NotInterested,
		stats.Busy,
	)
}

func publicationStateClass(state models.PublicationState) string {
	switch state {
	case models.PublicationStatePublished:
		return "dm-publication-status-published"
	case models.PublicationStateArchived:
		return "dm-publication-status-archived"
	case models.PublicationStateHidden:
		return "dm-publication-status-hidden"
	case models.PublicationStateCanceled:
		return "dm-publication-status-canceled"
	case models.PublicationStatePostponed:
		return "dm-publication-status-postponed"
	default:
		return "dm-publication-status-draft"
	}
}

func publicationStateUpdateURL(kind string, id uuid.UUID) string {
	return fmt.Sprintf("/administration/publications/%s/%s/state", publicationRouteSegment(kind), id.String())
}

func publicationArchiveURL(kind string, id uuid.UUID) string {
	return fmt.Sprintf("/administration/publications/%s/%s", publicationRouteSegment(kind), id.String())
}

func publicationRouteSegment(kind string) string {
	switch kind {
	case "news":
		return "news"
	case "activity":
		return "activities"
	case "event":
		return "events"
	default:
		return kind
	}
}

func publicationCanPublish(state models.PublicationState) bool {
	switch state {
	case models.PublicationStateDraft, models.PublicationStateHidden:
		return true
	default:
		return false
	}
}

func publicationCanArchive(state models.PublicationState) bool {
	return state != models.PublicationStateArchived
}
