package models

import (
	"time"

	"github.com/google/uuid"
)

type MaintenanceCategory string

const (
	MaintenanceCategoryYearly   MaintenanceCategory = "yearly"
	MaintenanceCategoryMonthly  MaintenanceCategory = "monthly"
	MaintenanceCategoryWeekly   MaintenanceCategory = "weekly"
	MaintenanceCategoryIncident MaintenanceCategory = "incident"
)

type MaintenanceSeverity string

const (
	MaintenanceSeverityHigh   MaintenanceSeverity = "high"
	MaintenanceSeverityMedium MaintenanceSeverity = "medium"
	MaintenanceSeverityLow    MaintenanceSeverity = "low"
)

type MaintenanceImpact string

const (
	MaintenanceImpactLifeThreatening MaintenanceImpact = "life_threatening"
	MaintenanceImpactAffectsDaily    MaintenanceImpact = "affects_daily_life"
	MaintenanceImpactInconvenience   MaintenanceImpact = "inconvenience"
	MaintenanceImpactBeautyFlaw      MaintenanceImpact = "beauty_flaw"
)

type MaintenanceStatus string

const (
	MaintenanceStatusReported   MaintenanceStatus = "reported"
	MaintenanceStatusDuplicate  MaintenanceStatus = "duplicate"
	MaintenanceStatusInProgress MaintenanceStatus = "in_progress"
	MaintenanceStatusHalted     MaintenanceStatus = "halted"
	MaintenanceStatusResolved   MaintenanceStatus = "resolved"
	MaintenanceStatusClosed     MaintenanceStatus = "closed"
)

type MaintenanceTicket struct {
	BaseModel
	RoomID          *uuid.UUID          `gorm:"type:uuid;index"`
	BuildingID      *uuid.UUID          `gorm:"type:uuid;index"`
	Category        MaintenanceCategory `gorm:"type:varchar(20);not null;index"`
	Severity        MaintenanceSeverity `gorm:"type:varchar(20);not null;index"`
	Impact          MaintenanceImpact   `gorm:"type:varchar(30);not null;index"`
	Status          MaintenanceStatus   `gorm:"type:varchar(20);not null;index"`
	Description     string              `gorm:"type:text;not null"`
	CreatedByUserID uuid.UUID           `gorm:"type:uuid;not null;index"`
	AssigneeUserID  *uuid.UUID          `gorm:"type:uuid;index"`
	DueAt           *time.Time          `gorm:"index"`
	ClosedAt        *time.Time          `gorm:"index"`

	Room              *Room                `gorm:"foreignKey:RoomID;references:ID"`
	Building          *Building            `gorm:"foreignKey:BuildingID;references:ID"`
	CreatedByUser     User                 `gorm:"foreignKey:CreatedByUserID;references:ID"`
	AssigneeUser      *User                `gorm:"foreignKey:AssigneeUserID;references:ID"`
	StatusTransitions []TicketStatusChange `gorm:"foreignKey:TicketID"`
}

type TicketStatusChange struct {
	BaseModel
	TicketID    uuid.UUID          `gorm:"type:uuid;not null;index"`
	ActorUserID uuid.UUID          `gorm:"type:uuid;not null;index"`
	FromStatus  *MaintenanceStatus `gorm:"type:varchar(20);index"`
	ToStatus    MaintenanceStatus  `gorm:"type:varchar(20);not null;index"`
	Note        string             `gorm:"type:text"`
	OccurredAt  time.Time          `gorm:"not null;default:CURRENT_TIMESTAMP;index"`

	Ticket    MaintenanceTicket `gorm:"foreignKey:TicketID;references:ID"`
	ActorUser User              `gorm:"foreignKey:ActorUserID;references:ID"`
}
