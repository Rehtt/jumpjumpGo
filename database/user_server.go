package database

import "gorm.io/gorm"

type UserServer struct {
	gorm.Model
	UserID        uint
	LoginUserName string
	LoginPassword *string
	PrivateKey    *string
	ServerAddr    string
	ServerPort    string
	Alias         string
}

func (UserServer) TableName() string {
	return "user_server"
}
