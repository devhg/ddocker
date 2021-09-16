package cmd

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/devhg/ddocker/cgroups"
	"github.com/devhg/ddocker/cgroups/subsystems"
	"github.com/devhg/ddocker/container"
	"github.com/devhg/ddocker/util"
)

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
		// deatch分离式容器，containerd容器统一管理
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
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
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "set environment",
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
		detach := ctx.Bool("d")
		if tty && detach {
			return errors.New("-it and -d parameter can not both provider")
		}

		resConf := &subsystems.ResourceConfig{
			MemoryLimit: ctx.String("mm"),
			CPUSet:      ctx.String("cpuset"),
			CPUShare:    ctx.String("cpushare"),
		}

		containerName := ctx.String("name")
		volumes := ctx.String("v")
		imageName := commands[0]
		commands = commands[1:]

		logrus.Infof("create tty[%v] name[%v]", tty, containerName)

		env := ctx.StringSlice("e")
		logrus.Infof("create env [%v]", env)

		run(tty, commands, resConf, containerName, volumes, imageName, env) // volume 临时放在这里
		return nil
	},
}

// run 这里是真正开始之前创建好的command调用，它首先会clone出来一个namespace隔离的
// 进程，然后在子进程中调用/proc/self/exe，也就是自己调用自己，发送init参数，
// 调用之前写的init方法，去初始化一些容器的参数，
func run(tty bool, commands []string, res *subsystems.ResourceConfig, name, volume, image string, env []string) {
	// 首先生成长度为10的容器id
	id := util.RandStringBytes(10)

	parentProcess, writePipe := container.NewParentProcess(tty, id, volume, image, env)
	if parentProcess == nil {
		logrus.Errorf("new parent process error")
		return
	}
	if err := parentProcess.Start(); err != nil {
		logrus.Error(err)
	}

	// 记录容器信息
	containerID, err := container.RecordContainerInfo(parentProcess.Process.Pid, commands, id, name, volume)
	if err != nil {
		logrus.Errorf("func[RecordContainerInfo] for %s error: %v", name, err)
		return
	}

	// 创建cgroupManager，并调用 Set 设置资源限制 和 Apply 在限制上生效
	cgroupManager := cgroups.NewCgroupManager("ddocker-cgroup")
	defer cgroupManager.Destroy()

	// 设置资源限制
	err = cgroupManager.Set(res)
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
	if tty {
		_ = parentProcess.Wait()
		container.DeleteContainerInfo(containerID)
		container.DeleteWorkSpace(containerID, volume)
	}
}
