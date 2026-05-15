package main

import (
	"fmt"
	"time"

	"log"
	"os"

	"github.com/urfave/cli"

	"tgbot/internal/cmd"
	"tgbot/internal/conf"
)

const (
	Version = "0.1"
	AppName = "tgbot"
	CodeDev = false
)

func init() {
	conf.App.Version = Version
	if CodeDev {
		conf.App.Version = fmt.Sprintf("%s%d", Version, time.Now().Unix())
	}
	conf.App.Name = AppName
}

func main() {
	app := cli.NewApp()
	app.Name = AppName
	app.Version = Version
	app.Usage = "tgbot service"
	app.Commands = []cli.Command{
		cmd.Web,
		cmd.Root,
		cmd.Install,
		cmd.Uninstall,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
