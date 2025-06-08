package configs

import (
	"time"
)

// Config 配置模型
type Config struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `json:"name" gorm:"not null;uniqueIndex" validate:"required"`
	Code      string    `json:"code" gorm:"not null;uniqueIndex" validate:"required"`
	Value     string    `json:"value" gorm:"not null" validate:"required"`
}

// TableName 表名
func (Config) TableName() string {
	return "configs"
}
