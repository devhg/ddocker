package container

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// ContainerInfo .
type ContainerInfo struct {
	PID         string   `json:"pid"`         // 容器的 init 进程在宿主机上的 PID
	ID          string   `json:"id"`          // 容器 ID
	Name        string   `json:"name"`        // 容器名
	Command     string   `json:"command"`     // 容器内 init 运行命令
	CreatedTime string   `json:"create_time"` // 创建时间
	Status      string   `json:"status"`      // 容器的状态
	Volume      string   `json:"volume"`      // 容器的数据卷
	PortMapping []string `json:"portmapping"` // 容器的端口映射
}

const (
	StatusRunning string = "running"
	StatusStopped string = "stopped"
	StatusExit    string = "exit"
)

const (
	DefaultInfoLocation string = "/var/run/ddocker/"
	ConfigName          string = "config.json"
	StdLogFileName      string = "std.log"
)

// NewParentProcess 这里是父进程（当前进程执行的内容）
// 1. 在/proc/self/exe的调用中, /proc/self/指的就是当前进程自己的环境,
// exec其实就是自己调用了自己。init和command参数是传递给本进程的。
// 2. 后面args是参数，其中init是传给本进程的第一个参数。简而言之，
// 先调用init, 即调用initCommand去执行一些环境和资源的初始化操作。
//
// 3. 下面指定了一些clone参数去fork新进程，并使用namespace隔离新创建的进程和外部环境。
// 4. 如果用指定了-it参数，就需要把进程的输入输出导入到标准的输入输出
func NewParentProcess(tty bool, cid, volume, image string, envs []string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		logrus.Errorf("New pipe error: %v", err)
	}

	cmd := exec.Command("/proc/self/exe", "init")
	logrus.Info(cmd.Args)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// /var/run/ddocker/${containerID}/std.log
		stdLogFile := RedirectContainerLog(cid)
		if stdLogFile == nil {
			return nil, nil
		}
		cmd.Stdout = stdLogFile
	}

	cmd.ExtraFiles = []*os.File{readPipe}   // 传入管道读取端的句柄
	cmd.Env = append(os.Environ(), envs...) // 继承父进程的环境变量

	NewWorkSpace(cid, volume, image)
	cmd.Dir = fmt.Sprintf(MntURL, cid)
	return cmd, writePipe
}

// NewPipe .
func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

// RecordContainerInfo
func RecordContainerInfo(cpid int, commandArr []string, id, name, volume string) (string, error) {
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArr, " ")

	if name == "" {
		name = id
	}

	info := &ContainerInfo{
		ID:          id,
		PID:         strconv.Itoa(cpid),
		Name:        name,
		Command:     command,
		CreatedTime: createTime,
		Status:      StatusRunning,
		Volume:      volume,
	}

	b, err := json.Marshal(info)
	if err != nil {
		return "", fmt.Errorf("marshal container info error[%v]", err)
	}

	// /var/run/ddocker/${containerID}/
	folder := path.Join(DefaultInfoLocation, id)
	if err := os.MkdirAll(folder, 0622); err != nil {
		return "", err
	}

	// /var/run/ddocker/${containerID}/config.json
	dstFile := path.Join(folder, ConfigName)
	f, err := os.Create(dstFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.WriteString(string(b)); err != nil {
		return "", fmt.Errorf("write container info error[%v]", err)
	}

	return id, nil
}

// DeleteContainerInfo
func DeleteContainerInfo(containerID string) {
	// /var/run/ddocker/${containerID}/
	folder := path.Join(DefaultInfoLocation, containerID)
	if err := os.RemoveAll(folder); err != nil {
		logrus.Errorf("func[DeleteContainerInfo] error: %v", err)
	}
}

// RedirectContainerLog
func RedirectContainerLog(containerID string) *os.File {
	// /var/run/ddocker/${containerID}
	dir := path.Join(DefaultInfoLocation, containerID)
	if err := os.MkdirAll(dir, 0622); err != nil {
		logrus.Errorf("func[RedirectContainerLog] error[%v]", err)
		return nil
	}

	// /var/run/ddocker/${containerID}/std.log
	stdLogFilePath := path.Join(dir, StdLogFileName)
	stdLogFile, err := os.Create(stdLogFilePath)
	if err != nil {
		logrus.Errorf("func[RedirectContainerLog] error[%v]", err)
		return nil
	}

	return stdLogFile
}
