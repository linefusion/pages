package commands

import (
	"context"
	"os"
	"os/signal"

	"github.com/linefusion/pages/pkg/common/app"
	"github.com/linefusion/pages/pkg/pages/config"
	"github.com/linefusion/pages/pkg/pages/server"
	"github.com/urfave/cli/v2"
	"github.com/zclconf/go-cty/cty"
)

var cfg config.Config

func init() {
	app.AddCommand(
		&cli.Command{
			Name:  "start",
			Usage: "Starts the server using a config file",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "config",
					Aliases:     []string{"c"},
					Usage:       "load configuration from file",
					DefaultText: "Pagesfile",
				},
			},
			Before: func(context *cli.Context) error {
				f := context.String("config")
				if f == "" {
					f = "Pagesfile"
				}

				c, err := config.LoadFile(f, map[string]cty.Value{})
				if err != nil {
					return err
				}

				cfg = c
				return nil
			},
			Action: commandStart,
		},
	)
}

func commandStart(c *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	servers := []server.Server{}
	for _, serverConfig := range cfg.Servers {
		servers = append(servers, server.New(ctx, serverConfig))
	}

	go func() {
		<-ch
		cancel()
		for _, server := range servers {
			server.Stop()
		}
	}()

	for _, server := range servers {
		server.Start()
	}

	for _, server := range servers {
		server.Wait()
	}

	return nil
}
