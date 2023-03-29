package database

import "gorm.io/gorm"

type Server struct {
	gorm.Model
	Ip   string
	Port int
}

func (Server) TableName() string {
	return "server"
}
