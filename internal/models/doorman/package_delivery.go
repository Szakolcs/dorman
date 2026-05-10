package models

import (
	"time"

	adm "dorm-man/internal/models/administration"
	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type PackageStatus string

const (
	PackageStatusReceived PackageStatus = "received"
	PackageStatusNotified PackageStatus = "notified"
	PackageStatusPickedUp PackageStatus = "picked_up"
)

type PackageNotificationChannel string

const (
	PackageNotificationChannelInApp PackageNotificationChannel = "in_app"
	PackageNotificationChannelEmail PackageNotificationChannel = "email"
)

type PackageNotificationStatus string

const (
	PackageNotificationStatusQueued PackageNotificationStatus = "queued"
	PackageNotificationStatusSent   PackageNotificationStatus = "sent"
	PackageNotificationStatusFailed PackageNotificationStatus = "failed"
)

type Package struct {
	platform.BaseModel
	RecipientLabel     string        `gorm:"not null;index"`
	Description        string        `gorm:"type:text"`
	TenantID           *uuid.UUID    `gorm:"type:uuid;index"`
	Status             PackageStatus `gorm:"type:varchar(20);not null;default:'received';index"`
	ReceivedAt         time.Time     `gorm:"not null;index"`
	RegisteredByUserID *uuid.UUID    `gorm:"type:uuid;index"`
	PickedUpAt         *time.Time    `gorm:"index"`
	PickedUpByUserID   *uuid.UUID    `gorm:"type:uuid;index"`

	Tenant               *adm.Tenant           `gorm:"foreignKey:TenantID;references:ID"`
	RegisteredByUser     *adm.User             `gorm:"foreignKey:RegisteredByUserID;references:ID"`
	PickedUpByUser       *adm.User             `gorm:"foreignKey:PickedUpByUserID;references:ID"`
	PackageNotifications []PackageNotification `gorm:"foreignKey:PackageID"`
}

type PackageNotification struct {
	platform.BaseModel
	PackageID     uuid.UUID                  `gorm:"type:uuid;not null;index"`
	TenantID      uuid.UUID                  `gorm:"type:uuid;not null;index"`
	Channel       PackageNotificationChannel `gorm:"type:varchar(20);not null;index"`
	Status        PackageNotificationStatus  `gorm:"type:varchar(20);not null;default:'queued';index"`
	SentAt        *time.Time                 `gorm:"index"`
	FailureReason string                     `gorm:"type:text"`

	Package *Package    `gorm:"foreignKey:PackageID;references:ID"`
	Tenant  *adm.Tenant `gorm:"foreignKey:TenantID;references:ID"`
}
