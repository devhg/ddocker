package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/devhg/ddocker/container"
)

var StopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return errors.New("missing containerID")
		}

		containerID := ctx.Args().Get(0)
		return stopContainer(containerID)
	},
}

func stopContainer(containerID string) error {
	cinfo := GetContainerInfo(containerID)

	// 根据容器id 获取进程 pid
	cpid := cinfo.PID
	if cpid == "" {
		return fmt.Errorf("canot find containerID[%v]'s PID", containerID)
	}

	pid, err := strconv.Atoi(cpid)
	if err != nil {
		return err
	}

	// 系统调用kill发送信号给容器进程，通过传递syscall.SIGTERM信号，杀掉容器主进程
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		logrus.Errorf("stop container[%v] error[%v]", containerID, err)
		return err
	}

	// 修改容器状态
	cinfo.Status = container.Stop
	cinfo.PID = ""

	if err := writeContainerInfo(containerID, cinfo); err != nil {
		logrus.Errorf("rewrite container info error[%v]", err)
		return err
	}

	return nil
}
