package subsystems

type CpusetSubSystem struct {
}

// 返回subsystem的名字，比如cpu memory
func (cs *CpusetSubSystem) Name() string {
	return "cpuset"
}

// 设置某个cgroup在这个Subsystem中的资源限制
func (cs *CpusetSubSystem) Set(path string, res *ResourceConfig) error {
	return nil
}

// 将进程添加到某个cgroup中
func (cs *CpusetSubSystem) Apply(path string, pid int) error {
	return nil
}

// 移除某个cgroup
func (cs *CpusetSubSystem) Remove(path string) error {
	return nil
}
