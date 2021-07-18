package cmd

import (
	"errors"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/devhg/ddocker/container"
)

// 需要单测

var RunCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit
			 ddocker run -it [command]`,
	Flags: []cli.Flag{
		// 交互式容器，重新分配终端
		cli.BoolFlag{
			Name:  "it",
			Usage: "enable tty",
		},
	},
	/*
		1. 判断参数是否包含command
		2. 获取用户指定的command
		3. 调用Run function去准备启动容器
	*/
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return errors.New("Missing container command")
		}
		cmd := ctx.Args().Get(0)
		tty := ctx.Bool("it")
		Run(tty, cmd)
		return nil
	},
}

// Run 这里是真正开始之前创建好的command调用，它首先会clone出来一个namespace隔离的
// 进程，然后在子进程中调用/proc/self/exe，也就是自己调用自己，发送init参数，
// 调用之前写的init方法，去初始化一些容器的参数，
func Run(tty bool, command string) {
	parentProcess := container.NewParentProcess(tty, command)
	if err := parentProcess.Start(); err != nil {
		logrus.Error(err)
	}
	parentProcess.Wait()
	os.Exit(-1)
}

var InitCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside.",
	Flags: nil,
	/*
		1. 获取参数
		2. 执行容器的初始化操作
	*/
	Action: func(ctx *cli.Context) error {
		logrus.Infof("init come on")
		cmd := ctx.Args().Get(0)
		err := container.RunContainerInitProcess(cmd, nil)
		return err
	},
}
