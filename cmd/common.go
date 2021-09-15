package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"

	"github.com/devhg/ddocker/container"
)

func GetContainerPID(contianerID string) string {
	info := GetContainerInfo(contianerID)
	if info == nil {
		return ""
	}
	return info.PID
}

func readContainerInfo(f os.FileInfo) (*container.ContainerInfo, error) {
	containerID := f.Name()
	configFile := path.Join(container.DefaultInfoLocation, containerID, container.ConfigName)

	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var info container.ContainerInfo
	if err = json.Unmarshal(bytes, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func GetContainerInfo(contianerID string) *container.ContainerInfo {
	config := path.Join(container.DefaultInfoLocation, contianerID)
	fileInfo, err := os.Stat(config)
	if err != nil || os.IsNotExist(err) {
		logrus.Errorf("func[getContainerPID] error[%v]", err)
		return nil
	}

	info, err := readContainerInfo(fileInfo)
	if err != nil {
		logrus.Errorf("func[getContainerPID] get container info error[%v]", err)
		return nil
	}
	return info
}

func writeContainerInfo(contianerID string, info *container.ContainerInfo) error {
	// /var/run/ddocker/${containerID}/config.json
	dstFile := path.Join(container.DefaultInfoLocation, contianerID, container.ConfigName)

	content, err := json.Marshal(info)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(dstFile, content, 0622); err != nil {
		return err
	}
	return nil
}

func removeContainerInfo(contianerID string) {
	// /var/run/ddocker/${containerID}/
	dir := path.Join(container.DefaultInfoLocation, contianerID)
	_, err := os.Stat(dir)
	if err != nil || os.IsNotExist(err) {
		logrus.Errorf("func[removeContainerInfo] error[%v]", err)
		return
	}

	if err := os.RemoveAll(dir); err != nil {
		logrus.Errorf("func[removeContainerInfo] error[%v]", err)
		return
	}
}
