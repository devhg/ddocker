package subsystems

import (
	"testing"
)

func TestFindCgroupMountpoint(t *testing.T) {
	t.Logf("cpu subsystem mount point %v\n", FindCgroupMountPoint("cpu"))
	t.Logf("cpuset subsystem mount point %v\n", FindCgroupMountPoint("cpuset"))
	t.Logf("memory subsystem mount point %v\n", FindCgroupMountPoint("memory"))
}
