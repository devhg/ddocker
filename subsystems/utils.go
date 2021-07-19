package subsystems

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

// GetCgroupPath 得到cgroup在文件系统的中的绝对路径
func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountPoint(subsystem)
	absCgroupPath := path.Join(cgroupRoot, cgroupPath)

	_, err := os.Stat(absCgroupPath)
	if err == nil {
		return absCgroupPath, nil
	}

	if autoCreate && os.IsNotExist(err) {
		if err = os.Mkdir(absCgroupPath, 0755); err != nil {
			return "", fmt.Errorf("cgroup create error %v", err)
		}
		return absCgroupPath, nil
	}

	return "", fmt.Errorf("cgroup path error %v", err)
}

// FindCgroupMountPoint
// 通过/proc/self/mountinfo找出挂载了某个subsystem的hierarchy cgroup所在的目录
// example: FindCgroupMountPoint("memory")
// mountinfo: 41 25 0:33 / /sys/fs/cgroup/memory rw,relatime - cgroup cgroup rw,memory
func FindCgroupMountPoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		if len(text) == 0 {
			continue
		}

		fields := strings.Split(text, " ") // ["41", "25", "0:33", "/", "/sys/fs/cgroup/memory", "rw,relatime", "-", "cgroup", "cgroup", "rw,memory"]
		lastField := fields[len(fields)-1] // rw,memory
		for _, opt := range strings.Split(lastField, ",") {
			if opt == subsystem {
				return fields[4] // "/sys/fs/cgroup/memory"
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logrus.Errorln(err)
	}
	return ""
}
