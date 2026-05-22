package maintenance

import (
	"errors"
	"time"

	"dorm-man/internal/models"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New("authorization denied")
	ErrNotFound     = errors.New("not found")
	ErrValidation   = errors.New("validation error")
)

type TicketFilter struct {
	Statuses   []models.Status
	Categories []models.Category
	Severities []models.Severity
	From       time.Time
	To         time.Time
	FlatID     *uuid.UUID
}

type CreateTicketRequest struct {
	Category    models.Category `json:"category"    form:"category"`
	Severity    models.Severity `json:"severity"    form:"severity"`
	Impact      models.Impact   `json:"impact"      form:"impact"`
	Description string          `json:"description" form:"description"`
	FlatID      *uuid.UUID      `json:"flat_id"     form:"flat_id"`
}

type UpdateTicketRequest struct {
	TicketID    uuid.UUID       `json:"ticket_id"   form:"ticket_id"   param:"id"`
	Status      models.Status   `json:"status"      form:"status"`
	Category    models.Category `json:"category"    form:"category"`
	Severity    models.Severity `json:"severity"    form:"severity"`
	Impact      models.Impact   `json:"impact"      form:"impact"`
	Description string          `json:"description" form:"description"`
	Note        string          `json:"note"        form:"note"`
}

type TenantLocation struct {
	RoomID *uuid.UUID
	FlatID *uuid.UUID
}
