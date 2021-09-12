package cmd

import (
	"errors"
	"fmt"

	"github.com/devhg/ddocker/container"
	"github.com/urfave/cli"
)

var RemoveCommand = cli.Command{
	Name:  "rm",
	Usage: "remove a container",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return errors.New("missing containerID")
		}

		containerID := ctx.Args().Get(0)
		return removeContainer(containerID)
	},
}

func removeContainer(containerID string) error {
	cinfo := GetContainerInfo(containerID)
	if cinfo == nil {
		return fmt.Errorf("container[%v] not found", containerID)
	}

	if cinfo.Status != container.StatusStopped {
		return fmt.Errorf("canot remove a %v container", cinfo.Status)
	}

	removeContainerInfo(containerID)
	return nil
}
