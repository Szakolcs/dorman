package models

import (
	"time"

	"github.com/google/uuid"
)

type PrincipalType string

const (
	PrincipalTypeStaff  PrincipalType = "staff"
	PrincipalTypeTenant PrincipalType = "tenant"
)

type RoleName string

const (
	RoleAdministrator RoleName = "administrator"
	RoleOfficeWorker  RoleName = "office_worker"
	RoleDirector      RoleName = "director"
	RoleDoorman       RoleName = "doorman"
	RoleTenant        RoleName = "tenant"
)

type User struct {
	BaseModel
	UniCode       string        `gorm:"uniqueIndex;not null"`
	Email         string        `gorm:"not null"`
	PasswordHash  string        `gorm:"not null"`
	Name          string        `gorm:"not null"`
	PrincipalType PrincipalType `gorm:"type:varchar(20);not null;index"`
	IsActive      bool          `gorm:"not null;default:true"`
	LastLoginAt   *time.Time

	UserRoles []UserRole `gorm:"foreignKey:UserID"`
}

type Role struct {
	BaseModel
	Name        RoleName `gorm:"type:varchar(40);uniqueIndex;not null"`
	Description string

	UserRoles []UserRole `gorm:"foreignKey:RoleID"`
}

type UserRole struct {
	BaseModel
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:idx_user_role_unique"`
	RoleID     uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:idx_user_role_unique"`
	AssignedBy *uuid.UUID `gorm:"type:uuid;index"`
	AssignedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	RevokedAt  *time.Time `gorm:"index"`

	User     User  `gorm:"foreignKey:UserID;references:ID"`
	Role     Role  `gorm:"foreignKey:RoleID;references:ID"`
	Assigner *User `gorm:"foreignKey:AssignedBy;references:ID"`
}
