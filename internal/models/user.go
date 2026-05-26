package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	BaseModel
	Name         string     `gorm:"type:varchar(255);not null"`
	Email        string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	Nickname     string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string     `gorm:"not null"`
	LastLoginAt  *time.Time `gorm:"index"`
	AvatarURL    string     `gorm:"type:text"`
	PhotoUrl     string     `gorm:"type:text"`
	RoleID       uuid.UUID  `gorm:"type:uuid"`
	Role         Role       `gorm:"foreignKey:RoleID;references:ID"`
}

type Role struct {
	BaseModel
	Name         string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PermissionID uuid.UUID `gorm:"type:uuid"`

	Permission Permission `gorm:"foreignKey:PermissionID;references:ID"`
}

type Permission struct {
	BaseModel
	Name        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Read        bool      `gorm:"not null;default:true"`
	Write       bool      `gorm:"not null;default:true"`
	OperationID uuid.UUID `gorm:"type:uuid"`

	Operation Operation `gorm:"foreignKey:OperationID;references:ID"`
}

type Operation struct {
	BaseModel
	Name string `gorm:"type:varchar(255);uniqueIndex;not null"`
}
