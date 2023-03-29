package conf

import (
	"jumpjump/database"
)

type Config struct {
	DB           *database.DB
	BuildVersion string
	MainVersion  string
}

var Conf = new(Config)
