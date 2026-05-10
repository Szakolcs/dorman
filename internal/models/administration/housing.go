package models

import (
	"time"

	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type Building struct {
	platform.BaseModel
	Name string `gorm:"not null;uniqueIndex"`
	Code string `gorm:"uniqueIndex"`

	Flats      []Flat          `gorm:"foreignKey:BuildingID"`
	Activities []Activity      `gorm:"foreignKey:BuildingID"`
	Events     []Event         `gorm:"foreignKey:BuildingID"`
	Inventory  []InventoryItem `gorm:"foreignKey:BuildingID"`
}

type Flat struct {
	platform.BaseModel
	BuildingID *uuid.UUID `gorm:"type:uuid;index"`
	Name       string     `gorm:"not null;index"`
	Floor      int        `gorm:"not null;default:0"`

	Building  *Building       `gorm:"foreignKey:BuildingID;references:ID"`
	Rooms     []Room          `gorm:"foreignKey:FlatID"`
	Inventory []InventoryItem `gorm:"foreignKey:FlatID"`
}

type Room struct {
	platform.BaseModel
	FlatID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Number     string    `gorm:"not null;index"`
	Capacity   int       `gorm:"not null;default:1"`
	IsArchived bool      `gorm:"not null;default:false"`

	Flat               Flat                `gorm:"foreignKey:FlatID;references:ID"`
	Assignments        []RoomAssignment    `gorm:"foreignKey:RoomID"`
	InventoryItems     []InventoryItem     `gorm:"foreignKey:RoomID"`
	MaintenanceTickets []MaintenanceTicket `gorm:"foreignKey:RoomID"`
	OperationalJobs    []OperationalJob    `gorm:"foreignKey:RoomID"`
}

type DegreeType string

const (
	DegreeBSc DegreeType = "bsc"
	DegreeBA  DegreeType = "ba"
	DegreeMSc DegreeType = "msc"
	DegreeMA  DegreeType = "ma"
	DegreePhD DegreeType = "phd"
)

type SexType string

const (
	SexFemale SexType = "female"
	SexMale   SexType = "male"
	SexOther  SexType = "other"
)

type NationalityType string

const (
	NationalityHungarian     NationalityType = "hungarian"
	NationalityInternational NationalityType = "international"
)

type FacultyType string

const (
	FacultyScience     FacultyType = "science"
	FacultyHumanities  FacultyType = "humanities"
	FacultyEngineering FacultyType = "engineering"
	FacultyMedicine    FacultyType = "medicine"
)

type Tenant struct {
	platform.BaseModel
	UserID       *uuid.UUID       `gorm:"type:uuid;index"`
	StudentCode  string           `gorm:"uniqueIndex;not null"`
	Name         string           `gorm:"not null;index"`
	Email        string           `gorm:"index"`
	Degree       *DegreeType      `gorm:"type:varchar(10);index"`
	Faculty      *FacultyType     `gorm:"index"`
	Age          *int             `gorm:"index"`
	Sex          *SexType         `gorm:"type:varchar(10);index"`
	Nationality  *NationalityType `gorm:"not null;index"`
	IsActive     bool             `gorm:"not null;default:true;index"`
	RegisteredAt time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP"`

	User            *User            `gorm:"foreignKey:UserID;references:ID"`
	RoomAssignments []RoomAssignment `gorm:"foreignKey:TenantID"`
}
