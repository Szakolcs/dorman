package maintenance

import (
	"dorm-man/internal/models/administration"
	"dorm-man/internal/models/crosscutting"

	"github.com/google/uuid"
)

type Ticket struct {
	crosscutting.BaseModel
	FlatID          *uuid.UUID `gorm:"type:uuid;index"`
	Category        Category   `gorm:"type:varchar(20);not null;index"`
	Severity        Severity   `gorm:"type:varchar(20);not null;index"`
	Impact          Impact     `gorm:"type:varchar(30);not null;index"`
	Status          Status     `gorm:"type:varchar(20);not null;index"`
	Description     string     `gorm:"type:text;not null"`
	CreatedByUserID uuid.UUID  `gorm:"type:uuid;not null;index"`

	CreatedByUser     *crosscutting.User   `gorm:"foreignKey:CreatedByUserID;references:ID"`
	Flat              *administration.Flat `gorm:"foreignKey:FlatID;references:ID"`
	StatusTransitions []StatusChange       `gorm:"foreignKey:TicketID"`
}
