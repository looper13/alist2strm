package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Username    string    `gorm:"unique;not null" json:"username"`
	Password    string    `gorm:"not null" json:"-"`
	Nickname    string    `json:"nickname"`
	Status      string    `gorm:"default:active" json:"status"`
	LastLoginAt time.Time `json:"lastLoginAt"`
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
