package tenantviews

import (
	"fmt"
	"net/url"
	"strings"

	"dorm-man/internal/administration"
	"dorm-man/internal/models"

	"github.com/google/uuid"
)

type PublicationCardItem struct {
	administration.PublicationListItem
	BookedByViewer    bool
	ViewerEventIntent string
}

func publicationCardID(id uuid.UUID) string {
	return "publication-" + id.String()
}

func activityDetailCardID(id uuid.UUID) string {
	return "activity-detail-" + id.String()
}

func activeKindFromReturn(returnURL string) string {
	if returnURL == "" {
		return ""
	}
	u, err := url.Parse(returnURL)
	if err != nil {
		return ""
	}
	return u.Query().Get("kind")
}

func activityBookTargetID(activityID uuid.UUID, returnURL string) string {
	if strings.Contains(returnURL, "/publications/") {
		return activityDetailCardID(activityID)
	}
	return publicationCardID(activityID)
}

func ActiveKindFromReturn(returnURL string) string {
	return activeKindFromReturn(returnURL)
}

func publicationKindFilterURL(kind string, activeKind string) string {
	if activeKind == kind {
		return "/tenant"
	}
	if kind == "" {
		return "/tenant"
	}
	return "/tenant?kind=" + kind
}

func publicationKindLabel(kind administration.PublicationKind) string {
	switch kind {
	case administration.PublicationKindNews:
		return "News"
	case administration.PublicationKindActivity:
		return "Activity"
	case administration.PublicationKindEvent:
		return "Event"
	default:
		return string(kind)
	}
}

func activityStatsLabel(stats *administration.ActivityStats) string {
	if stats == nil {
		return ""
	}
	spots := stats.Capacity - stats.BookedCount
	if spots < 0 {
		spots = 0
	}
	return fmt.Sprintf("%d / %d booked · %d spots left", stats.BookedCount, stats.Capacity, spots)
}

func eventInterestStatsLabel(stats *administration.EventInterestStats) string {
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

func publicationKindClass(kind administration.PublicationKind) string {
	switch kind {
	case administration.PublicationKindNews:
		return "dm-tenant-kind-news"
	case administration.PublicationKindActivity:
		return "dm-tenant-kind-activity"
	case administration.PublicationKindEvent:
		return "dm-tenant-kind-event"
	default:
		return ""
	}
}

func activityBookURL(activityID uuid.UUID, returnURL string) string {
	return actionURL("/tenant/publications/activities/"+activityID.String()+"/book", returnURL)
}

func activityCancelURL(activityID uuid.UUID, returnURL string) string {
	return actionURL("/tenant/publications/activities/"+activityID.String()+"/cancel", returnURL)
}

func eventIntentPostURL(eventID uuid.UUID, returnURL string) string {
	return actionURL("/tenant/publications/events/"+eventID.String()+"/intent", returnURL)
}

func actionURL(path string, returnURL string) string {
	if returnURL != "" {
		sep := "?"
		if strings.Contains(path, "?") {
			sep = "&"
		}
		return path + sep + "return=" + url.QueryEscape(returnURL)
	}
	return path
}

func eventIntentIsActive(intent models.EventAttendanceIntent, viewerIntent string) bool {
	return viewerIntent != "" && string(intent) == viewerIntent
}

func eventIntentClass(intent models.EventAttendanceIntent) string {
	switch intent {
	case models.EventAttendanceInterested:
		return "dm-event-intent-interested"
	case models.EventAttendanceNotInterested:
		return "dm-event-intent-not-interested"
	case models.EventAttendanceBusy:
		return "dm-event-intent-busy"
	default:
		return ""
	}
}

func activitySpotsLeft(stats *administration.ActivityStats) int {
	if stats == nil {
		return 0
	}
	spots := stats.Capacity - stats.BookedCount
	if spots < 0 {
		return 0
	}
	return spots
}

func publicationListReturnURL(activeKind string) string {
	if activeKind == "" {
		return "/tenant"
	}
	return "/tenant?kind=" + activeKind
}

func activityDetailReturnURL(activityID uuid.UUID) string {
	return "/tenant/publications/" + activityID.String()
}

func eventDetailReturnURL(eventID uuid.UUID) string {
	return "/tenant/publications/" + eventID.String()
}

func activityStatsFromActivity(activity models.Activity) *administration.ActivityStats {
	return &administration.ActivityStats{
		Capacity:    activity.Capacity,
		BookedCount: activity.BookedCount,
	}
}

func activitySpotsFromActivity(activity models.Activity) int {
	return activitySpotsLeft(activityStatsFromActivity(activity))
}
