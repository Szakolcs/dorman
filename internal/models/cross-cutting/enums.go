package cross_cutting

type PublicationState string

const (
	PublicationStateDraft     PublicationState = "draft"
	PublicationStatePublished PublicationState = "published"
	PublicationStateArchived  PublicationState = "archived"
	PublicationStateCanceled  PublicationState = "canceled"
	PublicationStatePostponed PublicationState = "postponed"
	PublicationStateHidden    PublicationState = "hidden"
)
