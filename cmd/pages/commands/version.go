package commands

import (
	"fmt"
	"runtime/debug"

	"github.com/linefusion/pages/pkg/common/app"
	"github.com/linefusion/pages/pkg/common/version"
	"github.com/urfave/cli/v2"
)

func init() {
	app.AddCommand(
		&cli.Command{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "Shows CLI version information",
			Action:  commandVersion,
		},
	)
}

func getVersion() string {
	result := version.GetVersion()
	if commit := version.GetCommit(); commit != "" {
		result = fmt.Sprintf("%s\n  commit: %s", result, commit)
	}
	if date := version.GetDate(); date != "" {
		result = fmt.Sprintf("%s\n  built at: %s", result, date)
	}
	if builtBy := version.GetBuiltBy(); builtBy != "" {
		result = fmt.Sprintf("%s\n  built by: %s", result, builtBy)
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf("%s\n  module version: %s, checksum: %s", result, info.Main.Version, info.Main.Sum)
	}
	return result
}

func commandVersion(c *cli.Context) error {
	fmt.Printf("Pages version: %s", getVersion())
	return nil
}
