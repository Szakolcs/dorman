package models

import "github.com/google/uuid"

type Ticket struct {
	BaseModel
	FlatID          *uuid.UUID `gorm:"type:uuid;index"`
	Category        Category   `gorm:"type:varchar(20);not null;index"`
	Severity        Severity   `gorm:"type:varchar(20);not null;index"`
	Impact          Impact     `gorm:"type:varchar(30);not null;index"`
	Status          Status     `gorm:"type:varchar(20);not null;index"`
	Description     string     `gorm:"type:text;not null"`
	CreatedByUserID uuid.UUID  `gorm:"type:uuid;not null;index"`

	CreatedByUser     *User          `gorm:"foreignKey:CreatedByUserID;references:ID"`
	Flat              *Flat          `gorm:"foreignKey:FlatID;references:ID"`
	StatusTransitions []StatusChange `gorm:"foreignKey:TicketID"`
}
