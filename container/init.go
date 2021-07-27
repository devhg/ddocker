package container

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

// NewParentProcess 这里是父进程（当前进程执行的内容）
// 1. 在/proc/self/exe的调用中, /proc/self/指的就是当前进程自己的环境,
// exec其实就是自己调用了自己。init和command参数是传递给本进程的。
// 2. 后面args是参数，其中init是传给本进程的第一个参数。简而言之，
// 先调用init, 即调用initCommand去执行一些环境和资源的初始化操作。
//
// 3. 下面指定了一些clone参数去fork新进程，并使用namespace隔离新创建的进程和外部环境。
// 4. 如果用指定了-it参数，就需要把进程的输入输出导入到标准的输入输出
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		logrus.Errorf("New pipe error: %v", err)
	}
	cmd := exec.Command("/proc/self/exe", "init")
	logrus.Info(cmd.Args)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe} // 传入管道读取端的句柄
	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}
