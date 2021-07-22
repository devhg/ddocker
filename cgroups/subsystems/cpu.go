package subsystems

type CpuSubsystem struct {
}

// 返回subsystem的名字，比如cpu memory
func (c *CpuSubsystem) Name() string {
	return "cpu"
}

// 设置某个cgroup在这个Subsystem中的资源限制
func (c *CpuSubsystem) Set(path string, res *ResourceConfig) error {
	return nil
}

// 将进程添加到某个cgroup中
func (c *CpuSubsystem) Apply(path string, pid int) error {
	return nil
}

// 移除某个cgroup
func (c *CpuSubsystem) Remove(path string) error {
	return nil
}
