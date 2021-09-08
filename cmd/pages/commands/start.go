package commands

import (
	"os"
	"os/signal"
	"sync"

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
	var wait sync.WaitGroup
	wait.Add(1)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		<-interrupt
		wait.Done()
	}()

	servers := []server.Server{}
	for _, serverConfig := range cfg.Servers {
		srv := server.New(serverConfig)
		servers = append(servers, srv)
		srv.Start()
	}

	// Wait for interruption/exit
	wait.Wait()
	close(interrupt)

	// Close each server in parallel
	wait.Add(len(servers))
	for _, srv := range servers {
		go func(srv server.Server) {
			srv.Stop()
			wait.Done()
		}(srv)
	}

	// Wait for all servers to shutdown
	wait.Wait()
	return nil
}
