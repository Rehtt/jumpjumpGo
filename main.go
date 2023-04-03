package main

import (
	"context"
	"github.com/Rehtt/Kit/i18n"
	"github.com/Xuanwo/go-locale"
	"jumpjumpGo/cmd"
	"jumpjumpGo/conf"
	"jumpjumpGo/database"
	"jumpjumpGo/server"
	"sync"
)

var (
	mainVersion  string
	buildVersion string
)

func main() {
	lang, err := locale.Detect()
	if err == nil {
		i18n.SetLang(&lang)
	}
	db, err := database.OpenDB("db")
	if err != nil {
		panic(err)
	}
	conf.Conf.DB = db
	conf.Conf.BuildVersion = buildVersion
	conf.Conf.MainVersion = mainVersion

	ctx, ch := context.WithCancelCause(context.Background())
	var start sync.WaitGroup
	start.Add(1)
	go server.StartSSH(ctx, ":2220", &start)
	start.Wait()

	cmd.StartLocalCMD(ctx, ch)
}
