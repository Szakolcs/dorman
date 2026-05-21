package models

import (
	crosscutting "dorm-man/internal/models/cross-cutting"
	"time"
)

type OperationalJob struct {
	crosscutting.BaseModel
	Title       string      `gorm:"not null;index"`
	Description string      `gorm:"type:text"`
	StartsAt    time.Time   `gorm:"not null;index"`
	EndsAt      time.Time   `gorm:"not null;index"`
	Priority    JobPriority `gorm:"type:varchar(20);not null;default:'medium';index"`
	Status      JobStatus   `gorm:"type:varchar(20);not null;default:'planned';index"`
}
