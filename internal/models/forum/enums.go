package forum

import (
	"dorm-man/internal/models/crosscutting"
)

type PublicationState = crosscutting.PublicationState

const (
	PublicationStateDraft     PublicationState = crosscutting.PublicationStateDraft
	PublicationStatePublished PublicationState = crosscutting.PublicationStatePublished
	PublicationStateArchived  PublicationState = crosscutting.PublicationStateArchived
	PublicationStateCanceled  PublicationState = crosscutting.PublicationStateCanceled
	PublicationStatePostponed PublicationState = crosscutting.PublicationStatePostponed
	PublicationStateHidden    PublicationState = crosscutting.PublicationStateHidden
)

// EventAttendanceIntent is what a user signals about attending an event.
type EventAttendanceIntent string

const (
	EventAttendanceInterested    EventAttendanceIntent = "interested"
	EventAttendanceNotInterested EventAttendanceIntent = "not_interested"
	EventAttendanceBusy          EventAttendanceIntent = "busy"
)
