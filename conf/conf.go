package conf

import (
	"github.com/Rehtt/jumpjumpGo/database"
)

type Config struct {
	DB               *database.DB
	BuildVersion     string
	MainVersion      string
	SSHServerVersion string
	SSHClientVersion string
}

var Conf = new(Config)

func init() {
	Conf.SSHServerVersion = "SSH-2.0-jumpjumpGo"
	Conf.SSHClientVersion = "SSH-2.0-jumpjumpGo"
}
