package cmd

import "github.com/urfave/cli"

var PsCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the container",
	Action: func(ctx *cli.Context) error {
		ListContainers()
		return nil
	},
}

func ListContainers() {

}
