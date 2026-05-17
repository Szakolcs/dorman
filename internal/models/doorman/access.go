package models

import (
	"time"

	adm "dorm-man/internal/models/administration"
	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

// TenantEntryToken stores material used to validate printed or in-app QR entry codes.
type TenantEntryToken struct {
	platform.BaseModel
	TenantID   uuid.UUID  `gorm:"type:uuid;not null;index"`
	PublicRef  string     `gorm:"not null;uniqueIndex"`    // stable id embedded in QR payload
	SecretHash *string    `gorm:"type:varchar(255);index"` // optional; verifier may hash PublicRef instead
	ExpiresAt  *time.Time `gorm:"index"`
	RevokedAt  *time.Time `gorm:"index"`

	Tenant *adm.Tenant `gorm:"foreignKey:TenantID;references:ID"`
}

type AccessEventOutcome string

const (
	AccessEventOutcomeGranted AccessEventOutcome = "granted"
	AccessEventOutcomeDenied  AccessEventOutcome = "denied"
)

// QR / validation failure reasons per doorman specification error semantics.
const (
	AccessDenialReasonInvalid = "invalid"
	AccessDenialReasonExpired = "expired"
	AccessDenialReasonRevoked = "revoked"
)

type AccessEventSource string

const (
	AccessEventSourceQRScan AccessEventSource = "qr_scan"
	AccessEventSourceManual AccessEventSource = "manual"
)

// AccessEvent is an immutable-oriented log row for dorm entry attempts (QR or manual).
type AccessEvent struct {
	platform.BaseModel
	TenantID    *uuid.UUID         `gorm:"type:uuid;index"`
	ActorUserID *uuid.UUID         `gorm:"type:uuid;index"`
	TokenID     *uuid.UUID         `gorm:"type:uuid;index"`
	Outcome     AccessEventOutcome `gorm:"type:varchar(20);not null;index"`
	Reason      *string            `gorm:"type:varchar(64);index"`
	Source      AccessEventSource  `gorm:"type:varchar(20);not null;index"`
	OccurredAt  time.Time          `gorm:"not null;index;default:CURRENT_TIMESTAMP"`

	Tenant     *adm.Tenant       `gorm:"foreignKey:TenantID;references:ID"`
	ActorUser  *adm.User         `gorm:"foreignKey:ActorUserID;references:ID"`
	EntryToken *TenantEntryToken `gorm:"foreignKey:TokenID;references:ID"`
}
