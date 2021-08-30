package subsystems

// ResourceConfig 用于传递资源配置的结构体，包含内存限制，cpu时间片权重，cpu核心数
type ResourceConfig struct {
	MemoryLimit string
	CPUShare    string
	CPUSet      string
}

// SubSystemer 接口，每个Subsystem可以实现下面4个接口
// 这里将cgroup抽象成path，原因是cgroup在hierarchy的路径，便是虚拟
// 文件系统中的虚拟路径
type SubSystemer interface {
	// 返回subsystem的名字，比如cpu memory
	Name() string

	// 设置某个cgroup在这个Subsystem中的资源限制
	Set(path string, res *ResourceConfig) error

	// 将进程添加到某个cgroup中
	Apply(path string, pid int) error

	// 移除某个cgroup
	Remove(path string) error
}

// SubsystemIns subsystem instances
var SubsystemIns = []SubSystemer{
	&CPUSubsystem{},
	// &CpusetSubSystem{}, // ??
	// Q: {"level":"info","msg":"set cgroup proc failed write /sys/fs/cgroup/cpuset/ddocker-cgroup/tasks: no space left on device","time":"2021-07-25T16:33:09+08:00"}
	// A: ?
	&MemorySubSystem{},
}

const (
	memoryLimitInBytes = "memory.limit_in_bytes"
	cpuShare           = "cpu.shares"
	cpuSet             = "cpuset.cpus"
)
