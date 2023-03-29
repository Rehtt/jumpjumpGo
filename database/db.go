package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func OpenDB(path string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(path))
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(
		&User{},
		&Server{},
		&UserServer{},
	)
	return &DB{db}, nil
}
