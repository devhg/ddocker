package cmd

import (
	"errors"
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var CommitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return errors.New("missing container name")
		}

		imageName := ctx.Args().Get(0)
		commitContainer(imageName)
		return nil
	},
}

func commitContainer(image string) {
	mntURL := "/root/mnt"
	imageTar := "/root/" + image + ".tar"
	logrus.Infof("new image:%v", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		logrus.Errorf("tar folder %s error: %v", imageTar, err)
	}
}
