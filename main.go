package main

import (
	"context"
	"flag"
	"github.com/Rehtt/Kit/i18n"
	"github.com/Rehtt/jumpjumpGo/cmd"
	"github.com/Rehtt/jumpjumpGo/conf"
	"github.com/Rehtt/jumpjumpGo/database"
	"github.com/Rehtt/jumpjumpGo/server"
	"github.com/Xuanwo/go-locale"
	"sync"
)

var (
	mainVersion  string
	buildVersion string
	addr         = flag.String("addr", ":2220", i18n.GetText("listening address"))
)

func main() {
	flag.Parse()
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
	go server.StartSSH(ctx, *addr, &start)
	start.Wait()

	cmd.StartLocalCMD(ctx, ch)
}
