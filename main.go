package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/devhg/ddocker/cmd"
)

const usage = `ddocker is a simple container runtime implementation.
			   The purpose of this project is to learn how docker works and how to write a docker by ourselves.
			   Enjoy it, just for fun (:`

func main() {
	app := cli.NewApp()
	app.Name = "ddocker"
	app.Usage = usage

	app.Commands = []cli.Command{
		cmd.InitCommand,
		cmd.RunCommand,
		cmd.CommitCommand,
		cmd.PsCommand,
		cmd.LogCommand,
	}

	app.Before = func(ctx *cli.Context) error {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
