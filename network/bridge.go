package network

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type BridgeNetworkDriver struct {
}

func (b *BridgeNetworkDriver) Name() string {
	return "bridge"
}

// Create creates a network with network name and subnet
func (b *BridgeNetworkDriver) Create(name string, subnet string) (*Network, error) {
	// 取到网段字符串中的网关ip地址和网络的ip段
	ip, IPRange, _ := net.ParseCIDR(subnet)
	IPRange.IP = ip

	n := &Network{
		Name:    name,
		IPRange: IPRange,
		Driver:  b.Name(),
	}

	err := b.initBridge(n)
	return n, err
}

func (b *BridgeNetworkDriver) initBridge(n *Network) error {
	// 1. 创建Bridge虚拟设备
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return fmt.Errorf("add bridge: %s, error: %v", bridgeName, err)
	}

	// 2. 设置Bridge设备的地址和路由
	gatewayIP := n.IPRange
	if err := setInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		return fmt.Errorf("error assigning address: %s on bridge: %s with an error: %v",
			gatewayIP, bridgeName, err)
	}

	// 3. 启动Bridge设备
	if err := setInterfaceUP(bridgeName); err != nil {
		return fmt.Errorf("set bridge error: %v", err)
	}

	// 4. 设置iptables的SNAT规则
	if err := setupIPTables(bridgeName, n.IPRange); err != nil {
		return fmt.Errorf("setting iptables for %s error: %v", bridgeName, err)
	}

	return nil
}

func createBridgeInterface(bridgeName string) error {
	// 先检查是否包含同名的Bridge设备
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	// 初始化一个netlink的Link基础对象
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	// 使用刚才创建的link属性创建netlink的Bridge对象
	br := &netlink.Bridge{LinkAttrs: la}

	// 创建一个Bridge虚拟设备，相当于ip link add xxxx
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("bridge %s creation error: %v", bridgeName, err)
	}

	return nil
}

// 设置网络接口的IP地址， example: setInterfaceIP("test", "192.168.0.1/24")
func setInterfaceIP(bridgeName, rawIP string) error {
	iface, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("error get interface: %v", err)
	}

	// netlink.ParseIPNet("192.168.0.1/24")
	// ipNet.IP will equal 192.168.0.1, not 192.168.0.0
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}

	addr := &netlink.Addr{
		IPNet: ipNet,
		Label: "",
		Flags: 0,
		Scope: 0,
	}

	// 给网络接口配置IP地址和路由表，相当于 $(ip addr add xxx)
	// 同时如果配置了地址所在的网段信息，例如192.168.0.0/24
	// 还会配置路由表192.168.0.0/24转发到这个网络接口上
	return netlink.AddrAdd(iface, addr)
}

// setInterfaceUP 设置网络接口为UP状态
func setInterfaceUP(bridgeName string) error {
	iface, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("error retrieving a link named [%s]: %v", iface.Attrs().Name, err)
	}

	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("error enabling interface for %s: %v", bridgeName, err)
	}
	return nil
}

// setupIPTables 由于go没有直接操作iptables的库，所以直接需要通过命令来实现
// 通过iptables 创建SNAT规则，只要是从这个网桥上出来的包，都会对其做源IP的转换，
// 保证了容器经过宿主机访问到宿主机外部网络请求的包转换成机器的IP，从而正确的送达和接收。

// [root@linux ~]# iptables -D INPUT 3  //删除input的第3条规则
// [root@linux ~]# iptables -t nat -D POSTROUTING 1  //删除nat表中postrouting的第一条规则
// [root@linux ~]# iptables -F INPUT   //清空 filter表INPUT所有规则
// [root@linux ~]# iptables -F    //清空所有规则
// [root@linux ~]# iptables -t nat -F POSTROUTING   //清空nat表POSTROUTING所有规则
func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	// iptables -t nat -A POSTROUTING -s <bridgeName> ! -o <bridgeName> -j MASQUERADE
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE",
		subnet.String(), bridgeName)

	cmds := strings.Split(iptablesCmd, " ")
	cmd := exec.Command("iptables", cmds...)

	// if err := cmd.Run(); err != nil {
	// 	logrus.Errorf("[%s] run error: %v", cmd.String(), err)
	// }

	output, err := cmd.Output()
	if err != nil {
		logrus.Errorf("iptables Output, %v", output)
	}
	return err
}

// Delete 删除网络
func (b *BridgeNetworkDriver) Delete(network Network) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	return netlink.LinkDel(br)
}

// Connect 容器网络端点连接到网络
func (b *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	// 通过Linux Bridge接口名获取到接口的对象
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	// 创建veth接口配置
	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]
	// 通过veth接口的master属性，设置这个veth的一端挂载到Linux Bridge网络上
	la.MasterIndex = br.Attrs().Index

	// 创建veth对象，通过PeerName配置veth另一端的接口名
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5],
	}

	// LinkAdd方法创建出这个veth'接口，因为上面制定了link的MasterIndex是对应的Linux bridge网络
	// 所以veth的另一端已经挂载了网路对应的Linux Bridge上
	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("error Add Endpoint Device: %v", err)
	}

	// 设置veth启动 $(ip link set xxx up)
	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("error Add Endpoint Device: %v", err)
	}
	return nil
}

// 从网络上移除容器的网络端点
func (b *BridgeNetworkDriver) Disconnect(network Network, endpoint *Endpoint) error {
	panic("not implemented") // TODO: Implement
}
