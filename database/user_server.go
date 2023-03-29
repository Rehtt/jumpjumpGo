package database

import "gorm.io/gorm"

type UserServer struct {
	gorm.Model
	UserID        uint
	ServerID      uint
	Server        Server
	LoginUserName string
	LoginPassword *string
	PrivateKey    *string
}

func (UserServer) TableName() string {
	return "user_server"
}
