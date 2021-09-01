package commands

import (
	"github.com/linefusion/pages/pkg/common/app"
	"github.com/urfave/cli/v2"
)

func init() {
	app.AddCommand(
		&cli.Command{
			Name:   "init",
			Usage:  "Initializes a project",
			Action: commandInit,
		},
	)
}

func commandInit(c *cli.Context) error {
	return nil
}
