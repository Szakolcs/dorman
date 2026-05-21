package models

import (
	crosscutting "dorm-man/internal/models/cross-cutting"
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	crosscutting.BaseModel
	UserID       *uuid.UUID       `gorm:"type:uuid;index"`
	StudentCode  string           `gorm:"uniqueIndex;not null"`
	Degree       *DegreeType      `gorm:"type:varchar(10);index"`
	Faculty      *FacultyType     `gorm:"index"`
	Age          *int             `gorm:"index"`
	Sex          *SexType         `gorm:"type:varchar(10);index"`
	Nationality  *NationalityType `gorm:"index"`
	IsActive     bool             `gorm:"not null;default:true;index"`
	RegisteredAt time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP"`

	User            *crosscutting.User `gorm:"foreignKey:UserID;references:ID"`
	RoomAssignments []RoomAssignment   `gorm:"foreignKey:TenantID"`
}
