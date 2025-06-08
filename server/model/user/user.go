package user

import (
	"time"
)

// User 用户模型
type User struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Username    string     `json:"username" gorm:"not null;uniqueIndex" validate:"required"`
	Password    string     `json:"-" gorm:"not null" validate:"required"`
	Nickname    string     `json:"nickname"`
	Status      string     `json:"status" gorm:"not null;default:active"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}
