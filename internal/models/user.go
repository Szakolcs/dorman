package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email string `gorm:"uniqueIndex;not null"`
	Name  string `gorm:"not null"`
}
