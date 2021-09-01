package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/linefusion/pages/pkg/common/app"
	v "github.com/linefusion/pages/pkg/common/version"

	_ "github.com/linefusion/pages/internal/pages/commands"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

func init() {
	v.Initialize(version, commit, date, builtBy)
}

func main() {
	app := &cli.App{
		Name:        "Pages",
		Usage:       "Linefusion Pages CLI",
		Version:     v.GetVersion(),
		Description: fmt.Sprintf("Linefusion Pages CLI %s", version),
		Flags:       app.GetFlags(),
		Commands:    app.GetCommands(),
		Before: func(c *cli.Context) error {
			for _, init := range app.GetInitializers() {
				err := init(c)
				if err != nil {
					log.Fatal(err)
				}
			}
			return nil
		},
		After: func(c *cli.Context) error {
			for _, finalize := range app.GetFinalizers() {
				err := finalize(c)
				if err != nil {
					log.Fatal(err)
				}
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			cli.ShowAppHelpAndExit(c, 1)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
