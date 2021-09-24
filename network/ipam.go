package network

import "net"

// 用于网络IP地址的分配和释放
type IPAM struct {
	SubnetAllocatorPath string
}

var ipAllocator = &IPAM{
	SubnetAllocatorPath: "ipamDefaultAllocatorPath",
}

// 通过网段去分配一个可用的IP地址
func (ia *IPAM) Allocate(cidr *net.IPNet) (net.IP, error) {
	return net.IP(""), nil // TODO
}

func (ia *IPAM) Release(subnet *net.IPNet, ipaddr net.IP) error {
	return nil // TODO
}
