package chat

const (
	maxMessageBodyLen = 10_000
	maxBioLen         = 2_000
	maxGroupTitleLen  = 200
	maxPreviewLen     = 500
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}
