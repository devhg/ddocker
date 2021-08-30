package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/devhg/ddocker/cgroups"
	"github.com/devhg/ddocker/cgroups/subsystems"
	"github.com/devhg/ddocker/container"
)

// 需要单测

// RunCommand .
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
		cli.StringFlag{
			Name:  "mm",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
	},
	/*
		1. 判断参数是否包含command
		2. 获取用户指定的command
		3. 调用Run function去准备启动容器
	*/
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return errors.New("missing container command")
		}

		var commands []string
		for _, arg := range ctx.Args() {
			fmt.Println(arg)
			commands = append(commands, arg)
		}

		tty := ctx.Bool("it")
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: ctx.String("mm"),
			CPUSet:      ctx.String("cpuset"),
			CPUShare:    ctx.String("cpushare"),
		}
		fmt.Println(tty, resConf.CPUSet, resConf.CPUShare, resConf.MemoryLimit)
		Run(tty, commands, resConf, ctx.String("v")) // volume 临时放在这里
		return nil
	},
}

// Run 这里是真正开始之前创建好的command调用，它首先会clone出来一个namespace隔离的
// 进程，然后在子进程中调用/proc/self/exe，也就是自己调用自己，发送init参数，
// 调用之前写的init方法，去初始化一些容器的参数，
func Run(tty bool, commands []string, res *subsystems.ResourceConfig, volume string) {
	parentProcess, writePipe := container.NewParentProcess(tty, volume)
	if err := parentProcess.Start(); err != nil {
		logrus.Error(err)
	}

	// 创建cgroupManager，并调用 Set 设置资源限制 和 Apply 在限制上生效
	cgroupManager := cgroups.NewCgroupManager("ddocker-cgroup")
	defer cgroupManager.Destroy()

	// 设置资源限制
	err := cgroupManager.Set(res)
	if err != nil {
		panic(err)
	}
	// 将容器进程加入到各个subsystem挂载对应的cgroup中
	_ = cgroupManager.Apply(parentProcess.Process.Pid)
	if err != nil {
		panic(err)
	}
	// 初始化容器
	sendInitCommand(commands, writePipe)

	_ = parentProcess.Wait()

	mntURL := "/root/mnt/"
	rootURL := "/root/"
	container.DeleteWorkSpace(rootURL, mntURL, volume)
	// os.Exit(-1)
}

// InitCommand .
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
		err := container.RunContainerInitProcess()
		return err
	},
}

func sendInitCommand(commands []string, writePipe *os.File) {
	command := strings.Join(commands, " ")
	logrus.Infof("command all is %s", command)
	_, _ = writePipe.WriteString(command)
	writePipe.Close()
}
