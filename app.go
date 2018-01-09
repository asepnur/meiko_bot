package main

import (
	"flag"
	"log"
	"runtime"

	"github.com/asepnur/meiko_bot/src/util/alias"
	"github.com/asepnur/meiko_bot/src/util/auth"
	"github.com/asepnur/meiko_bot/src/util/conn"
	"github.com/asepnur/meiko_bot/src/util/env"
	"github.com/asepnur/meiko_bot/src/util/jsonconfig"
	"github.com/asepnur/meiko_bot/src/webserver"
	"github.com/asepnur/meiko_bot/src/webserver/handler/bot"
)

type configuration struct {
	Directory alias.DirectoryConfig `json:"directory"`
	Database  conn.DatabaseConfig   `json:"database"`
	Webserver webserver.Config      `json:"webserver"`
	Auth      auth.Config           `json:"auth"`
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	flag.Parse()

	// load configuration
	cfgenv := env.Get()
	config := &configuration{}
	isLoaded := jsonconfig.Load(&config, "/etc/meiko", cfgenv) || jsonconfig.Load(&config, "./files/etc/meiko", cfgenv)
	if !isLoaded {
		log.Fatal("Failed to load configuration")
	}

	// initiate instance
	alias.InitDirectory(config.Directory)
	conn.InitDB(config.Database)
	bot.Init()
	auth.Init(config.Auth)
	webserver.Start(config.Webserver)
}
