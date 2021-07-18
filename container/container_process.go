package container

import (
	"os"
	"syscall"

	"github.com/sirupsen/logrus"
)

// RunContainerInitProcess 是在容器内部执行的，也就是说代码执行到这里后，
// 容器所在的进程其实就已经创建完成了，这是本容器执行的第一个进程。
// 使用mount先去挂在proc文件系统，以后后面通过ps等系统命令去查看进程的资源占用情况
//
// syscall.MS_NOEXEC 本文件系统中不允许运行其他程序
// syscall.MS_NOSUID 本文件系统运行程序，禁止set-user-ID或set-group-ID
// syscall.MS_NODEV  所有mount的系统都会默认设定的参数
//
// syscall.Exec()最终调用了kernel的int execve(const char* filename, char* const argv[], char* const emvp[])
// 这个系统调用完成了初始化动作并将用户进程运行起来的动作，
// 在前面的代码中，容器的第一个进程是init初始化的进程，我们希望是我们自己的进程。但是PID=1的进程是不能kill的，
// 这个系统调用的作用就是，将原来的init进程替换成用户自己的进程，这样当进入容器的时候，PID=1的程序就是我们指定的进程了。
// 容器 === 进程。这其实也是目前docker使用容器引擎runC的实现方式之一
func RunContainerInitProcess(command string, args []string) error {
	logrus.Infof("command is %s", command)

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	argv := []string{}
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		logrus.Errorln(err.Error())
	}
	return nil
}
