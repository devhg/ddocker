package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/devhg/ddocker/container"
)

var PsCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the container",
	Action: func(ctx *cli.Context) error {
		ListContainers()
		return nil
	},
}

func ListContainers() {
	dir := container.DefaultInfoLocation
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		logrus.Errorf("read dir %s error[%v]", dir, err)
	}

	var infos []*container.ContainerInfo
	for _, file := range files {
		info, err := readContainerInfo(file)
		if err != nil {
			logrus.Errorf("get container info error[%v]", err)
			continue
		}
		infos = append(infos, info)
	}

	// 控制台打印对齐的表格
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tPID\tNAME\tSTATUS\tCOMMAND\tCREATE\n")
	for _, info := range infos {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			info.ID,
			info.PID,
			info.Name,
			info.Status,
			info.Command,
			info.CreatedTime,
		)
	}

	// 刷新输出流缓冲区
	if err := w.Flush(); err != nil {
		logrus.Errorf("tabwriter flush error[%v]", err)
	}
}
