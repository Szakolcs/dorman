package models

// PublicationState is the shared lifecycle for administration publications and forum posts.
type PublicationState string

const (
	PublicationStateDraft     PublicationState = "draft"
	PublicationStatePublished PublicationState = "published"
	PublicationStateArchived  PublicationState = "archived"
	PublicationStateCanceled  PublicationState = "canceled"
	PublicationStatePostponed PublicationState = "postponed"
	PublicationStateHidden    PublicationState = "hidden"
)
