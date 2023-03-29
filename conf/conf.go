package conf

import (
	"jumpjumpGo/database"
)

type Config struct {
	DB           *database.DB
	BuildVersion string
	MainVersion  string
}

var Conf = new(Config)
