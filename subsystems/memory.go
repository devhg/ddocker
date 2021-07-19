package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// MemorySubSystem 是 memory subsystem的实现
type MemorySubSystem struct {
}

// 返回subsystem的名字，比如cpu memory
func (m *MemorySubSystem) Name() string {
	return "memory"
}

// 设置某个cgroup在这个Subsystem中的资源限制
func (m *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subSysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, true)
	if err != nil {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}

	if res.MemoryLimit == "" {
		return nil
	}

	// 设置这个cgroup的内存限制，即将限制写到cgroup对应目录的memory.limit_in_bytes文件中
	dstFile := path.Join(subSysCgroupPath, memoryLimitInBytes)
	if err := ioutil.WriteFile(dstFile, []byte(res.MemoryLimit), 0644); err != nil {
		return fmt.Errorf("set cgroup memory failed %v", err)
	}
	return nil
}

// 将进程添加到某个cgroup中
func (m *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	subSysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, true)
	if err != nil {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}

	// 把进程的PID写到cgroup的虚拟文件系统对应目录下的task文件中
	dstFile := path.Join(subSysCgroupPath, memoryTasks)
	if err := ioutil.WriteFile(dstFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("set cgroup proc failed %v", err)
	}
	return nil
}

// 移除某个cgroup
func (m *MemorySubSystem) Remove(cgroupPath string) error {
	subSysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, false)
	if err != nil {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
	return os.Remove(subSysCgroupPath)
}
