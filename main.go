package main

import (
	"jumpjumpGo/conf"
	"jumpjumpGo/database"
	"jumpjumpGo/server"
)

var (
	mainVersion  string
	buildVersion string
)

func main() {
	db, err := database.OpenDB("db")
	if err != nil {
		panic(err)
	}
	conf.Conf.DB = db
	conf.Conf.BuildVersion = buildVersion
	conf.Conf.MainVersion = mainVersion
	server.StartSSH(":2220")
}
