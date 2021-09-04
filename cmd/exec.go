package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	// setns
	_ "github.com/devhg/ddocker/cmd/enterns"
	"github.com/devhg/ddocker/container"
)

const (
	ENV_EXEC_PID = "ddocker_pid"
	ENV_EXEC_CMD = "ddocker_cmd"
)

var ExecCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(ctx *cli.Context) error {

		if os.Getenv(ENV_EXEC_PID) != "" {
			logrus.Infof("pid callback pid %d", os.Getgid())
			return nil
		}

		if len(ctx.Args()) < 2 {
			return fmt.Errorf("expect shell command is: [ddocker exec containerID command]")
		}

		containerID := ctx.Args().Get(0)

		var commandArr []string
		commandArr = append(commandArr, ctx.Args().Tail()...)

		execConatiner(containerID, commandArr)

		return nil
	},
}

func execConatiner(contianerID string, cmds []string) {
	// 根据容器id 获取进程 pid
	cpid := getContainerPID(contianerID)
	if cpid == "" {
		return
	}

	command := strings.Join(cmds, " ")
	logrus.Infof("containerPID[%v] command[%v]", cpid, command)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Setenv(ENV_EXEC_PID, cpid)
	os.Setenv(ENV_EXEC_CMD, command)

	if err := cmd.Run(); err != nil {
		logrus.Errorf("exec container[%v] error[%v]", contianerID, err)
	}
}

func getContainerPID(contianerID string) string {
	config := path.Join(container.DefaultInfoLocation, contianerID)
	fileInfo, err := os.Stat(config)
	if err != nil || os.IsNotExist(err) {
		logrus.Errorf("func[getContainerPID] error[%v]", err)
		return ""
	}

	info, err := getContainerInfo(fileInfo)
	if err != nil {
		logrus.Errorf("func[getContainerPID] get container info error[%v]", err)
		return ""
	}
	return info.PID
}
