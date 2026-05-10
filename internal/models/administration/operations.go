package models

import (
	"time"

	"github.com/google/uuid"
)

type JobPriority string

const (
	JobPriorityLow    JobPriority = "low"
	JobPriorityMedium JobPriority = "medium"
	JobPriorityHigh   JobPriority = "high"
)

type JobStatus string

const (
	JobStatusPlanned    JobStatus = "planned"
	JobStatusScheduled  JobStatus = "scheduled"
	JobStatusInProgress JobStatus = "in_progress"
	JobStatusFinished   JobStatus = "finished"
	JobStatusCanceled   JobStatus = "canceled"
)

type OperationalJob struct {
	BaseModel
	Title           string `gorm:"not null;index"`
	Description     string
	RoomID          *uuid.UUID  `gorm:"type:uuid;index"`
	AssigneeUserID  uuid.UUID   `gorm:"type:uuid;not null;index"`
	CreatedByUserID *uuid.UUID  `gorm:"type:uuid;index"`
	StartsAt        time.Time   `gorm:"not null;index"`
	EndsAt          time.Time   `gorm:"not null;index"`
	Priority        JobPriority `gorm:"type:varchar(20);not null;default:'medium';index"`
	Status          JobStatus   `gorm:"type:varchar(20);not null;default:'planned';index"`

	Room          *Room `gorm:"foreignKey:RoomID;references:ID"`
	AssigneeUser  User  `gorm:"foreignKey:AssigneeUserID;references:ID"`
	CreatedByUser *User `gorm:"foreignKey:CreatedByUserID;references:ID"`
}
