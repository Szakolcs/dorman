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
	Status   models.Status   `json:"status"    query:"status"`
	Category models.Category `json:"category"  query:"category"`
	Severity models.Severity `json:"severity"  query:"severity"`
	From     time.Time       `json:"from"      query:"from"`
	To       time.Time       `json:"to"        query:"to"`
	FlatID   *uuid.UUID      `json:"flat_id"   query:"flat_id"`
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
