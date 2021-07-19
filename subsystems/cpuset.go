package subsystems

type CpusetSubSystem struct {
}

// 返回subsystem的名字，比如cpu memory
func (cs *CpusetSubSystem) Name() string {
	panic("not implemented") // TODO: Implement
}

// 设置某个cgroup在这个Subsystem中的资源限制
func (cs *CpusetSubSystem) Set(path string, res *ResourceConfig) error {
	panic("not implemented") // TODO: Implement
}

// 将进程添加到某个cgroup中
func (cs *CpusetSubSystem) Apply(path string, pid int) error {
	panic("not implemented") // TODO: Implement
}

// 移除某个cgroup
func (cs *CpusetSubSystem) Remove(path string) error {
	panic("not implemented") // TODO: Implement
}
