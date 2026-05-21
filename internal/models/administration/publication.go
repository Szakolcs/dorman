package models

import (
	"dorm-man/internal/models/cross-cutting"
	"time"

	"github.com/google/uuid"
)

type Activity struct {
	cross_cutting.BaseModel
	Title        string           `gorm:"not null;index"`
	Description  string           `gorm:"type:text"`
	SharedAreaID *uuid.UUID       `gorm:"type:uuid;index"`
	Capacity     int              `gorm:"not null;default:1"`
	State        PublicationState `gorm:"type:varchar(20);not null;default:'draft';index"`

	Building *SharedArea `gorm:"foreignKey:BuildingID;references:ID"`
}

type Event struct {
	cross_cutting.BaseModel
	Title        string           `gorm:"not null;index"`
	Description  string           `gorm:"type:text"`
	SharedAreaID *uuid.UUID       `gorm:"type:uuid;index"`
	StartsAt     time.Time        `gorm:"not null;index"`
	EndsAt       time.Time        `gorm:"not null;index"`
	State        PublicationState `gorm:"type:varchar(20);not null;default:'draft';index"`

	SharedArea *SharedArea `gorm:"foreignKey:SharedAreaID;references:ID"`
}
