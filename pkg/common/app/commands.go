package app

import "github.com/urfave/cli/v2"

var (
	commands []*cli.Command = []*cli.Command{}
)

func GetCommands() []*cli.Command {
	return commands
}

func AddCommand(command *cli.Command) {
	commands = append(commands, command)
}
