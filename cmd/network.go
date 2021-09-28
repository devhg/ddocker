package cmd

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/devhg/ddocker/network"
)

var NetworkCommand = cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Subcommands: []cli.Command{
		createCommand,
		listCommand,
		removeCommand,
	},
}

var createCommand = cli.Command{
	Name:  "create",
	Usage: "create a container network",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "driver",
			Usage: "network driver",
		},
		cli.StringFlag{
			Name:  "subnet",
			Usage: "subnet cidr",
		},
	},
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("missing network name")
		}

		if err := network.Init(); err != nil {
			return err
		}

		err := network.CreateNetwork(ctx.String("driver"), ctx.String("subnet"), ctx.Args()[0])
		if err != nil {
			return fmt.Errorf("create network error: %+v", err)
		}

		fmt.Printf("create network %s with dirver=%s subnet=%s success\n", ctx.Args()[0],
			ctx.String("driver"), ctx.String("subnet"))
		return nil
	},
}

var listCommand = cli.Command{
	Name:  "list",
	Usage: "list a container network",
	Action: func(ctx *cli.Context) error {
		if err := network.Init(); err != nil {
			return err
		}

		network.ListNetwork()

		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "remove",
	Usage: "remove container network",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("missing network name")
		}

		if err := network.Init(); err != nil {
			return err
		}

		err := network.DeleteNetwork(ctx.Args()[0])
		if err != nil {
			return fmt.Errorf("remove network error: %+v", err)
		}
		fmt.Println("delete network", ctx.Args()[0], "success")
		return nil
	},
}
