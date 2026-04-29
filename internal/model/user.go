package model

import "time"

type User struct {
	ID           uint    `gorm:"primaryKey"`
	Phone        string  `gorm:"uniqueIndex;type:varchar(20);not null"`
	PasswordHash string  `gorm:"type:varchar(255);not null"`
	Balance      float64 `gorm:"type:decimal(18,4);default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
