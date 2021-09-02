package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/devhg/ddocker/container"
)

var LogCommand = cli.Command{
	Name:  "log",
	Usage: "print logs of a container",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("please input your container id :)")
		}

		containerID := ctx.Args().Get(0)
		logContainer(containerID)
		return nil
	},
}

func logContainer(contianerID string) {
	stdLogFile := path.Join(container.DefaultInfoLocation, contianerID, container.StdLogFileName)
	bytes, err := os.ReadFile(stdLogFile)
	if err != nil {
		logrus.Errorf("container log open file %v error", err)
		return
	}

	fmt.Fprint(os.Stdout, string(bytes))
}
