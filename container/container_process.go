package container

import (
	"github.com/sirupsen/logrus"
	"os"
	"syscall"
)

// RunContainerInitProcess 是在容器内部执行的，也就是说，
// 代码执行到这里后，容器所在的进程其实就已经创建完成了。
// 这是本容器执行的一个进程。
func RunContainerInitProcess(command string, args []string) error {
	logrus.Infof("command is %s", command)

	// logic
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "/proc", uintptr(defaultMountFlags), "")
	argv := []string{command}
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		logrus.Errorln(err.Error())
	}
	return nil
}
