package subsystems

type CpuSubsystem struct {
}

// 返回subsystem的名字，比如cpu memory
func (c *CpuSubsystem) Name() string {
	panic("not implemented") // TODO: Implement
}

// 设置某个cgroup在这个Subsystem中的资源限制
func (c *CpuSubsystem) Set(path string, res *ResourceConfig) error {
	panic("not implemented") // TODO: Implement
}

// 将进程添加到某个cgroup中
func (c *CpuSubsystem) Apply(path string, pid int) error {
	panic("not implemented") // TODO: Implement
}

// 移除某个cgroup
func (c *CpuSubsystem) Remove(path string) error {
	panic("not implemented") // TODO: Implement
}
