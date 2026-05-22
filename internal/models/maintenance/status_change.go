package maintenance

import (
	"dorm-man/internal/models/crosscutting"

	"github.com/google/uuid"
)

type StatusChange struct {
	crosscutting.BaseModel
	TicketID   uuid.UUID `gorm:"type:uuid;not null;index"`
	FromStatus *Status   `gorm:"type:varchar(20);index"`
	ToStatus   Status    `gorm:"type:varchar(20);not null;index"`
	Note       string    `gorm:"type:text"`

	Ticket Ticket `gorm:"foreignKey:TicketID;references:ID"`
}
