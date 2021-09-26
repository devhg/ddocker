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

func (b *BridgeNetworkDriver) Create(name string, subnet string) (*Network, error) {
	// 取到网段字符串中的网关ip地址和网络的ip段
	ip, IPRange, _ := net.ParseCIDR(subnet)
	IPRange.IP = ip

	n := &Network{
		Name:    name,
		IPRange: IPRange,
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

func createBridgeInterface(name string) error {
	// 先检查是否包含同名的Bridge设备
	_, err := net.InterfaceByName(name)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	// 初始化一个netlink的Link基础对象
	la := netlink.NewLinkAttrs()
	la.Name = name

	// 使用刚才创建的link属性创建netlink的Bridge对象
	br := &netlink.Bridge{LinkAttrs: la}

	// 创建一个Bridge虚拟设备，相当于ip link add xxxx
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("bridge %s creation error: %v", name, err)
	}

	return nil
}

// 设置网络接口的IP地址， example: setInterfaceIP("test", "192.168.0.1/24")
func setInterfaceIP(name, rawIP string) error {
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("error get interface: %v", err)
	}

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

	return netlink.AddrAdd(iface, addr)
}

func setInterfaceUP(name string) error {
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("error retrieving a link named [ %s ]: %v", iface.Attrs().Name, err)
	}

	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("error enabling interface for %s: %v", name, err)
	}
	return nil
}

func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE",
		subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	err := cmd.Run()
	output, err := cmd.Output()
	if err != nil {
		logrus.Errorf("iptables Output, %v", output)
	}
	return err
}

// 删除网络
func (b *BridgeNetworkDriver) Delete(network Network) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	return netlink.LinkDel(br)
}

// 连接容器网络端点到网络
func (b *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]
	la.MasterIndex = br.Attrs().Index

	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5],
	}

	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("error Add Endpoint Device: %v", err)
	}

	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("error Add Endpoint Device: %v", err)
	}
	return nil
}

// 从网络上移除容器的网络端点
func (b *BridgeNetworkDriver) Disconnect(network Network, endpoint *Endpoint) error {
	panic("not implemented") // TODO: Implement
}
