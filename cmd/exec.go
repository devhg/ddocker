package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	// setns
	_ "github.com/devhg/ddocker/cmd/enterns"
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
	cpid := GetContainerPID(contianerID)
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

	envs := getEnvByPID(cpid)
	cmd.Env = append(os.Environ(), envs...)

	if err := cmd.Run(); err != nil {
		logrus.Errorf("exec container[%v] error[%v]", contianerID, err)
	}
}

func getEnvByPID(pid string) []string {
	// 进程环境变量存放的位置是 /proc/PID/environ
	path := fmt.Sprintf("/proc/%s/environ", pid)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.Errorf("read file %s error %v", path, err)
		return nil
	}

	// 多个环境变量的分隔符是 \u0000
	return strings.Split(string(b), "\u0000")
}
