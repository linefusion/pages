package commands

import (
	"github.com/linefusion/pages/pkg/common/app"
	"github.com/linefusion/pages/pkg/pages/config"
	"github.com/urfave/cli/v2"
	"github.com/zclconf/go-cty/cty"
)

func init() {
	app.AddCommand(
		&cli.Command{
			Name:  "serve",
			Usage: "Serves a local (current working) directory with default configs",
			Before: func(context *cli.Context) error {
				c, err := config.LoadString(`
          server "default" {
            listen {
              bind = "${lookup(env, "BIND", "0.0.0.0")}"
              port = "${lookup(env, "PORT", "80")}"
            }
            pages {
              page "default" {
                source "local" {}
              }
            }
          }
        `, map[string]cty.Value{})
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
