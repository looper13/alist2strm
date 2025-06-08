package model

import (
	"alist2strm/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	var err error
	DB, err = gorm.Open(sqlite.Open(config.GlobalConfig.Database.Path), &gorm.Config{})
	if err != nil {
		return err
	}

	// 自动迁移
	err = DB.AutoMigrate(
		&User{},
		&Config{},
		&Task{},
		&TaskLog{},
		&FileHistory{},
		&NotificationQueue{},
		&ValidationTask{},
		&SystemLog{},
	)
	if err != nil {
		return err
	}

	return nil
}
