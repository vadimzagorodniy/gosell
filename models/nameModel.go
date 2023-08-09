package models

import (
	"time"
)

type Name struct {
	UID       uint `gorm:"primaryKey"`
	FirstName string
	LastName  string
	FullName  string `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
