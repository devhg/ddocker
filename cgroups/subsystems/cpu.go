package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpuSubsystem struct {
}

// 返回subsystem的名字，比如cpu memory
func (c *CpuSubsystem) Name() string {
	return "cpu"
}

// 设置某个cgroup在这个Subsystem中的资源限制
func (c *CpuSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	subSysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, true)
	if err != nil {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}

	if res.CpuShare == "" {
		return nil
	}

	// 设置这个cgroup的内存限制，即将限制写到cgroup对应目录的cpu.shares文件中
	dstFile := path.Join(subSysCgroupPath, cpuShare)
	if err := ioutil.WriteFile(dstFile, []byte(res.CpuShare), 0644); err != nil {
		return fmt.Errorf("set cgroup memory failed %v", err)
	}
	return nil
}

// 将进程添加到某个cgroup中
func (c *CpuSubsystem) Apply(cgroupPath string, pid int) error {
	subSysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, false)
	if err != nil {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}

	// 把进程的PID写到cgroup的虚拟文件系统对应目录下的task文件中
	// "/sys/fs/cgroup/cpu/${cgroupPath}/tasks"
	dstFile := path.Join(subSysCgroupPath, "tasks")
	if err := ioutil.WriteFile(dstFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("set cgroup proc failed %v", err)
	}
	return nil
}

// 移除某个cgroup
func (c *CpuSubsystem) Remove(cgroupPath string) error {
	subSysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, false)
	if err == nil {
		return os.RemoveAll(subSysCgroupPath)
	}
	return err
}
