package maintenance

import (
	adm "dorm-man/internal/models/administration"
	crosscutting "dorm-man/internal/models/cross-cutting"

	"github.com/google/uuid"
)

type Ticket struct {
	crosscutting.BaseModel
	RoomID          *uuid.UUID `gorm:"type:uuid;index"`
	Category        Category   `gorm:"type:varchar(20);not null;index"`
	Severity        Severity   `gorm:"type:varchar(20);not null;index"`
	Impact          Impact     `gorm:"type:varchar(30);not null;index"`
	Status          Status     `gorm:"type:varchar(20);not null;index"`
	Description     string     `gorm:"type:text;not null"`
	CreatedByUserID uuid.UUID  `gorm:"type:uuid;not null;index"`

	Room              *adm.Room            `gorm:"foreignKey:RoomID;references:ID"`
	CreatedByUser     crosscutting.User    `gorm:"foreignKey:CreatedByUserID;references:ID"`
	StatusTransitions []TicketStatusChange `gorm:"foreignKey:TicketID"`
}
