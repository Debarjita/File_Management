package models

import (
	"time"

	"gorm.io/gorm"
)

type File struct {
	ID        uint   `gorm:"primaryKey"`
	FileName  string `gorm:"not null"`
	FileType  string `gorm:"not null"`
	Size      int64  `gorm:"not null"`
	URL       string `gorm:"not null"`
	UserID    uint   `gorm:"not null"` // Foreign key
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	ExpiresAt *time.Time     `gorm:"column:expires_at" json:"expires_at"`
}
