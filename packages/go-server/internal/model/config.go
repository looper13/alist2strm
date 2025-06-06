package model

import (
	"time"
)

type Config struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `gorm:"unique;not null" json:"name"`
	Code      string    `gorm:"unique;not null" json:"code"`
	Value     string    `gorm:"not null" json:"value"`
}
