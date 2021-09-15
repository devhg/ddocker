package container

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

// NewWorkSpace creates a workspace
func NewWorkSpace(containerID, volume, image string) {
	roLayer := createReadOnlyLayer(image)   // /root/${image}/
	wLayer := createWriteLayer(containerID) // /root/writeLayer/${containerID}/
	mnt := fmt.Sprintf(MntURL, containerID) // /root/mnt/${containerID}/

	if roLayer != "" && wLayer != "" {
		createMountPoint(roLayer, wLayer, mnt)
	}

	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			mountVolume(mnt, volumeURLs)
			return
		}
		logrus.Errorf("bad volume: %v", volume)
	}
}

func createReadOnlyLayer(image string) string {
	// ${RootURL}/${image}
	roLayer := path.Join(RootURL, image)         // ReadOnlyLayer location
	imageURL := path.Join(RootURL, image+".tar") // image file

	if exist, _ := pathExist(roLayer); exist {
		return roLayer
	}

	exist, err := pathExist(imageURL)
	if err != nil {
		logrus.Infof("failed to judge whether %v exists. %v", imageURL, err)
	}

	if exist {
		if err := os.MkdirAll(roLayer, 0622); err != nil {
			logrus.Errorf("mkdir %v error: %v", roLayer, err)
		}

		if _, err := exec.Command("tar", "-xvf", imageURL, "-C", roLayer).CombinedOutput(); err != nil {
			logrus.Errorf("unTar %v error: %v", imageURL, err)
		}
	}

	return roLayer
}

func createWriteLayer(containerID string) string {
	// "/root/writeLayer/${containerID}"
	wLayer := fmt.Sprintf(WriteLayerURL, containerID)

	if exist, _ := pathExist(wLayer); !exist {
		if err := os.MkdirAll(wLayer, 0777); err != nil {
			logrus.Errorf("mkdir %v error: %v", wLayer, err)
		}
	}

	// work是overlay必须的，具体为什么？？？？ 暂时放这里吧
	if exist, _ := pathExist(WorkDirURL); !exist {
		if err := os.MkdirAll(WorkDirURL, 0777); err != nil {
			logrus.Errorf("mkdir %v error: %v", WorkDirURL, err)
		}
	}
	return wLayer
}

func createMountPoint(roLayer, wLayer, mnt string) {
	if exist, _ := pathExist(mnt); !exist {
		if err := os.MkdirAll(mnt, 0777); err != nil {
			logrus.Errorf("mkdir %v error: %v", mnt, err)
		}
	}

	// mount: unknown filesystem type 'aufs' aufs已经过时了，改成overlay
	// cat /proc/filesystems 查看支持的文件系统类型
	//
	// mount -t overlay overlay -o lowerdir=./lower,upperdir=./upper,workdir=./work ./merged
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", roLayer, wLayer, WorkDirURL)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mnt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("func[CreateMountPoint] %v", err)
	}
}

func pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// mountVolume 对用户要挂载进来的路径进行挂载
func mountVolume(mnt string, volumeURLs []string) {
	// 创建宿主机文件目录
	parentURL := volumeURLs[0]
	if err := os.MkdirAll(parentURL, 0777); err != nil {
		logrus.Errorf("mkdir parent dir %v error: %v", parentURL, err)
	}

	// 在容器目录创建挂载点目录
	containerURL := path.Join(mnt, volumeURLs[1])
	if err := os.MkdirAll(containerURL, 0777); err != nil {
		logrus.Errorf("mkdir container dir %v error: %v", containerURL, err)
	}

	// 把宿主机文件目录挂在到容器内挂载点
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", containerURL, parentURL, WorkDirURL)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, containerURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("func[mountVolume] %v", err)
	}
}

// DeleteWorkSpace .
func DeleteWorkSpace(containerID, volume string) {
	wLayer := fmt.Sprintf(WriteLayerURL, containerID) // /root/writeLayer/${containerID}/
	mnt := fmt.Sprintf(MntURL, containerID)           // /root/mnt/${containerID}/

	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			deleteMountPointWithVolume(mnt, volumeURLs)
			deleteWritePlayer(wLayer)
			return
		}
	}

	deleteMountPoint(mnt)
	deleteWritePlayer(wLayer)
}

func deleteMountPoint(mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logrus.Errorf("umount %v error: %v", mntURL, err)
	}

	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Errorf("remove dir %v error: %v", mntURL, err)
	}
}

func deleteMountPointWithVolume(mnt string, volumeURLs []string) {
	// 卸载容器里面volome挂载点的文件系统
	containerURL := path.Join(mnt, volumeURLs[1])

	cmd := exec.Command("umount", containerURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logrus.Errorf("umount volume %v:%v error: %v", volumeURLs[0], containerURL, err)
	}

	// 卸载整个容器文件系统的挂载点
	deleteMountPoint(mnt)
}

func deleteWritePlayer(wLayer string) {
	if err := os.RemoveAll(wLayer); err != nil {
		logrus.Errorf("remove dir %v error: %v", wLayer, err)
	}

	// work是overlay必须的，具体为什么？？？？
	// workURL := rootURL + "work/"
	// if err := os.RemoveAll(workURL); err != nil {
	// 	logrus.Errorf("remove dir %v error: %v", workURL, err)
	// }
}
