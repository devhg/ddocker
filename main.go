package main

import (
	"github.com/devhg/ddocker/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
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

/*
// 挂载了memory subsystem 的 hierarchy 的根目录
const (
	cgroupMemoryHierarchMount = "/sys/fs/cgroup/memory"
	memoryLimitInBytes        = "memory.limit_in_bytes"
	memoryTasks               = "tasks"
)

func main() {
	if os.Args[0] == "/proc/self/exe" {
		//容器进程
		fmt.Printf("current pid %d\n", syscall.Getpid())
		cmd := exec.Command("sh", "-c", `stress --vm-bytes 200m --vm-keep -m 1`)
		cmd.SysProcAttr = &syscall.SysProcAttr{}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// 用来指定fork出来新进程内的初始命令
	cmd := exec.Command("/proc/self/exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: 0, Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: 0, Size: 1},
		},
	}
	//cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(1.txt), Gid: uint32(1.txt)}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	} else {
		// 得到fork出来进程映射在外部命名空间的pid
		fmt.Printf("宿主机空间的pid %v\n", cmd.Process.Pid)

		// 1.在系统默认创建挂载了memory subsystem 的Hierarchy上创建cgroup
		os.Mkdir(path.Join(cgroupMemoryHierarchMount, "testMemorylimit"), 0755)

		// 2.将容器进程加入到这个cgroup
		containerPid := strconv.Itoa(cmd.Process.Pid)
		ioutil.WriteFile(path.Join(cgroupMemoryHierarchMount, "testMemorylimit", memoryTasks),
			[]byte(containerPid), 0644)

		// 3.限制cgroup进程使用
		ioutil.WriteFile(path.Join(cgroupMemoryHierarchMount, "testMemorylimit", memoryLimitInBytes),
			[]byte("100m"), 0644)
	}
	cmd.Process.Wait()
}*/
