package forum

import "errors"

var (
	ErrUnauthorized        = errors.New("authorization denied")
	ErrNotFound            = errors.New("not found")
	ErrValidation          = errors.New("validation error")
	ErrPollClosed          = errors.New("poll closed")
	ErrAlreadyVoted        = errors.New("already voted")
	ErrIneligibleVoter     = errors.New("ineligible voter")
	ErrResultsNotVisible   = errors.New("results not visible")
	ErrEditNotAllowed      = errors.New("edit not allowed")
	ErrConcurrencyConflict = errors.New("concurrency conflict")
)
