package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/devhg/ddocker/container"
)

var CommitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 2 {
			return errors.New("missing containerID or image name")
		}

		containerID := ctx.Args().Get(0)
		image := ctx.Args().Get(1)
		commitContainer(containerID, image)
		return nil
	},
}

func commitContainer(containerID, image string) {
	mntURL := fmt.Sprintf(container.MntURL, containerID)
	imageTar := path.Join(container.RootURL, image+".tar")
	logrus.Infof("new image:%v", imageTar)

	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		logrus.Errorf("tar folder %s error: %v", imageTar, err)
	}
}
