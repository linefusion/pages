package app

import "github.com/urfave/cli/v2"

var (
	flags []cli.Flag = []cli.Flag{}
)

func GetFlags() []cli.Flag {
	return flags
}

func AddFlag(flag cli.Flag) {
	flags = append(flags, flag)
}
