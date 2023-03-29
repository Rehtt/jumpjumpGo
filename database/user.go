package database

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Username   string         `gorm:"column:username"`
	Password   string         `gorm:"column:password"`
	PublicKeys JSON[[]string] `gorm:"column:public_keys;type:text"`
	LastLogin  *time.Time
}

func (User) TableName() string {
	return "user"
}
