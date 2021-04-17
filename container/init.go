package container

import (
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess 是父进程（当前进程执行的内容）。
// 子进程在/proc/self/exe的调用中，/proc/self/指的就是当前进程自己的环境。
// exe其实就是自己调用了自己。init和command参数是传递给本进程的。
// 简而言之，就是会去调用initCommand去执行一些环境和资源的初始化操作。
//
// 下面指定了一些clone参数就是去fork一个新进程，并且使用namespace隔离
// 新创建的进程和外部环境。
// 如果用指定了-it参数，就需要把进程的输入输出导入到标准的输入输出
func NewParentProcess(tty bool, command string) *exec.Cmd {
	args := []string{"init", command}
	cmd := exec.Command("/proc/self/exe", args...) // ?? 不太明白
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: 0, Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: 0, Size: 1},
		},
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}
