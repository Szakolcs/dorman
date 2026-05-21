package administration

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized         = errors.New("authorization denied")
	ErrNotFound             = errors.New("not found")
	ErrValidation           = errors.New("validation error")
	ErrCapacityConflict     = errors.New("capacity conflict")
	ErrStateTransition      = errors.New("state transition invalid")
	ErrConcurrencyConflict  = errors.New("concurrency conflict")
	ErrStudentStatusInvalid = errors.New("student status invalid")
)
